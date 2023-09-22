package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/Real-Dev-Squad/feature-flag-backend/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	lambdaInvoke "github.com/aws/aws-sdk-go/service/lambda"
)

var raterLimiterFunctionName string
var found bool
var requestLimitTableName = "requestLimit"

func init() {
	raterLimiterFunctionName, found = os.LookupEnv("RateLimiterFunction")
	if !found {
		log.Println("Rate limiter function name is not found")
	}
}

func CheckRequestAllowed(db *dynamodb.DynamoDB, concurrencyValue int) {
	// check the quota values
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

	if requestLimitResponse.LimitValue > 1 {
		// reducing the request quota limit
		requestLimitUpdateInput := models.RequestLimit{
			LimitType:  requestLimitResponse.LimitType,
			LimitValue: requestLimitResponse.LimitValue - 1,
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
			return
		}
		log.Println("The updated limit is ", requestLimitUpdateInput.LimitValue)
	} else {
		//mark the concurrency of all the other lambdas to zero
		sess, err := session.NewSession()

		if err != nil {
			log.Println("Error in creating AWS session to access any service")
			ServerError(err)
		}
		lambdaClient := lambdaInvoke.New(sess)

		concurrencyValue := 0
		lambdaInvokeInput := lambdaInvoke.InvokeInput{
			FunctionName: aws.String(raterLimiterFunctionName),
			Payload:      []byte(fmt.Sprintf(`{"intValue" : %d}`, concurrencyValue)),
		}
		result, err := lambdaClient.Invoke(&lambdaInvokeInput)
		if err != nil {
			log.Println("There is some error in calling the new lambda created")
			ServerError(err)
		}

		log.Println("The result of the invocation of the another lambda is ", result)

	}
}
