package main

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/aws/aws-sdk-go/service/dynamodb"

)

func TestHandler(t *testing.T) {

	checkRequestAllowed = func(db *dynamodb.DynamoDB, concurrencyValue int) {
        // Simulate the behavior of the function (e.g., do nothing or log)
    }
	
	jwtHandler = func() func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, string, error) {
        return func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, string, error) {
            // Simulate different scenarios based on the "Cookie" header
            cookie := req.Headers["Cookie"]
            if cookie == "" {
                return events.APIGatewayProxyResponse{
                    StatusCode: http.StatusUnauthorized,
                    Body:       "Unauthenticated",
                }, "", nil
            }

            if cookie == "valid-cookie" {
                return events.APIGatewayProxyResponse{
                    StatusCode: http.StatusOK,
                }, "valid-user-id", nil
            }

            return events.APIGatewayProxyResponse{
                StatusCode: http.StatusUnauthorized,
                Body:       "Invalid token",
            }, "", nil
        }
    }

	tests := []struct {
		name         string
		request      events.APIGatewayProxyRequest
		responseBody string
		statusCode   int
		headers      map[string]string
	}{
		{
			name: "Valid OPTIONS request with Origin header",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: "OPTIONS",
				Headers: map[string]string{
					"Origin": "https://realdevsquad.com",
					"Cookie":  "valid-cookie",
				},
			},
			responseBody: "",
			statusCode:   http.StatusOK,
			headers: map[string]string{
				"Access-Control-Allow-Origin":      "https://realdevsquad.com",
				"Access-Control-Allow-Methods":     "GET, POST, OPTIONS, PATCH",
				"Access-Control-Allow-Headers":     "Authorization, Content-Type, Cache-Control, Cookie",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name: "OPTIONS request without Origin header",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: "OPTIONS",
				Headers:    map[string]string{},
			},
			responseBody: "CORS policy: No origin header found.",
			statusCode:   http.StatusForbidden,
			headers:      nil,
		},
		{
			name: "Non-OPTIONS request with valid Origin header",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				Headers: map[string]string{
					"Origin": "https://realdevsquad.com",
					"Cookie":  "valid-cookie",
				},
			},
			responseBody: "",
			statusCode:   http.StatusOK,
			headers: map[string]string{
				"Access-Control-Allow-Origin":      "https://realdevsquad.com",
				"Access-Control-Allow-Methods":     "GET, POST, OPTIONS, PATCH",
				"Access-Control-Allow-Headers":     "Authorization, Content-Type, Cache-Control, Cookie",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name: "Non-OPTIONS request without Origin header",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				Headers:    map[string]string{},
			},
			responseBody: "CORS policy: No origin header found.",
			statusCode:   http.StatusForbidden,
			headers:      nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := handler(test.request)

			assert.NoError(t, err)
			assert.Equal(t, test.statusCode, response.StatusCode)
			assert.Equal(t, test.responseBody, response.Body)

			// Check headers if they are expected
			if test.headers != nil {
				for key, value := range test.headers {
					assert.Equal(t, value, response.Headers[key])
				}
			}
		})
	}
}
