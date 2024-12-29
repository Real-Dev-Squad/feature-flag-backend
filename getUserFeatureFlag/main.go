package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Real-Dev-Squad/feature-flag-backend/database"
	"github.com/Real-Dev-Squad/feature-flag-backend/jwt"
	middleware "github.com/Real-Dev-Squad/feature-flag-backend/middlewares"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-lambda-go/events"
	lambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func processGetById(userId string, flagId string) (*utils.FeatureFlagUserMappingResponse, error) {

	db := database.CreateDynamoDB()

	utils.CheckRequestAllowed(db, utils.ConcurrencyDisablingLambda)

	input := &dynamodb.GetItemInput{
		TableName: aws.String(utils.FEATURE_FLAG_USER_MAPPING_TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			utils.UserId: { // partition key
				S: aws.String(userId),
			},
			utils.FlagId: { // sort key
				S: aws.String(flagId),
			},
		},
	}

	result, err := db.GetItem(input)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}

	featureFlagUserMapping := new(utils.FeatureFlagUserMappingResponse)
	err = dynamodbattribute.UnmarshalMap(result.Item, &featureFlagUserMapping)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	return featureFlagUserMapping, nil
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	corsResponse, err := middleware.CORSMiddleware()(req)
	if err != nil {
		log.Printf("CORS error: %v", err)
		return corsResponse, err
	}

	if corsResponse.StatusCode != http.StatusOK {
		return corsResponse, nil
	}

	response, _, err := jwt.JWTMiddleware()(req)
	if err != nil {
		log.Printf("JWT middleware error: %v", err)
		return response, err
	}

	if response.StatusCode != http.StatusOK {
		return response, nil
	}

	userId := req.PathParameters["userId"]

	flagId := req.PathParameters["flagId"]

	result, err := processGetById(userId, flagId)

	if err != nil {
		return utils.ServerError(err)
	}

	if result == nil {
		log.Println("User feature flag not found")
		return utils.ClientError(http.StatusNotFound, "User feature flag not found")
	}

	resultJson, err := json.Marshal(result)

	if err != nil {
		log.Println("Error converting featureFlagUserMapping to JSON")
		return utils.ServerError(err)
	}

	origin := req.Headers["Origin"]
	corsHeaders := middleware.GetCORSHeaders(origin)

	return events.APIGatewayProxyResponse{
		Body:       string(resultJson),
		StatusCode: http.StatusOK,
		Headers:    corsHeaders,
	}, nil

}

func main() {
	lambda.Start(handler)
}
