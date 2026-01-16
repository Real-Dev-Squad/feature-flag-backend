package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

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
)

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

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	// Check permission: READ_USER (Week 3 RBAC)
	permResponse, err := utils.RequirePermission(userContext, utils.PermissionReadUser)
	if err != nil || permResponse.StatusCode != http.StatusOK {
		permResponse.Headers = corsHeaders
		return permResponse, err
	}

	// Check if user can access this resource (own resources or ADMIN)
	if !utils.CanAccessUserResource(userContext, userId) {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusForbidden,
			Body:       "You can only view your own profile",
			Headers:    corsHeaders,
		}, nil
	}

	user, err := getUserById(ctx, db, userId)
	if err != nil {
		log.Printf("Database error: %v", err)
		serverErrorResponse, _ := utils.ServerError(err)
		return serverErrorResponse, nil
	}

	if user == nil {
		log.Println("User not found")
		clientErrorResponse, _ := utils.ClientError(http.StatusNotFound, "User not found")
		clientErrorResponse.Headers = corsHeaders
		return clientErrorResponse, nil
	}

	// Don't return password hash
	userResponse := utils.UserResponse{
		Id:        user.Id,
		Email:     user.Email,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	jsonResponse, err := json.Marshal(userResponse)
	if err != nil {
		log.Printf("Error converting User to JSON: %v", err)
		serverErrorResponse, _ := utils.ServerError(err)
		return serverErrorResponse, nil
	}

	response := events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    corsHeaders,
		Body:       string(jsonResponse),
	}
	return response, nil
}

func main() {
	lambda.Start(handler)
}

