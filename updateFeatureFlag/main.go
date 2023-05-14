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
	"log"
	"net/http"
	"time"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func handleValidationError(err error) []utils.ValidationError {
	var errors []utils.ValidationError

	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, utils.ValidationError{
			Field: err.Field(),
			Error: err.Tag(),
		})
		log.Println("Errors are ", errors)
	}
	return errors
}

// creating instance of databae
func updateFeatureFlag(featureFlag *models.FeatureFlag) (events.APIGatewayProxyResponse, error) {
	db := database.CreateDynamoDB()

	item, err := dynamodbattribute.MarshalMap(featureFlag)
	if err != nil {
		return utils.ServerError(err)
	}

	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(database.GetFeatureFlagTableName()),
	}

	_, err = db.PutItem(input)
	if err != nil {
		return utils.ServerError(err)
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       "Feature flag updated sucessfully",
	}, nil
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	updateFeatureFlagRequest := models.UpdateFeatureFlagRequest{}

	//marshal to updateFeatureFlag
	bytes := []byte(request.Body)
	err := json.Unmarshal(bytes, &updateFeatureFlagRequest)
	if err != nil {
		log.Println("Error in reading input.")
		return utils.ClientError(http.StatusBadRequest, "Error in reading input")
	}

	//validate the request
	if err := validate.Struct(&updateFeatureFlagRequest); err != nil {
		errs := handleValidationError(err)

		//TODO: use the errs response and pass it to the user instead of hardcoded message.
		if len(errs) > 0 {
			return events.APIGatewayProxyResponse{
				Body:       "Check the request body passed name, description, flagId and userId are required.",
				StatusCode: http.StatusBadRequest,
			}, nil
		}

	}

	//check if the status is valid one.
	allowedStatuses := []string{"ENABLED", "DISABLED"}
	found := false
	for _, allowedStatus := range allowedStatuses {
		if updateFeatureFlagRequest.Status == allowedStatus {
			found = true
			break
		}
	}
	// if the status is not valid one.
	if !found {
		return events.APIGatewayProxyResponse{
			Body:       "Allowed values of Status are ENABLED, DISABLED",
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	//updating feature flag
	featureFlag, err := database.ProcessGetFeatureFlagByHashKey(utils.Id, updateFeatureFlagRequest.Id)
	if err != nil {
		return utils.ServerError(err)
	}

	if featureFlag == nil {
		return utils.ClientError(http.StatusNotFound, "Feature flag not found")
	}

	if updateFeatureFlagRequest.Description != "" {
		featureFlag.Description = updateFeatureFlagRequest.Description
	}

	if updateFeatureFlagRequest.Status != "" {
		featureFlag.Status = updateFeatureFlagRequest.Status
	}

	//updating the lastUpdatedBy and lastUpdatedAt
	featureFlag.UpdatedAt = time.Now().Unix()
	featureFlag.UpdatedBy = updateFeatureFlagRequest.UserId

	//update feature flag
	response, err := updateFeatureFlag(featureFlag)
	if err != nil {
		return utils.ServerError(err)
	}
	return response, nil
}

func main() {
	lambda.Start(handler)
}
