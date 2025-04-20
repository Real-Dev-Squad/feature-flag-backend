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
		"Access-Control-Allow-Methods":     "GET, POST, OPTIONS, PATCH",
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

func GetCORSHeadersV1(headers map[string]string) map[string]string {
	var origin string
	if val, exists := headers["Origin"]; exists {
		origin = val
	} else if val, exists := headers["origin"]; exists {
		origin = val
	}

	for _, pattern := range AllowedOrigins {
		if pattern.MatchString(origin) {
			return generateCORSHeaders(origin)
		}
	}

	return generateCORSHeaders("null")
}

func HandleCORS(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error, bool) {
	corsResponse, err := CORSMiddleware()(req)
	if err != nil {
		log.Printf("CORS error: %v", err)
		return corsResponse, err, false
	}

	if corsResponse.StatusCode != http.StatusOK {
		return corsResponse, nil, false
	}

	return corsResponse, nil, true
}

func CORSMiddleware() func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		log.Println("Received Headers:", req.Headers)

		var origin string
		if val, exists := req.Headers["Origin"]; exists {
			origin = val
		} else if val, exists := req.Headers["origin"]; exists {
			origin = val
		} else {
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
