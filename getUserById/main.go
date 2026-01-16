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

	response, _, err := jwt.JWTMiddleware()(req)
	if err != nil || response.StatusCode != http.StatusOK {
		return response, err
	}

	corsHeaders := middleware.GetCORSHeadersV1(req.Headers)

	userId := req.PathParameters["userId"]
	if userId == "" {
		log.Println("userId is required")
		clientErrorResponse, _ := utils.ClientError(http.StatusBadRequest, "userId is required")
		return clientErrorResponse, nil
	}

	// TODO: Add role-based access control in Week 3
	// Users can only view their own profile unless they're ADMIN

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

	response = events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    corsHeaders,
		Body:       string(jsonResponse),
	}
	return response, nil
}

func main() {
	lambda.Start(handler)
}

