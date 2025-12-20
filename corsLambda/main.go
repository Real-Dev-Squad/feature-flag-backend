package main

import (
	"context"
	"net/http"

	"feature-flag-backend/layer/database"
	middleware "feature-flag-backend/layer/middlewares"
	"feature-flag-backend/layer/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
