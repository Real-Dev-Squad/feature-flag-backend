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

func processUpdateByIds(userId string, flagId string, requestBody models.UpdateUserMapping) (*models.FeatureFlagUserMapping, error) {

	db := database.CreateDynamoDB()

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(database.GetTableName(utils.FFUM_TABLE_NAME)),
		Key: map[string]*dynamodb.AttributeValue{
			utils.UserId: {
				S: aws.String(userId),
			},
			utils.FlagId: {
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
		log.Printf("Error unmarshal request body: \n %v", err)
		return utils.ClientError(http.StatusUnprocessableEntity, "Error unmarshalling request body")
	}

	//validate the request
	if err := validate.Struct(&requestBody); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Check the request body passed status and userId are required.",
			StatusCode: http.StatusBadRequest,
		}, nil
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

	result, err := processUpdateByIds(userId, flagId, requestBody)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
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
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	lambda.Start(handler)
}
