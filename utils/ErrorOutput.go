package utils

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
	"net/http"
)

func ClientError(status int, body string) (events.APIGatewayProxyResponse, error) {

	resp := events.APIGatewayProxyResponse{
		Body:       body,
		StatusCode: status,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	return resp, nil
}
func ServerError(err error) (events.APIGatewayProxyResponse, error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	return events.APIGatewayProxyResponse{
		Body:       errMsg,
		StatusCode: http.StatusInternalServerError,
	}, fmt.Errorf("Internal server error: %v", err)
}

func DdbError(err error) {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case dynamodb.ErrCodeProvisionedThroughputExceededException:
			log.Println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
		case dynamodb.ErrCodeResourceNotFoundException:
			log.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error(), "is the aws error")
		case dynamodb.ErrCodeRequestLimitExceeded:
			log.Println(dynamodb.ErrCodeRequestLimitExceeded, aerr.Error())
		case dynamodb.ErrCodeInternalServerError:
			log.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
		case dynamodb.ErrCodeConditionalCheckFailedException:
			log.Println(dynamodb.ErrCodeConditionalCheckFailedException, aerr.Error())
		case dynamodb.ErrCodeTransactionConflictException:
			log.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
		default:
			log.Println(aerr.Error())
		}
	} else {
		log.Println(err.Error())
	}
}
