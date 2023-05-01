package utils

import (
	"github.com/aws/aws-lambda-go/events"
	"log"
	"net/http"
)

func ClientError(status int) (events.APIGatewayProxyResponse, error) {

	return events.APIGatewayProxyResponse{
		Body:       http.StatusText(status),
		StatusCode: status,
	}, nil
}

func ServerError(err error) (events.APIGatewayProxyResponse, error) {
	log.Println(err)

	return events.APIGatewayProxyResponse{
		Body:       http.StatusText(http.StatusInternalServerError),
		StatusCode: http.StatusInternalServerError,
	}, nil
}
