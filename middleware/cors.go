package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

func CORSMiddleware() func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		log.Println("Received Headers:", req.Headers)
		origin, exists := req.Headers["Origin"]
		if !exists {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusForbidden,
				Body:       "CORS policy: No origin header found.",
			}, nil
		}

		if strings.HasSuffix(origin, ".realdevsquad.com") ||
			origin == "https://realdevsquad.com" ||
			origin == "http://127.0.0.1:3000" ||
			origin == "http://localhost:8080" {

			if req.HTTPMethod == "OPTIONS" {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusOK,
					Headers: map[string]string{
						"Access-Control-Allow-Origin":      origin,
						"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS",
						"Access-Control-Allow-Headers":     "Authorization, Content-Type, Cache-Control",
						"Access-Control-Allow-Credentials": "true",
					},
					Body: "",
				}, nil
			}

			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
				Headers: map[string]string{
					"Access-Control-Allow-Origin":      origin,
					"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS",
					"Access-Control-Allow-Headers":     "Authorization, Content-Type, Cache-Control",
					"Access-Control-Allow-Credentials": "true",
				},
			}, nil
		}

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusForbidden,
			Body:       "CORS policy does not allow access from this origin.",
		}, nil
	}
}
