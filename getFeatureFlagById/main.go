package main

import (
	"fmt"
	"github.com/Real-Dev-Squad/feature-flag-backend/database"
	"github.com/Real-Dev-Squad/feature-flag-backend/models"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
)

func processGetById(id string) (*models.FeatureFlag, error) {

	db := database.CreateDynamoDB()

	input := &dynamodb.GetItemInput{
		TableName: aws.String(database.GetFeatureFlagTableName()),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": { //TODO remove hardcoded value
				S: aws.String(id),
			},
		},
	}

	result, err := db.GetItem(input)
	if err != nil {
		return nil, err
	}

	featureFlag := new(models.FeatureFlag)
	err = dynamodbattribute.UnmarshalMap(result.Item, &featureFlag)

	if err != nil {
		return nil, err
	}
	return featureFlag, nil
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id, found := req.PathParameters["flagId"]
	if !found {
		log.Panic("Id not passed")
	}
	item, err := processGetById(id)
	if err != nil {
		utils.ServerError(err)
	}
	fmt.Println(item)
	return events.APIGatewayProxyResponse{}, nil

}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	lambda.Start(handler)
}
