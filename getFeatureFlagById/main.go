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
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	corsResponse, err, passed := middleware.HandleCORS(req)

	if !passed {
		return corsResponse, err
	}

	response, _, err := jwt.JWTMiddleware()(req)
	if err != nil || response.StatusCode != http.StatusOK {
		return response, err
	}

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

	origin := req.Headers["Origin"]
	corsHeaders := middleware.GetCORSHeaders(origin)

	response = events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    corsHeaders,
		Body:       string(jsonResponse),
	}
	return response, nil
}

func main() {
	lambda.Start(handler)
}
