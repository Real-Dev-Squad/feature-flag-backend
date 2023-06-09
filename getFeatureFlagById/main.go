package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Real-Dev-Squad/feature-flag-backend/database"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id, _ := req.PathParameters["flagId"]
	featureFlag, err := database.ProcessGetFeatureFlagByHashKey(utils.Id, id)
	if err != nil {
		return utils.ServerError(err)
	}

	if featureFlag == nil {
		log.Println("Feature Flag not found")
		return utils.ClientError(http.StatusNotFound, "Feature flag not found")

	}
	log.Println(featureFlag, " is the feature flag")

	jsonResponse, err := json.Marshal(featureFlag)
	if err != nil {
		log.Printf("Error converting FeatureFlag to JSON \n %v", err)
		return utils.ServerError(err)
	}

	response := events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(jsonResponse),
	}
	return response, nil
}

func main() {
	lambda.Start(handler)
}
