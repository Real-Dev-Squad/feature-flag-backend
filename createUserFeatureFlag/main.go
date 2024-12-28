package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Real-Dev-Squad/feature-flag-backend/database"
	"github.com/Real-Dev-Squad/feature-flag-backend/jwt"
	"github.com/Real-Dev-Squad/feature-flag-backend/middleware"
	"github.com/Real-Dev-Squad/feature-flag-backend/models"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func processPutById(userId string, flagId string, featureFlagUserMapping models.FeatureFlagUserMapping) (*models.FeatureFlagUserMapping, error) {

	db := database.CreateDynamoDB()

	utils.CheckRequestAllowed(db, utils.ConcurrencyDisablingLambda)

	item, err := database.MarshalMap(featureFlagUserMapping)
	if err != nil {
		return nil, err
	}

	input := &dynamodb.PutItemInput{
		TableName:           aws.String(utils.FEATURE_FLAG_USER_MAPPING_TABLE_NAME),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(userId)"),
	}

	_, err = db.PutItem(input)
	if err != nil {
		utils.DdbError(err)
		return nil, err
	}

	return &featureFlagUserMapping, nil
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userId := req.PathParameters["userId"]
	flagId := req.PathParameters["flagId"]

	corsResponse, err := middleware.CORSMiddleware()(req)
	if err != nil {
		log.Printf("CORS error: %v", err)
		return corsResponse, err
	}

	if corsResponse.StatusCode != http.StatusOK {
		return corsResponse, nil
	}

	jwtResponse, _, err := jwt.JWTMiddleware()(req)
	if err != nil {
		log.Printf("JWT middleware error: %v", err)
		return jwtResponse, err
	}

	if jwtResponse.StatusCode != http.StatusOK {
		return jwtResponse, nil
	}

	var requestBody utils.CreateFeatureFlagUserMappingRequest
	err = json.Unmarshal([]byte(req.Body), &requestBody)
	if err != nil {
		log.Printf("Error unmarshal request body: \n %v", err)
		return utils.ClientError(http.StatusUnprocessableEntity, "Error unmarshal request body")
	}

	if err := validate.Struct(&requestBody); err != nil {
		response := events.APIGatewayProxyResponse{
			Body:       "Check the request body passed status and userId are required.",
			StatusCode: http.StatusBadRequest,
		}
		return response, nil
	}

	found := utils.ValidateFeatureFlagStatus(requestBody.Status)
	if !found {
		response := events.APIGatewayProxyResponse{
			Body:       "Allowed values of Status are ENABLED, DISABLED",
			StatusCode: http.StatusBadRequest,
		}
		return response, nil
	}

	featureFlagUserMapping := models.FeatureFlagUserMapping{
		UserId:    userId,
		FlagId:    flagId,
		Status:    strings.ToUpper(requestBody.Status),
		CreatedAt: time.Now().Unix(),
		CreatedBy: requestBody.UserId,
		UpdatedAt: time.Now().Unix(),
		UpdatedBy: requestBody.UserId,
	}

	result, err := processPutById(userId, flagId, featureFlagUserMapping)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
				return utils.ClientError(http.StatusNotFound, "Mapping of User Id and Flag Id already exists")
			}
		}
		return utils.ServerError(err)
	}
	resultJson, err := json.Marshal(result)
	if err != nil {
		log.Println("Error converting featureFlagUserMapping to JSON")
		return utils.ServerError(err)
	}

	origin := req.Headers["Origin"]
	corsHeaders := middleware.GetCORSHeaders(origin)

	response := events.APIGatewayProxyResponse{
		Body:       string(resultJson),
		StatusCode: http.StatusOK,
		Headers:    corsHeaders,
	}

	return response, nil

}

func main() {
	lambda.Start(handler)
}
