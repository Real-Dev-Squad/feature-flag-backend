package main

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {

	tests := []struct {
		request      events.APIGatewayProxyRequest
		responseBody string
		statusCode   int
		err          error
	}{
		{
			request:      events.APIGatewayProxyRequest{},
			responseBody: "Server health is good!!",
			statusCode:   http.StatusOK,
			err:          nil,
		},
	}

	for _, test := range tests {
		response, err := handler(test.request)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.statusCode, response.StatusCode)
		assert.Equal(t, test.responseBody, response.Body)
	}

}
