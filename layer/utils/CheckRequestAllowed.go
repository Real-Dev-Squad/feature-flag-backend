package utils

import (
	"context"
	"fmt"
	"log"
	"os"

	"feature-flag-backend/layer/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	lambdaInvoke "github.com/aws/aws-sdk-go-v2/service/lambda"
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

func CheckRequestAllowed(ctx context.Context, db *dynamodb.Client, concurrencyValue int) {
	// check the quota values
	requestLimitInput := &dynamodb.GetItemInput{
		TableName: aws.String(requestLimitTableName),
		Key: map[string]types.AttributeValue{
			"limitType": &types.AttributeValueMemberS{
				Value: "pendingLimit",
			},
		},
	}

	requestLimitResult, err := db.GetItem(ctx, requestLimitInput)
	if err != nil {
		log.Println(err, "is the error in request limit fetching")
	}

	requestLimitResponse := new(models.RequestLimit)
	err = attributevalue.UnmarshalMap(requestLimitResult.Item, requestLimitResponse)

	if err != nil {
		log.Println(err, "is the error")
	}

	if requestLimitResponse.LimitValue > 1 {
		// reducing the request quota limit
		requestLimitUpdateInput := models.RequestLimit{
			LimitType:  requestLimitResponse.LimitType,
			LimitValue: requestLimitResponse.LimitValue - 1,
		}

		marshalledInput, err := attributevalue.MarshalMap(requestLimitUpdateInput)
		if err != nil {
			log.Println("Error in marshalling the request")
		}

		putItemInput := &dynamodb.PutItemInput{
			TableName: aws.String(requestLimitTableName),
			Item:      marshalledInput,
		}

		_, err = db.PutItem(ctx, putItemInput)
		if err != nil {
			log.Println("Error in updating the request limit counters", err)
			return
		}
		log.Println("The updated limit is ", requestLimitUpdateInput.LimitValue)
	} else {
		//mark the concurrency of all the other lambdas to zero
		cfg, err := config.LoadDefaultConfig(ctx)

		if err != nil {
			log.Println("Error in creating AWS config to access any service")
			ServerError(err)
			return
		}
		lambdaClient := lambdaInvoke.NewFromConfig(cfg)

		concurrencyValue := 0
		lambdaInvokeInput := lambdaInvoke.InvokeInput{
			FunctionName: aws.String(raterLimiterFunctionName),
			Payload:      []byte(fmt.Sprintf(`{"intValue" : %d}`, concurrencyValue)),
		}
		result, err := lambdaClient.Invoke(ctx, &lambdaInvokeInput)
		if err != nil {
			log.Println("There is some error in calling the new lambda created")
			ServerError(err)
		}

		log.Println("The result of the invocation of the another lambda is ", result)

	}
}
