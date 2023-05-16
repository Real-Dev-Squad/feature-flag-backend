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
	errorMessage := "Something went wrong, please try again."

	//logging for internal use
	log.Printf("Internal Server Error: %v", err)
	return events.APIGatewayProxyResponse{
		Body:       errorMessage,
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
