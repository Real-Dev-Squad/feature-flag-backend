package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/Real-Dev-Squad/feature-flag-backend/database"
	"github.com/Real-Dev-Squad/feature-flag-backend/jwt"
	"github.com/Real-Dev-Squad/feature-flag-backend/middleware"
	"github.com/Real-Dev-Squad/feature-flag-backend/models"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-lambda-go/events"
	lambda1 "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	lambda "github.com/aws/aws-sdk-go/service/lambda"
)

type Request struct {
	FunctionNames []string `json:"functionNames"`
}

type LambdaConcurrencyValue struct {
	IntValue int `json:"intValue"`
}

type ConcurrencyLimitRequest struct {
	PendingLimit int16 `json:"pendingLimit"`
}

var createFeatureFlagFunctionName string
var getUserFeatureFlagFunctionName string
var createUserFeatureFlagFunctionName string
var getAllFeatureFlagsFunctionName string
var getUserFeatureFlagsFunctionName string
var updateFeatureFlagFunctionName string
var getFeatureFlagFunctionName string
var getUserFeatureFlagFunction string
var requestLimitTableName = "requestLimit"

func init() {
	env, found := os.LookupEnv(utils.ENV)
	if !found {
		log.Print("Env variable not set, making it by default PROD")
		os.Setenv(utils.ENV, utils.PROD)
	}
	log.Printf("The env is %v", env)

	createFeatureFlagFunctionName, found = os.LookupEnv("CreateFeatureFlagFunction")
	if !found {
		log.Println("Create feature flag function name not being set")
	}

	getUserFeatureFlagFunctionName, found = os.LookupEnv("GetUserFeatureFlagFunction")
	if !found {
		log.Println("Create feature flag function name not being set")
	}

	createUserFeatureFlagFunctionName, found = os.LookupEnv("CreateUserFeatureFlagFunction")
	if !found {
		log.Println("Create user feature flag function name not being set")
	}

	getUserFeatureFlagsFunctionName, found = os.LookupEnv("GetUserFeatureFlagsFunction")
	if !found {
		log.Println("get user feature flags function name not being set")
	}

	getAllFeatureFlagsFunctionName, found = os.LookupEnv("GetAllFeatureFlagFunction")
	if !found {
		log.Println("get all feature flag function name not being set")
	}

	updateFeatureFlagFunctionName, found = os.LookupEnv("UpdateFeatureFlagFunction")
	if !found {
		log.Println("Update feature flag function name not being set")
	}

	getFeatureFlagFunctionName, found = os.LookupEnv("GetFeatureFlagFunction")
	if !found {
		log.Println("get feature flag function name not being set")
	}
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	corsResponse, err := middleware.CORSMiddleware()(event)
	if err != nil {
		log.Printf("CORS error: %v", err)
		return corsResponse, err
	}

	if corsResponse.StatusCode != http.StatusOK {
		return corsResponse, nil
	}

	jwtResponse, _, err := jwt.JWTMiddleware()(event)
	if err != nil {
		log.Printf("JWT middleware error: %v", err)
		return jwtResponse, err
	}

	if jwtResponse.StatusCode != http.StatusOK {
		return jwtResponse, nil
	}

	var concurrencyLimitRequest ConcurrencyLimitRequest
	if err := json.Unmarshal([]byte(event.Body), &concurrencyLimitRequest); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Unable to read input",
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	sess, err := session.NewSession()
	if err != nil {
		log.Println("Error in creation of AWS session, please contact on #feature-flag-service discord channel.")
	}
	lambdaClient := lambda.New(sess)

	var request = Request{
		FunctionNames: []string{createFeatureFlagFunctionName,
			createUserFeatureFlagFunctionName,
			getFeatureFlagFunctionName,
			getUserFeatureFlagFunctionName,
			getAllFeatureFlagsFunctionName,
			updateFeatureFlagFunctionName,
			getUserFeatureFlagsFunctionName,
		},
	}

	var wg sync.WaitGroup
	for _, functionName := range request.FunctionNames {
		wg.Add(1)
		go func(fn string) {
			defer wg.Done()
			input := &lambda.DeleteFunctionConcurrencyInput{
				FunctionName: &fn,
			}
			_, err := lambdaClient.DeleteFunctionConcurrency(input)
			if err != nil {
				log.Printf("Error in resetting the concurrency for the lambda name %s: %v", fn, err)
				utils.ServerError(err)
			}
			log.Printf("Reset the reserved concurrency for the function %s", fn)
		}(functionName)
	}
	wg.Wait()

	origin := event.Headers["Origin"]
	corsHeaders := middleware.GetCORSHeaders(origin)

	err = updateConcurrencyLimitInDB(concurrencyLimitRequest.PendingLimit)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Failed to update concurrency limit in database",
			StatusCode: http.StatusInternalServerError,
			Headers:    corsHeaders,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		Body:       "Successfully updated concurrency limit",
		StatusCode: http.StatusOK,
		Headers:    corsHeaders,
	}, nil
}

func updateConcurrencyLimitInDB(limit int16) error {
	db := database.CreateDynamoDB()

	requestLimitInput := &dynamodb.GetItemInput{
		TableName: aws.String(requestLimitTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"limitType": {
				S: aws.String("pendingLimit"),
			},
		},
	}

	requestLimitResult, err := db.GetItem(requestLimitInput)
	if err != nil {
		log.Println(err, "is the error in request limit fetching")
	}

	requestLimitResponse := new(models.RequestLimit)
	err = dynamodbattribute.UnmarshalMap(requestLimitResult.Item, requestLimitResponse)

	if err != nil {
		log.Println(err, "is the error")
	}

	requestLimitUpdateInput := models.RequestLimit{
		LimitType:  requestLimitResponse.LimitType,
		LimitValue: limit,
	}

	marshalledInput, err := dynamodbattribute.MarshalMap(requestLimitUpdateInput)
	if err != nil {
		log.Println("Error in marshalling the request")
	}

	putItemInput := &dynamodb.PutItemInput{
		TableName: aws.String(requestLimitTableName),
		Item:      marshalledInput,
	}

	_, err = db.PutItem(putItemInput)
	if err != nil {
		log.Println("Error in updating the request limit counters", err)
		return nil
	}
	log.Println("The updated limit is ", requestLimitUpdateInput.LimitValue)
	return nil
}

func main() {
	lambda1.Start(handler)
}
