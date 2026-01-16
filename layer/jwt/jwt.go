package jwt

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"feature-flag-backend/layer/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtUtilsInstance *JWTUtils
	initError        error
	once             sync.Once
)

type JWTUtils struct {
	publicKey *rsa.PublicKey
}

type EnvConfig struct {
	SessionCookieName string
	Environment       string
	PublicKeyName     string
}

func LoadEnvConfig() (*EnvConfig, error) {
	return &EnvConfig{
		SessionCookieName: os.Getenv("SESSION_COOKIE_NAME"),
		Environment:       os.Getenv("ENVIRONMENT"),
		PublicKeyName:     os.Getenv("RDS_BACKEND_PUBLIC_KEY_NAME"),
	}, nil
}

func GetInstance() (*JWTUtils, error) {
	once.Do(func() {
		jwtUtilsInstance = &JWTUtils{}
		if err := jwtUtilsInstance.initialize(); err != nil {
			initError = fmt.Errorf("internal server error")
			jwtUtilsInstance = nil
		}
	})

	if initError != nil {
		return nil, initError
	}

	if jwtUtilsInstance == nil {
		return nil, errors.New("internal server error")
	}

	return jwtUtilsInstance, nil
}

func (j *JWTUtils) initialize() error {
	if j == nil {
		return errors.New("internal server error")
	}

	envConfig, _ := LoadEnvConfig()

	parameterName := envConfig.PublicKeyName
	if parameterName == "" {
		switch envConfig.Environment {
		case utils.PROD:
			parameterName = utils.RDS_BACKEND_PUBLIC_KEY_NAME_PROD
		case utils.DEV:
			parameterName = utils.RDS_BACKEND_PUBLIC_KEY_NAME_DEV
		default:
			parameterName = utils.RDS_BACKEND_PUBLIC_KEY_NAME_LOCAL
		}
	}

	log.Printf("Attempting to fetch public key from SSM parameter: %s (Environment: %s)", parameterName, envConfig.Environment)
	publicKeyString, err := getPublicKeyFromParameterStore(parameterName)
	if err != nil {
		log.Printf("Failed to get public key from SSM: %v", err)
		return err
	}
	publicKeyString = strings.TrimSpace(publicKeyString)
	
	block, _ := pem.Decode([]byte(publicKeyString))
	if block == nil {
		log.Printf("Failed to decode PEM block from public key. First 100 chars: %s", publicKeyString[:min(100, len(publicKeyString))])
		return errors.New("internal server error")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Printf("Failed to parse PKIX public key: %v", err)
		return fmt.Errorf("internal server error")
	}

	rsaPublicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		log.Printf("Public key is not an RSA public key")
		return errors.New("internal server error")
	}

	log.Printf("Successfully initialized JWT utils with public key")
	j.publicKey = rsaPublicKey
	return nil
}

func getPublicKeyFromParameterStore(parameterName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return "", err
	}

	svc := ssm.NewFromConfig(cfg)
	input := &ssm.GetParameterInput{
		Name:           aws.String(parameterName),
		WithDecryption: aws.Bool(true),
	}

	result, err := svc.GetParameter(ctx, input)
	if err != nil {
		log.Printf("Failed to get parameter %s from SSM: %v", parameterName, err)
		return "", err
	}

	if result.Parameter == nil || result.Parameter.Value == nil {
		log.Printf("Parameter %s exists but has no value", parameterName)
		return "", errors.New("parameter has no value")
	}

	value := strings.TrimSpace(*result.Parameter.Value)
	return value, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (j *JWTUtils) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	if j == nil || j.publicKey == nil {
		return nil, errors.New("internal server error")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("invalid token")
		}
		return j.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (j *JWTUtils) ExtractClaim(claims jwt.MapClaims, claimKey string) (string, error) {
	if claims == nil {
		return "", errors.New("internal server error")
	}

	value, ok := claims[claimKey].(string)
	if !ok || value == "" {
		return "", fmt.Errorf("unauthorized")
	}
	return value, nil
}

func handleMiddlewareResponse(statusCode int, message string) (events.APIGatewayProxyResponse, string, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       message,
	}, "", nil
}

func JWTMiddleware() func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, string, error) {
	return func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, string, error) {
		jwtUtils, err := GetInstance()
		if err != nil {
			log.Printf("Failed to get JWTUtils instance: %v", err)
			return handleMiddlewareResponse(http.StatusInternalServerError, "Internal server error")
		}

		cookie := ""
		// Check for cookie header in case-insensitive manner
		for key, val := range req.Headers {
			if strings.ToLower(key) == "cookie" {
				cookie = val
				break
			}
		}
		if cookie == "" {
			return handleMiddlewareResponse(http.StatusUnauthorized, "Unauthenticated")
		}

		envConfig, _ := LoadEnvConfig()

		cookieName := envConfig.SessionCookieName
		if cookieName == "" {
			switch envConfig.Environment {
			case utils.PROD:
				cookieName = utils.SESSION_COOKIE_NAME_PROD
			case utils.DEV:
				cookieName = utils.SESSION_COOKIE_NAME_DEV
			default:
				cookieName = utils.SESSION_COOKIE_NAME_LOCAL
			}
		}

		var jwtToken string
		cookies := strings.Split(cookie, ";")
		for _, c := range cookies {
			c = strings.TrimSpace(c)
			if strings.HasPrefix(c, cookieName+"=") {
				jwtToken = strings.TrimPrefix(c, cookieName+"=")
				break
			}
		}

		if jwtToken == "" {
			return handleMiddlewareResponse(http.StatusUnauthorized, "Unauthenticated")
		}

		claims, err := jwtUtils.ValidateToken(jwtToken)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			return handleMiddlewareResponse(http.StatusUnauthorized, "Invalid token")
		}

		userId, err := jwtUtils.ExtractClaim(claims, "userId")
		if err != nil {
			return handleMiddlewareResponse(http.StatusUnauthorized, "Unauthorized")
		}

		return handleMiddlewareResponse(http.StatusOK, userId)
	}
}
