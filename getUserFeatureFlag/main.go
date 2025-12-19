package main

import (
	"encoding/json"
	"log"
	"net/http"

	"feature-flag-backend/layer/database"
	"feature-flag-backend/layer/jwt"
	middleware "feature-flag-backend/layer/middlewares"
	"feature-flag-backend/layer/utils"
	"github.com/aws/aws-lambda-go/events"
	lambda "github.com/aws/aws-lambda-go/lambda"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func processGetById(ctx context.Context, userId string, flagId string) (*utils.FeatureFlagUserMappingResponse, error) {

	db := database.CreateDynamoDB()

	utils.CheckRequestAllowed(ctx, db, utils.ConcurrencyDisablingLambda)

	input := &dynamodb.GetItemInput{
		TableName: aws.String(utils.FEATURE_FLAG_USER_MAPPING_TABLE_NAME),
		Key: map[string]types.AttributeValue{
			utils.UserId: &types.AttributeValueMemberS{ // partition key
				Value: userId,
			},
			utils.FlagId: &types.AttributeValueMemberS{ // sort key
				Value: flagId,
			},
		},
	}

	result, err := db.GetItem(ctx, input)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}

	featureFlagUserMapping := new(utils.FeatureFlagUserMappingResponse)
	err = attributevalue.UnmarshalMap(result.Item, featureFlagUserMapping)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	return featureFlagUserMapping, nil
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

	flagId := req.PathParameters["flagId"]

	result, err := processGetById(ctx, userId, flagId)

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

	return events.APIGatewayProxyResponse{
		Body:       string(resultJson),
		StatusCode: http.StatusOK,
		Headers:    corsHeaders,
	}, nil

}

func main() {
	lambda.Start(handler)
}
