package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/Real-Dev-Squad/feature-flag-backend/database"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"

)

var checkRequestAllowed = utils.CheckRequestAllowed
func handler(request events.APIGatewayProxyRequest)(events.APIGatewayProxyResponse, error){
	db := database.CreateDynamoDB()

	checkRequestAllowed(db, utils.ConcurrencyDisablingLambda)
	return events.APIGatewayProxyResponse{
		Body: "Server health is good!!",
		StatusCode: 200,
	}, nil
}

func main(){
	lambda.Start(handler)
}