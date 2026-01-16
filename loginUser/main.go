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
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func getUserByEmail(ctx context.Context, db *dynamodb.Client, email string) (*models.User, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(utils.USER_TABLE_NAME),
		IndexName:              aws.String("email-index"),
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{
				Value: email,
			},
		},
		Limit: aws.Int32(1),
	}

	result, err := db.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var user models.User
	err = database.UnmarshalMap(result.Items[0], &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var loginRequest utils.LoginUserRequest

	db := database.CreateDynamoDB()

	utils.CheckRequestAllowed(ctx, db, utils.ConcurrencyDisablingLambda)

	corsResponse, err, passed := middleware.HandleCORS(req)
	if !passed {
		return corsResponse, err
	}

	corsHeaders := middleware.GetCORSHeadersV1(req.Headers)

	err = json.Unmarshal([]byte(req.Body), &loginRequest)
	if err != nil {
		log.Printf("Error unmarshal request body: %v", err)
		return utils.ClientError(http.StatusUnprocessableEntity, "Error unmarshalling request body")
	}

	if err := validate.Struct(&loginRequest); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Invalid request: email and password are required.",
			StatusCode: http.StatusBadRequest,
			Headers:    corsHeaders,
		}, nil
	}

	user, err := getUserByEmail(ctx, db, loginRequest.Email)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return utils.ServerError(err)
	}

	if user == nil {
		return events.APIGatewayProxyResponse{
			Body:       "Invalid email or password",
			StatusCode: http.StatusUnauthorized,
			Headers:    corsHeaders,
		}, nil
	}

	// Check if user is active
	if !user.IsActive {
		return events.APIGatewayProxyResponse{
			Body:       "User account is inactive",
			StatusCode: http.StatusForbidden,
			Headers:    corsHeaders,
		}, nil
	}

	// Verify password
	if !utils.CheckPasswordHash(loginRequest.Password, user.PasswordHash) {
		return events.APIGatewayProxyResponse{
			Body:       "Invalid email or password",
			StatusCode: http.StatusUnauthorized,
			Headers:    corsHeaders,
		}, nil
	}

	// Generate JWT token
	jwtUtils, err := jwt.GetInstance()
	if err != nil {
		log.Printf("Error getting JWT utils: %v", err)
		return utils.ServerError(err)
	}

	token, err := jwtUtils.GenerateToken(user.Id, user.Role)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		return events.APIGatewayProxyResponse{
			Body:       "Error generating authentication token",
			StatusCode: http.StatusInternalServerError,
			Headers:    corsHeaders,
		}, nil
	}

	response := events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    corsHeaders,
	}

	userResponse := utils.UserResponse{
		Id:        user.Id,
		Email:     user.Email,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	loginResponse := utils.LoginResponse{
		Token: token,
		User:  userResponse,
	}

	responseBody, err := json.Marshal(map[string]interface{}{
		"message": "Login successful",
		"data":    loginResponse,
	})

	if err != nil {
		log.Printf("Error marshalling response body: %v", err)
		return utils.ServerError(err)
	}

	response.Body = string(responseBody)

	return response, nil
}

func main() {
	lambda.Start(handler)
}

