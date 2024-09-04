package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Real-Dev-Squad/feature-flag-backend/database"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func processGetById(userId string) ([]utils.FeatureFlagUserMappingResponse, error) {

	db := database.CreateDynamoDB()

	utils.CheckRequestAllowed(db, utils.ConcurrencyDisablingLambda)

	input := &dynamodb.QueryInput{
		TableName:              aws.String(utils.FEATURE_FLAG_USER_MAPPING_TABLE_NAME),
		KeyConditionExpression: aws.String(fmt.Sprintf("%v = :uid", utils.UserId)),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":uid": {
				S: aws.String(userId),
			},
		},
	}

	result, err := db.Query(input)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, nil
	}

	var listOfFeatureFlagUserMapping []utils.FeatureFlagUserMappingResponse

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &listOfFeatureFlagUserMapping)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return listOfFeatureFlagUserMapping, nil
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userId := req.PathParameters["userId"]
	result, err := processGetById(userId)
	if err != nil {
		return utils.ServerError(err)
	}
	if result == nil {
		log.Println("User feature flags not found")
		return utils.ClientError(http.StatusNotFound, "User feature flags not found")
	}
	resultJson, err := json.Marshal(result)
	if err != nil {
		log.Println("Error converting listOfFeatureFlagUserMapping to JSON")
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
