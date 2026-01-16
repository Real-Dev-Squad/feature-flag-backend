package main

import (
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
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func createFeatureFlag(ctx context.Context, db *dynamodb.Client, createFeatureFlagRequest utils.CreateFeatureFlagRequest) (models.FeatureFlag, error) {
	featureFlag := models.FeatureFlag{
		Id:          uuid.New().String(),
		Name:        createFeatureFlagRequest.FlagName,
		Description: createFeatureFlagRequest.Description,
		CreatedBy:   createFeatureFlagRequest.UserId,
		UpdatedBy:   createFeatureFlagRequest.UserId,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		Status:      utils.ENABLED,
	}

	item, err := database.MarshalMap(featureFlag)
	if err != nil {
		log.Printf("Error marshalling object to DynamoDB AttributeValue: \n %v", err)
		return models.FeatureFlag{}, err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(utils.FEATURE_FLAG_TABLE_NAME),
		Item:      item,
	}

	_, err = db.PutItem(ctx, input)
	if err != nil {
		log.Printf("Error putting item to Dynamodb: \n %v", err)
		return models.FeatureFlag{}, err
	}
	return featureFlag, nil
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var createFeatureFlagRequest utils.CreateFeatureFlagRequest

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

	// Check permission: CREATE_FEATURE_FLAG (Week 3 RBAC)
	permResponse, err := utils.RequirePermission(userContext, utils.PermissionCreateFeatureFlag)
	if err != nil || permResponse.StatusCode != http.StatusOK {
		permResponse.Headers = middleware.GetCORSHeadersV1(req.Headers)
		return permResponse, err
	}

	corsHeaders := middleware.GetCORSHeadersV1(req.Headers)

	err = json.Unmarshal([]byte(req.Body), &createFeatureFlagRequest)
	if err != nil {
		log.Printf("Error unmarshal request body: \n %v", err)
		return utils.ClientError(http.StatusUnprocessableEntity, "Error unmarshalling request body")
	}

	if err := validate.Struct(&createFeatureFlagRequest); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Check the request body passed name and description are required.",
			StatusCode: http.StatusBadRequest,
			Headers:    corsHeaders,
		}, nil
	}

	// Use userId from authenticated user context (Week 2 migration)
	// Override any userId in request body with authenticated user
	createFeatureFlagRequest.UserId = userContext.UserId

	featureFlag, err := createFeatureFlag(ctx, db, createFeatureFlagRequest)
	if err != nil {
		log.Printf("Error while creating feature flag: \n %v ", err)
		return utils.ServerError(err)
	}

	response := events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Headers:    corsHeaders,
	}

	responseBody, err := json.Marshal(map[string]interface{}{
		"message": "Created feature flag successfully",
		"data":    featureFlag,
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
