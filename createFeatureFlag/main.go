package main

import (
	"encoding/json"
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
	"log"
	"net/http"
	"time"
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

	//marshal to ddb attribute value
	item, err := dynamodbattribute.MarshalMap(featureFlag)
	if err != nil {
		log.Println("Error marshalling object to DynamoDB AttributeValue: %v", err)
		return err
	}

	log.Println("item is", item)
	input := &dynamodb.PutItemInput{
		TableName: aws.String(database.GetFeatureFlagTableName()),
		Item:      item,
	}

	log.Println(db, " is the db object")
	_, err = db.PutItem(input)
	if err != nil {
		log.Println("Error putting item to Dynamodb: %v", err)
		return err
	}
	return nil
}

// //TODO: use the validation Error struct response and send it to the user

func handleValidationError(err error) []utils.ValidationError {
	var errors []utils.ValidationError

	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, utils.ValidationError{
			Field: err.Field(),
			Error: err.Tag(),
		})
	}
	return errors

}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var createFeatureFlagRequest utils.CreateFeatureFlagRequest

	db := database.CreateDynamoDB()

	err := json.Unmarshal([]byte(req.Body), &createFeatureFlagRequest)
	if err != nil {
		log.Println("Error unmarshal request body: %v", err)
		return utils.ClientError(http.StatusUnprocessableEntity, "Error unmarshalling request body")
	}

	if err := validate.Struct(&createFeatureFlagRequest); err != nil {
		errs := handleValidationError(err)

		//TODO: use the errs response and pass it to the user instead of hardcoded message.
		if len(errs) > 0 {
			return events.APIGatewayProxyResponse{
				Body:       "Check the request body passed name, description and userId are required.",
				StatusCode: http.StatusBadRequest,
			}, nil
		}
	}
	//creating a feature flag
	err = createFeatureFlag(db, createFeatureFlagRequest)
	if err != nil {
		log.Printf("error is %v ", err)
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
