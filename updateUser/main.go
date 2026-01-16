package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"feature-flag-backend/layer/database"
	"feature-flag-backend/layer/jwt"
	middleware "feature-flag-backend/layer/middlewares"
	"feature-flag-backend/layer/models"
	"feature-flag-backend/layer/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

type UpdateUserRequest struct {
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	Role     string `json:"role,omitempty" validate:"omitempty,oneof=ADMIN DEVELOPER VIEWER"`
	IsActive *bool  `json:"isActive,omitempty"`
}

func getUserById(ctx context.Context, db *dynamodb.Client, userId string) (*models.User, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(utils.USER_TABLE_NAME),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{
				Value: userId,
			},
		},
	}

	result, err := db.GetItem(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(result.Item) == 0 {
		return nil, nil
	}

	var user models.User
	err = database.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func updateUser(ctx context.Context, db *dynamodb.Client, userId string, updateRequest UpdateUserRequest, updatedBy string) (*models.User, error) {
	// Get existing user
	existingUser, err := getUserById(ctx, db, userId)
	if err != nil {
		return nil, err
	}
	if existingUser == nil {
		return nil, nil
	}

	// Update fields if provided
	if updateRequest.Email != "" {
		existingUser.Email = updateRequest.Email
	}
	if updateRequest.Role != "" {
		existingUser.Role = updateRequest.Role
	}
	if updateRequest.IsActive != nil {
		existingUser.IsActive = *updateRequest.IsActive
	}

	existingUser.UpdatedBy = updatedBy
	existingUser.UpdatedAt = time.Now().Unix()

	item, err := database.MarshalMap(existingUser)
	if err != nil {
		log.Printf("Error marshalling user to DynamoDB AttributeValue: %v", err)
		return nil, err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(utils.USER_TABLE_NAME),
		Item:      item,
	}

	_, err = db.PutItem(ctx, input)
	if err != nil {
		log.Printf("Error updating user in Dynamodb: %v", err)
		return nil, err
	}

	return existingUser, nil
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var updateRequest UpdateUserRequest

	db := database.CreateDynamoDB()

	utils.CheckRequestAllowed(ctx, db, utils.ConcurrencyDisablingLambda)

	corsResponse, err, passed := middleware.HandleCORS(req)
	if !passed {
		return corsResponse, err
	}

	// Use enhanced middleware with user verification and RBAC (Week 3)
	jwtResponse, userContext, err := jwt.JWTMiddlewareWithUserVerification()(req)
	if err != nil || jwtResponse.StatusCode != http.StatusOK {
		return jwtResponse, err
	}

	if userContext == nil {
		return utils.ClientError(http.StatusUnauthorized, "User context not available")
	}

	corsHeaders := middleware.GetCORSHeadersV1(req.Headers)

	userId := req.PathParameters["userId"]
	if userId == "" {
		log.Println("userId is required")
		clientErrorResponse, _ := utils.ClientError(http.StatusBadRequest, "userId is required")
		return clientErrorResponse, nil
	}

	// Check permission: UPDATE_USER (Week 3 RBAC)
	// Only ADMIN can update other users, users can update themselves
	if userId != userContext.UserId && userContext.Role != utils.ROLE_ADMIN {
		// Check if trying to update role (only ADMIN can do this)
		var tempRequest UpdateUserRequest
		json.Unmarshal([]byte(req.Body), &tempRequest)
		if tempRequest.Role != "" {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusForbidden,
				Body:       "Only ADMIN can update user roles",
				Headers:    corsHeaders,
			}, nil
		}
	}

	// Check if user can access this resource (own resources or ADMIN)
	if !utils.CanAccessUserResource(userContext, userId) {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusForbidden,
			Body:       "You can only update your own profile",
			Headers:    corsHeaders,
		}, nil
	}

	// For non-ADMIN users updating themselves, restrict role updates
	if userId == userContext.UserId && userContext.Role != utils.ROLE_ADMIN {
		var tempRequest UpdateUserRequest
		json.Unmarshal([]byte(req.Body), &tempRequest)
		if tempRequest.Role != "" && tempRequest.Role != userContext.Role {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusForbidden,
				Body:       "You cannot change your own role",
				Headers:    corsHeaders,
			}, nil
		}
	}

	err = json.Unmarshal([]byte(req.Body), &updateRequest)
	if err != nil {
		log.Printf("Error unmarshal request body: %v", err)
		return utils.ClientError(http.StatusUnprocessableEntity, "Error unmarshalling request body")
	}

	if err := validate.Struct(&updateRequest); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Invalid request: email must be valid, role must be one of ADMIN, DEVELOPER, or VIEWER",
			StatusCode: http.StatusBadRequest,
			Headers:    corsHeaders,
		}, nil
	}

	user, err := updateUser(ctx, db, userId, updateRequest, userContext.UserId)
	if err != nil {
		log.Printf("Error while updating user: %v", err)
		return utils.ServerError(err)
	}

	if user == nil {
		return events.APIGatewayProxyResponse{
			Body:       "User not found",
			StatusCode: http.StatusNotFound,
			Headers:    corsHeaders,
		}, nil
	}

	userResponse := utils.UserResponse{
		Id:        user.Id,
		Email:     user.Email,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	responseBody, err := json.Marshal(map[string]interface{}{
		"message": "User updated successfully",
		"data":    userResponse,
	})

	if err != nil {
		log.Printf("Error marshalling response body: %v", err)
		return utils.ServerError(err)
	}

	response := events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    corsHeaders,
		Body:       string(responseBody),
	}

	return response, nil
}

func main() {
	lambda.Start(handler)
}

