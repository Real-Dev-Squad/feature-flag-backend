package main

import (
	"encoding/json"
	"github.com/Real-Dev-Squad/feature-flag-backend/database"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"net/http"
)

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id, found := req.PathParameters["flagId"]

	if !found {
		log.Println("flagId not passed")
		utils.ClientError(http.StatusBadRequest, "Flag ID not passed in request")
	}
	featureFlag, err := database.ProcessGetFeatureFlagByHashKey(utils.Id, id)
	if err != nil {
		utils.ServerError(err)
	}

	if featureFlag == nil {
		log.Println("Feature Flag not found")
		return utils.ClientError(http.StatusNotFound, "Feature flag not found")
		
	}
	log.Println(featureFlag, " is the feature flag")
	
	//marshal the item to JSON
	jsonResponse, err := json.Marshal(featureFlag)
	if err != nil {
		log.Println("Error converting FeatureFlag to JSON")
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
