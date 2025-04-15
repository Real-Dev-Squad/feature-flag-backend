package main

import (
	"github.com/Real-Dev-Squad/feature-flag-backend/jwt"
	middleware "github.com/Real-Dev-Squad/feature-flag-backend/middlewares"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
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
	corsHeaders := middleware.GetCORSHeaders(req.Headers)

	return events.APIGatewayProxyResponse{
		Body:       "",
		StatusCode: http.StatusOK,
		Headers:    corsHeaders,
	}, nil
}

func main() {
	lambda.Start(handler)
}
