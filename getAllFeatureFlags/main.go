package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"feature-flag-backend/layer/database"
	"feature-flag-backend/layer/jwt"
	middleware "feature-flag-backend/layer/middlewares"
	"feature-flag-backend/layer/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func getAllFeatureFlags(ctx context.Context, db *dynamodb.Client) ([]utils.FeatureFlagResponse, error) {

	input := &dynamodb.ScanInput{
		TableName: aws.String(utils.FEATURE_FLAG_TABLE_NAME),
	}
	result, err := db.Scan(ctx, input)

	if err != nil {
		utils.DdbError(err)
		return nil, err
	}

	if len(result.Items) == 0 {
		return []utils.FeatureFlagResponse{}, nil
	}

	var featureFlagsResponse []utils.FeatureFlagResponse

	err = attributevalue.UnmarshalListOfMaps(result.Items, &featureFlagsResponse)
	if err != nil {
		log.Println("Something went wrong in unmarshalling all feature flags response", err)
		return nil, err
	}

	return featureFlagsResponse, nil
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	db := database.CreateDynamoDB()

	utils.CheckRequestAllowed(ctx, db, utils.ConcurrencyDisablingLambda)

	corsResponse, err, passed := middleware.HandleCORS(request)
	if !passed {
		return corsResponse, err
	}

	// Use enhanced middleware with user verification and RBAC (Week 3)
	jwtResponse, userContext, err := jwt.JWTMiddlewareWithUserVerification()(request)
	if err != nil || jwtResponse.StatusCode != http.StatusOK {
		return jwtResponse, err
	}

	// Check permission: READ_FEATURE_FLAG
	permResponse, err := utils.RequirePermission(userContext, utils.PermissionReadFeatureFlag)
	if err != nil || permResponse.StatusCode != http.StatusOK {
		permResponse.Headers = middleware.GetCORSHeadersV1(request.Headers)
		return permResponse, err
	}
	
	corsHeaders := middleware.GetCORSHeadersV1(request.Headers)

	featureFlagsResponse, err := getAllFeatureFlags(ctx, db)
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
		Headers:    corsHeaders,
	}, nil
}

func main() {
	lambda.Start(handler)
}
