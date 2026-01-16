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

	// Use enhanced middleware with user verification and RBAC (Week 3)
	jwtResponse, userContext, err := jwt.JWTMiddlewareWithUserVerification()(req)
	if err != nil || jwtResponse.StatusCode != http.StatusOK {
		return jwtResponse, err
	}

	// Check permission: READ_USER_MAPPING (Week 3 RBAC)
	permResponse, err := utils.RequirePermission(userContext, utils.PermissionReadUserMapping)
	if err != nil || permResponse.StatusCode != http.StatusOK {
		permResponse.Headers = middleware.GetCORSHeadersV1(req.Headers)
		return permResponse, err
	}

	corsHeaders := middleware.GetCORSHeadersV1(req.Headers)

	userId := req.PathParameters["userId"]
	
	// Check if user can access this resource (own resources or ADMIN)
	if !utils.CanAccessUserResource(userContext, userId) {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusForbidden,
			Body:       "You can only access your own feature flag mappings",
			Headers:    corsHeaders,
		}, nil
	}

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
