package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/Real-Dev-Squad/feature-flag-backend/models"
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

var (
	createFeatureFlagFunctionName     string
	getUserFeatureFlagFunctionName    string
	createUserFeatureFlagFunctionName string
	getAllFeatureFlagsFunctionName    string
	getUserFeatureFlagsFunctionName   string
	updateFeatureFlagFunctionName     string
	getFeatureFlagFunctionName        string
	getUserFeatureFlagFunction        string
	db                                *dynamodb.DynamoDB
	requestLimitTableName             = "requestLimit"
)

func init() {
	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("Failed to create AWS session: %v", err)
	}

	db = dynamodb.New(sess)

	env, found := os.LookupEnv("ENV")
	if !found {
		log.Print("Env variable not set, making it by default PROD")
		os.Setenv("ENV", "PROD")
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

func handler(ctx context.Context, event json.RawMessage) (events.APIGatewayProxyResponse, error) {
	sess, err := session.NewSession()
	if err != nil {
		log.Println("Error in creation of AWS session, please contact on #feature-flag-service discord channel.")
	}
	lambdaClient := lambda.New(sess)

	var lambdaConcurrencyValue LambdaConcurrencyValue
	if err := json.Unmarshal(event, &lambdaConcurrencyValue); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Unable to read input",
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	var request = Request{
		FunctionNames: []string{
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
	for _, functionName := range request.FunctionNames {
		wg.Add(1)
		go func(fn string) {
			defer wg.Done()
			input := &lambda.PutFunctionConcurrencyInput{
				FunctionName:                 &fn,
				ReservedConcurrentExecutions: aws.Int64(int64(lambdaConcurrencyValue.IntValue)),
			}
			log.Println("Is the function name", fn)
			_, err := lambdaClient.PutFunctionConcurrency(input)
			if err != nil {
				log.Printf("Error in setting the concurrency for the lambda name %s: %v", fn, err)
			} else {
				log.Printf("Changed the reserved concurrency for the function %s to %d", fn, lambdaConcurrencyValue.IntValue)
			}
		}(functionName)
	}
	wg.Wait()

	return events.APIGatewayProxyResponse{
		Body:       "Changed the reserved concurrency of the lambda functions",
		StatusCode: 200,
	}, nil
}

func resetLimitHandler(ctx context.Context, event json.RawMessage) (events.APIGatewayProxyResponse, error) {
	var lambdaConcurrencyValue LambdaConcurrencyValue
	if err := json.Unmarshal(event, &lambdaConcurrencyValue); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Unable to read input",
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	err := updateConcurrencyLimitInDB(lambdaConcurrencyValue.IntValue)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Failed to update concurrency limit in database",
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		Body:       "Successfully updated concurrency limit",
		StatusCode: http.StatusOK,
	}, nil
}

func updateConcurrencyLimitInDB(limit int) error {
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
	http.HandleFunc("/reset-limit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			ctx := context.Background()
			event := json.RawMessage{}
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&event); err != nil {
				http.Error(w, "Unable to read input", http.StatusBadRequest)
				return
			}

			response, err := resetLimitHandler(ctx, event)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(response.StatusCode)
			w.Write([]byte(response.Body))
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	lambda1.Start(handler)
}
