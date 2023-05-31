package utils

import (
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/go-playground/validator/v10"
)

type validationError struct {
	Field string
	Error string
}

func ClientError(statusCode int, body string) (events.APIGatewayProxyResponse, error) {
	//check if the status sent is in the range of 400 and 500.
	if !(statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError) {
		log.Printf("Wrong Status code used: %d for Client Error, allowed range is %d to %d", statusCode, http.StatusBadRequest, http.StatusInternalServerError)

		return events.APIGatewayProxyResponse{
			Body:       "Something went wrong, please try again.",
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	//return the response to the user
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

	//logging for internal use
	log.Printf("Internal Server Error: %v", err)
	return events.APIGatewayProxyResponse{
		Body:       errMsg,
		StatusCode: http.StatusInternalServerError,
	}, nil
}

func DdbError(err error) {
	if aerr, ok := err.(awserr.Error); ok {
		log.Printf("Error code %s, Error message %s", aerr.Code(), aerr.Error())
	} else {
		log.Println(err.Error())
	}
}

func HandleValidationError(err error) []validationError {
	var errors []validationError

	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, validationError{
			Field: err.Field(),
			Error: err.Tag(),
		})
		log.Println("Errors are ", errors)
	}
	return errors
}

func ValidateFeatureFlagStatus(status string) bool {
	allowedStatuses := map[string]bool{
		"ENABLED":  true,
		"DISABLED": true,
	}

	_, found := allowedStatuses[strings.ToUpper(status)]
	return found
}
