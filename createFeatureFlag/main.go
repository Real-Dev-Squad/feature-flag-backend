package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Real-Dev-Squad/feature-flag-backend/database"
	"github.com/Real-Dev-Squad/feature-flag-backend/jwt"
	middleware "github.com/Real-Dev-Squad/feature-flag-backend/middlewares"
	"github.com/Real-Dev-Squad/feature-flag-backend/models"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func createFeatureFlag(db *dynamodb.DynamoDB, createFeatureFlagRequest utils.CreateFeatureFlagRequest) (models.FeatureFlag, error) {
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

	_, err = db.PutItem(input)
	if err != nil {
		log.Printf("Error putting item to Dynamodb: \n %v", err)
		return models.FeatureFlag{}, err
	}
	return featureFlag, nil
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var createFeatureFlagRequest utils.CreateFeatureFlagRequest

	db := database.CreateDynamoDB()

	utils.CheckRequestAllowed(db, utils.ConcurrencyDisablingLambda)

	corsResponse, err := middleware.CORSMiddleware()(req)
	if err != nil {
		log.Printf("CORS error: %v", err)
		return corsResponse, err
	}

	if corsResponse.StatusCode != http.StatusOK {
		return corsResponse, nil
	}

	jwtResponse, _, err := jwt.JWTMiddleware()(req)
	if err != nil {
		log.Printf("JWT middleware error: %v", err)
		return jwtResponse, err
	}

	if jwtResponse.StatusCode != http.StatusOK {
		return jwtResponse, nil
	}

	err = json.Unmarshal([]byte(req.Body), &createFeatureFlagRequest)
	if err != nil {
		log.Printf("Error unmarshal request body: \n %v", err)
		return utils.ClientError(http.StatusUnprocessableEntity, "Error unmarshalling request body")
	}

	if err := validate.Struct(&createFeatureFlagRequest); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Check the request body passed name, description and userId are required.",
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	featureFlag, err := createFeatureFlag(db, createFeatureFlagRequest)
	if err != nil {
		log.Printf("Error while creating feature flag: \n %v ", err)
		return utils.ServerError(err)
	}

	origin := req.Headers["Origin"]
	corsHeaders := middleware.GetCORSHeaders(origin)

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
