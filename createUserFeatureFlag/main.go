package main

import (
	"encoding/json"
	"log"
	"net/http"
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

func processPutById(userId string, flagId string, featureFlagUserMapping models.FeatureFlagUserMapping) (*models.FeatureFlagUserMapping, error) {

	db := database.CreateDynamoDB()

	item, err := dynamodbattribute.MarshalMap(featureFlagUserMapping)
	if err != nil {
		return nil, err
	}

	input := &dynamodb.PutItemInput{
		TableName:           aws.String(database.GetTableName(utils.FFUM_TABLE_NAME)),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(userId)"),
	}

	_, err = db.PutItem(input)
	if err != nil {
		utils.DdbError(err)
		return nil, err
	}

	return &featureFlagUserMapping, nil
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userId := req.PathParameters["userId"]
	flagId := req.PathParameters["flagId"]

	var requestBody models.CreateUserMapping
	err := json.Unmarshal([]byte(req.Body), &requestBody)
	if err != nil {
		log.Printf("Error unmarshal request body: \n %v", err)
		return utils.ClientError(http.StatusUnprocessableEntity, "Error unmarshal request body")
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

	featureFlagUserMapping := models.FeatureFlagUserMapping{
		UserId:    userId,
		FlagId:    flagId,
		Status:    strings.ToUpper(requestBody.Status),
		CreatedAt: time.Now().Unix(),
		CreatedBy: requestBody.UserId,
	}

	result, err := processPutById(userId, flagId, featureFlagUserMapping)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
				return utils.ClientError(http.StatusNotFound, "Mapping of User Id and Flag Id already exists")
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
