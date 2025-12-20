package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"feature-flag-backend/layer/database"
	"feature-flag-backend/layer/jwt"
	middleware "feature-flag-backend/layer/middlewares"
	"feature-flag-backend/layer/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func processGetById(ctx context.Context, userId string) ([]utils.FeatureFlagUserMappingResponse, error) {

	db := database.CreateDynamoDB()

	utils.CheckRequestAllowed(ctx, db, utils.ConcurrencyDisablingLambda)

	input := &dynamodb.QueryInput{
		TableName:              aws.String(utils.FEATURE_FLAG_USER_MAPPING_TABLE_NAME),
		KeyConditionExpression: aws.String(fmt.Sprintf("%v = :uid", utils.UserId)),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{
				Value: userId,
			},
		},
	}

	result, err := db.Query(ctx, input)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, nil
	}

	var listOfFeatureFlagUserMapping []utils.FeatureFlagUserMappingResponse

	err = attributevalue.UnmarshalListOfMaps(result.Items, &listOfFeatureFlagUserMapping)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return listOfFeatureFlagUserMapping, nil
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	corsResponse, err, passed := middleware.HandleCORS(req)
	if !passed {
		return corsResponse, err
	}

	response, _, err := jwt.JWTMiddleware()(req)
	if err != nil || response.StatusCode != http.StatusOK {
		return response, err
	}

	corsHeaders := middleware.GetCORSHeadersV1(req.Headers)

	userId := req.PathParameters["userId"]
	result, err := processGetById(ctx, userId)
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
		Headers:    corsHeaders,
	}, nil

}

func main() {
	lambda.Start(handler)
}
