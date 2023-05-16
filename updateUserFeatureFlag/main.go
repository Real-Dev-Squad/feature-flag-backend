package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Real-Dev-Squad/feature-flag-backend/database"
	"github.com/Real-Dev-Squad/feature-flag-backend/models"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func processUpdateById(userId string, flagId string, requestBody models.UpdateUserMapping) (*models.FeatureFlagUserMapping, error) {

	db := database.CreateDynamoDB()

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(database.GetTableName(utils.FFUM_TABLE_NAME)),
		Key: map[string]*dynamodb.AttributeValue{
			"userId": {
				S: aws.String(userId),
			},
			"flagId": {
				S: aws.String(flagId),
			},
		},
		UpdateExpression: aws.String("set #status = :status, #updatedBy = :updatedBy, #updatedAt = :updatedAt"),
		ExpressionAttributeNames: map[string]*string{
			"#status":    aws.String("status"),
			"#updatedBy": aws.String("updatedBy"),
			"#updatedAt": aws.String("updatedAt"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":status": {
				S: aws.String(strings.ToUpper(requestBody.Status)),
			},
			":updatedBy": {
				S: aws.String(requestBody.UserId),
			},
			":updatedAt": {
				N: aws.String(strconv.FormatInt(time.Now().Unix(), 10)),
			},
		},
		ConditionExpression: aws.String("attribute_exists(userId)"),
		ReturnValues:        aws.String("ALL_NEW"),
	}

	// Perform the update operation
	result, err := db.UpdateItem(input)

	if err != nil {
		utils.DdbError(err)
		return nil, err
	}

	featureFlagUserMapping := new(models.FeatureFlagUserMapping)
	err = dynamodbattribute.UnmarshalMap(result.Attributes, &featureFlagUserMapping)

	if err != nil {
		return nil, err
	}
	return featureFlagUserMapping, nil
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userId := req.PathParameters["userId"]
	flagId := req.PathParameters["flagId"]

	var requestBody models.UpdateUserMapping
	err := json.Unmarshal([]byte(req.Body), &requestBody)
	if err != nil {
		log.Println("Error in reading input")
		return utils.ClientError(http.StatusBadRequest, "Error in reading input")
	}

	//validate the request
	if err := validate.Struct(&requestBody); err != nil {
		errs := utils.HandleValidationError(err)

		//TODO: use the errs response and pass it to the user instead of hardcoded message.
		if len(errs) > 0 {
			return events.APIGatewayProxyResponse{
				Body:       "Check the request body passed status and userId are required.",
				StatusCode: http.StatusBadRequest,
			}, nil
		}

	}

	//check if the status is valid one.
	found := utils.ValidateFeatureFlagStatus(requestBody.Status)
	// if the status is not valid one.
	if !found {
		return events.APIGatewayProxyResponse{
			Body:       "Allowed values of Status are ENABLED, DISABLED",
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	result, err := processUpdateById(userId, flagId, requestBody)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
				return utils.ClientError(http.StatusNotFound, "Mapping of User Id and Flag Id does not exist")
			}
		}
		return utils.ServerError(err)
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		log.Println("Error converting featureFlagUserMapping to JSON")
		return utils.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		Body:       string(resultJson),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
