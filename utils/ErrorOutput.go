package utils

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/awserr"
)

type ValidationError struct {
	Field string
	Error string
}

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
		log.Printf("Error code %s, Error message %s", aerr.Code(), aerr.Error())
	} else {
		log.Println(err.Error())
	}
}
