package main

import (
	"net/http"

	"github.com/Real-Dev-Squad/feature-flag-backend/jwt"
	"github.com/Real-Dev-Squad/feature-flag-backend/database"
	middleware "github.com/Real-Dev-Squad/feature-flag-backend/middlewares"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var checkRequestAllowed = utils.CheckRequestAllowed
var jwtHandler = jwt.JWTMiddleware

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	db := database.CreateDynamoDB()

	checkRequestAllowed(db, utils.ConcurrencyDisablingLambda)
	corsResponse, err, passed := middleware.HandleCORS(req)
	if !passed {
		return corsResponse, err
	}
	
	response, _, err := jwtHandler()(req)
	if err != nil || response.StatusCode != http.StatusOK {
		return response, err
	}
	corsHeaders := middleware.GetCORSHeadersV1(req.Headers)

	return events.APIGatewayProxyResponse{
		Body:       "",
		StatusCode: http.StatusOK,
		Headers:    corsHeaders,
	}, nil
}

func main() {
	lambda.Start(handler)
}
