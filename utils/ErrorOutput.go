package utils

import (
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/awserr"
)

func ClientError(statusCode int, body string) (events.APIGatewayProxyResponse, error) {
	if !(statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError) {
		log.Printf("Wrong Status code used: %d for Client Error, allowed range is %d to %d", statusCode, http.StatusBadRequest, http.StatusInternalServerError)

		return events.APIGatewayProxyResponse{
			Body:       "Something went wrong, please try again.",
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		Body:       body,
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func ServerError(err error) (events.APIGatewayProxyResponse, error) {
	errMsg := "Something went wrong, please try again."

	log.Printf("Internal Server Error:\n %v", err)
	return events.APIGatewayProxyResponse{
		Body:       errMsg,
		StatusCode: http.StatusInternalServerError,
	}, nil
}

func DdbError(err error) {
	if awsErr, ok := err.(awserr.Error); ok {
		log.Printf("Error code %s, Error message %s", awsErr.Code(), awsErr.Error())
	} else {
		log.Println(err.Error())
	}
}

func ValidateFeatureFlagStatus(status string) bool {
	allowedStatuses := map[string]bool{
		"ENABLED":  true,
		"DISABLED": true,
	}

	_, found := allowedStatuses[strings.ToUpper(status)]
	return found
}
