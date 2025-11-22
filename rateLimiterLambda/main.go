package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-lambda-go/events"
	lambda1 "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	lambda "github.com/aws/aws-sdk-go-v2/service/lambda"
)

type Request struct {
	FunctionNames []string `json:"functionNames"`
}

type LambdaConcurrencyValue struct {
	IntValue int `json:"intValue"`
}

var createFeatureFlagFunctionName string
var getUserFeatureFlagFunctionName string
var createUserFeatureFlagFunctionName string
var getAllFeatureFlagsFunctionName string
var getUserFeatureFlagsFunctionName string
var updateFeatureFlagFunctionName string
var getFeatureFlagFunctionName string
var corsFunctionName string

func init() {
	env, found := os.LookupEnv(utils.ENV)
	if !found {
		log.Print("Env variable not set, making it by default PROD")
		os.Setenv(utils.ENV, utils.PROD)
	}
	log.Printf("The env is %v", env)

	corsFunctionName, found = os.LookupEnv("CorsLambda")
	if !found {
		log.Println("CORS function name not being set")
	}

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

func handler(ctx context.Context, event json.RawMessage) (events.APIGatewayProxyResponse, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Error loading AWS config: %v", err)
		return events.APIGatewayProxyResponse{
			Body:       "Internal server error: failed to initialize AWS configuration",
			StatusCode: http.StatusInternalServerError,
		}, nil
	}
	lambdaClient := lambda.NewFromConfig(cfg)

	var lambdaConcurrencyValue LambdaConcurrencyValue
	if err := json.Unmarshal(event, &lambdaConcurrencyValue); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Unable to read input",
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	var request = Request{
		FunctionNames: []string{
			corsFunctionName,
			createFeatureFlagFunctionName,
			createUserFeatureFlagFunctionName,
			getFeatureFlagFunctionName,
			getUserFeatureFlagFunctionName,
			getAllFeatureFlagsFunctionName,
			updateFeatureFlagFunctionName,
			getUserFeatureFlagsFunctionName,
		},
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(request.FunctionNames))
	
	for _, functionName := range request.FunctionNames {
		// Increment the WaitGroup counter
		wg.Add(1)

		// Start a goroutine to update the concurrency for the Lambda function
		go func(fn string) {
			defer wg.Done()

			input := &lambda.PutFunctionConcurrencyInput{
				FunctionName:                 aws.String(fn),
				ReservedConcurrentExecutions: aws.Int32(int32(lambdaConcurrencyValue.IntValue)),
			}

			log.Println("Is the function name", fn)
			_, err := lambdaClient.PutFunctionConcurrency(ctx, input)
			if err != nil {
				log.Printf("Error in setting the concurrency for the lambda name %s: %v", fn, err)
				errChan <- err
				return
			}

			log.Printf("Changed the reserved concurrency for the function %s to %d", fn, lambdaConcurrencyValue.IntValue)
		}(functionName)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)

	// Collect any errors from goroutines
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	// If any operations failed, return an error response
	if len(errors) > 0 {
		log.Printf("Failed to update concurrency for %d function(s)", len(errors))
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("Failed to update concurrency for %d function(s)", len(errors)),
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		Body:       "Changed the reserved concurrency of the lambda function GetFeatureFlagFunction",
		StatusCode: http.StatusOK,
	}, nil

}

func main() {

	lambda1.Start(handler)
}
