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
)

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

	// Check permission: READ_FEATURE_FLAG
	permResponse, err := utils.RequirePermission(userContext, utils.PermissionReadFeatureFlag)
	if err != nil || permResponse.StatusCode != http.StatusOK {
		permResponse.Headers = middleware.GetCORSHeadersV1(req.Headers)
		return permResponse, err
	}

	corsHeaders := middleware.GetCORSHeadersV1(req.Headers)

	featureFlagId, ok := req.PathParameters["flagId"]
	if !ok {
		log.Println("flagId is required")
		clientErrorResponse, _ := utils.ClientError(http.StatusBadRequest, "flagId is required")
		return clientErrorResponse, nil
	}

	featureFlag, err := database.ProcessGetFeatureFlagByHashKey(utils.Id, featureFlagId)
	if err != nil {
		log.Printf("Database error: %v", err)
		serverErrorResponse, _ := utils.ServerError(err)
		return serverErrorResponse, nil
	}

	if featureFlag == nil {
		log.Println("Feature Flag not found")
		clientErrorResponse, _ := utils.ClientError(http.StatusNotFound, "Feature flag not found")
		return clientErrorResponse, nil
	}
	log.Println(featureFlag, " is the feature flag")

	jsonResponse, err := json.Marshal(featureFlag)
	if err != nil {
		log.Printf("Error converting FeatureFlag to JSON: %v", err)
		serverErrorResponse, _ := utils.ServerError(err)
		return serverErrorResponse, nil
	}

	response := events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    corsHeaders,
		Body:       string(jsonResponse),
	}
	return response, nil
}

func main() {
	lambda.Start(handler)
}
