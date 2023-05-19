package main

import (
	"encoding/json"
	"github.com/Real-Dev-Squad/feature-flag-backend/database"
	"github.com/Real-Dev-Squad/feature-flag-backend/models"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
	"strconv"
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
func updateFeatureFlag(updateFeatureFlagRequest models.UpdateFeatureFlagRequest) (events.APIGatewayProxyResponse, error) {
	db := database.CreateDynamoDB()

	input := &dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(updateFeatureFlagRequest.Id),
			},
		},
		TableName: aws.String(database.GetFeatureFlagTableName()),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":status": {
				S: aws.String(updateFeatureFlagRequest.Status),
			},
			":updatedAt": {
				N: aws.String(strconv.Itoa(int(time.Now().Unix()))),
			},
			":updatedBy": {
				S: aws.String(updateFeatureFlagRequest.UserId),
			},
		},
		UpdateExpression: aws.String("set #status = :status, #updatedAt = :updatedAt, #updatedBy = :updatedBy"),
		ExpressionAttributeNames: map[string]*string{
			"#status":    aws.String("status"),
			"#updatedAt": aws.String("updatedAt"),
			"#updatedBy": aws.String("updatedBy"),
		},
		ReturnValues:        aws.String("ALL_NEW"),
		ConditionExpression: aws.String("attribute_exists(Id)"),
	}

	result, err := db.UpdateItem(input)

	//throw the response on conditional check failed exception
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
				return utils.ClientError(http.StatusNotFound, "Feature flag with given flagId doesn't exists")
			}
		}
		utils.ServerError(err)
	}

	featureFlag := new(models.FeatureFlag)
	err = dynamodbattribute.UnmarshalMap(result.Attributes, &featureFlag)

	if err != nil {
		log.Println()
		utils.ServerError(err)
	}

	//marshal to JSON
	resultJson, err := json.Marshal(featureFlag)
	if err != nil {
		log.Println("Unable to marshal to JSON", featureFlag.Id)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(resultJson),
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
				Body:       "Check the request body passed status, flagId and userId are required.",
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

	return updateFeatureFlag(updateFeatureFlagRequest)
}

func main() {
	lambda.Start(handler)
}
