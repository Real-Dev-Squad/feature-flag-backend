package main

import (
	"encoding/json"
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

func getAllFeatureFlags(db *dynamodb.DynamoDB) ([]utils.FeatureFlagResponse, error) {

	input := &dynamodb.ScanInput{
		TableName: aws.String(utils.FEATURE_FLAG_TABLE_NAME),
	}
	result, err := db.Scan(input)

	if err != nil {
		utils.DdbError(err)
		return nil, err
	}

	if len(result.Items) == 0 {
		return []utils.FeatureFlagResponse{}, nil
	}

	var featureFlagsResponse []utils.FeatureFlagResponse

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &featureFlagsResponse)
	if err != nil {
		log.Println("Something went wrong in unmarshalling all feature flags response", err)
		return nil, err
	}

	return featureFlagsResponse, nil
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	db := database.CreateDynamoDB()

	featureFlagsResponse, err := getAllFeatureFlags(db)

	if err != nil {
		return utils.ServerError(err)
	}

	if len(featureFlagsResponse) == 0 {
		return utils.ClientError(http.StatusNotFound, "No feature flags found.")
	}

	jsonResult, err := json.Marshal(featureFlagsResponse)
	if err != nil {
		log.Println("Error converting feature flags to JSON")
		return utils.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		Body:       string(jsonResult),
		StatusCode: http.StatusOK,
	}, nil

}

func main() {
	lambda.Start(handler)
}
