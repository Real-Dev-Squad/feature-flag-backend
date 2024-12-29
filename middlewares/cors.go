package cors

import (
	"log"
	"net/http"
	"regexp"

	"github.com/aws/aws-lambda-go/events"
)

var AllowedOrigins = []*regexp.Regexp{
	regexp.MustCompile(`^https?://([a-zA-Z0-9-]+\.)*realdevsquad\.com$`),
}

func generateCORSHeaders(origin string) map[string]string {
	return map[string]string{
		"Access-Control-Allow-Origin":      origin,
		"Access-Control-Allow-Methods":     "GET, POST, OPTIONS",
		"Access-Control-Allow-Headers":     "Authorization, Content-Type, Cache-Control, Cookie",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Expose-Headers":    "Set-Cookie",
		"Vary":                             "Origin",
	}
}

func GetCORSHeaders(origin string) map[string]string {
	for _, pattern := range AllowedOrigins {
		if pattern.MatchString(origin) {
			return generateCORSHeaders(origin)
		}
	}

	return generateCORSHeaders("null")
}

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

		isRDSDomain := AllowedOrigins[0].MatchString(origin)

		if isRDSDomain {
			headers := generateCORSHeaders(origin)

			if req.HTTPMethod == "OPTIONS" {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusOK,
					Headers:    headers,
					Body:       "",
				}, nil
			}

			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
				Headers:    headers,
			}, nil
		}

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusForbidden,
			Body:       "CORS policy does not allow access from this origin.",
		}, nil
	}
}
