package main

import (
	"context"
	"net/http"

	"github.com/Real-Dev-Squad/feature-flag-backend/database"
	middleware "github.com/Real-Dev-Squad/feature-flag-backend/middlewares"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ctx := context.TODO()
	db := database.CreateDynamoDB()

	utils.CheckRequestAllowed(ctx, db, utils.ConcurrencyDisablingLambda)
	corsResponse, err, passed := middleware.HandleCORS(req)
	if !passed {
		return corsResponse, err
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
