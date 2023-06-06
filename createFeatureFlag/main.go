package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Real-Dev-Squad/feature-flag-backend/database"
	"github.com/Real-Dev-Squad/feature-flag-backend/models"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func createFeatureFlag(db *dynamodb.DynamoDB, createFeatureFlagRequest utils.CreateFeatureFlagRequest) error {
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

	item, err := dynamodbattribute.MarshalMap(featureFlag)
	if err != nil {
		log.Printf("Error marshalling object to DynamoDB AttributeValue: \n %v", err)
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(database.GetTableName(utils.FF_TABLE_NAME)),
		Item:      item,
	}

	_, err = db.PutItem(input)
	if err != nil {
		log.Printf("Error putting item to Dynamodb: \n %v", err)
		return err
	}
	return nil
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var createFeatureFlagRequest utils.CreateFeatureFlagRequest

	db := database.CreateDynamoDB()

	err := json.Unmarshal([]byte(req.Body), &createFeatureFlagRequest)
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

	err = createFeatureFlag(db, createFeatureFlagRequest)
	if err != nil {
		log.Printf("Error while creating feature flag: \n %v ", err)
		return utils.ServerError(err)
	}

	response := events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: "Created feature flag sucessfully",
	}
	return response, nil
}

func main() {
	lambda.Start(handler)
}
