package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"feature-flag-backend/layer/database"
	middleware "feature-flag-backend/layer/middlewares"
	"feature-flag-backend/layer/models"
	"feature-flag-backend/layer/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
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

func createUser(ctx context.Context, db *dynamodb.Client, registerRequest utils.RegisterUserRequest) (models.User, error) {
	existingUser, err := getUserByEmail(ctx, db, registerRequest.Email)
	if err != nil {
		log.Printf("Error checking existing user: %v", err)
		return models.User{}, err
	}
	if existingUser != nil {
		return models.User{}, &utils.UserExistsError{Email: registerRequest.Email}
	}

	passwordHash, err := utils.HashPassword(registerRequest.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return models.User{}, err
	}

	role := registerRequest.Role
	if role == "" {
		role = utils.ROLE_VIEWER
	}

	now := time.Now().Unix()
	userId := uuid.New().String()

	user := models.User{
		Id:           userId,
		Email:        registerRequest.Email,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    now,
		CreatedBy:    userId,
		UpdatedAt:    now,
		UpdatedBy:    userId,
		IsActive:     true,
	}

	item, err := database.MarshalMap(user)
	if err != nil {
		log.Printf("Error marshalling user to DynamoDB AttributeValue: %v", err)
		return models.User{}, err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(utils.USER_TABLE_NAME),
		Item:      item,
	}

	_, err = db.PutItem(ctx, input)
	if err != nil {
		log.Printf("Error putting user to Dynamodb: %v", err)
		return models.User{}, err
	}

	return user, nil
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var registerRequest utils.RegisterUserRequest

	db := database.CreateDynamoDB()

	utils.CheckRequestAllowed(ctx, db, utils.ConcurrencyDisablingLambda)

	corsResponse, err, passed := middleware.HandleCORS(req)
	if !passed {
		return corsResponse, err
	}

	// Registration endpoint doesn't require authentication
	corsHeaders := middleware.GetCORSHeadersV1(req.Headers)

	err = json.Unmarshal([]byte(req.Body), &registerRequest)
	if err != nil {
		log.Printf("Error unmarshal request body: %v", err)
		return utils.ClientError(http.StatusUnprocessableEntity, "Error unmarshalling request body")
	}

	if err := validate.Struct(&registerRequest); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Invalid request: email and password are required. Password must be at least 8 characters.",
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	user, err := createUser(ctx, db, registerRequest)
	if err != nil {
		if _, ok := err.(*utils.UserExistsError); ok {
			return events.APIGatewayProxyResponse{
				Body:       "User with this email already exists",
				StatusCode: http.StatusConflict,
				Headers:    corsHeaders,
			}, nil
		}
		log.Printf("Error while creating user: %v", err)
		return utils.ServerError(err)
	}

	response := events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
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

	responseBody, err := json.Marshal(map[string]interface{}{
		"message": "User registered successfully",
		"data":    userResponse,
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

