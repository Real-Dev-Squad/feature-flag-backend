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

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	db := database.CreateDynamoDB()

	utils.CheckRequestAllowed(db, utils.ConcurrencyDisablingLambda)
	corsResponse, err, passed := middleware.HandleCORS(req)
	if !passed {
		return corsResponse, err
	}
	
	jwtHandler := jwt.JWTMiddleware()
	response, _, err := jwtHandler(req)
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
