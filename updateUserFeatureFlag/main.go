package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Real-Dev-Squad/feature-flag-backend/database"
	"github.com/Real-Dev-Squad/feature-flag-backend/jwt"
	"github.com/Real-Dev-Squad/feature-flag-backend/middleware"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func processUpdateByIds(userId string, flagId string, requestBody utils.UpdateFeatureFlagUserMappingRequest) (*utils.FeatureFlagUserMappingResponse, error) {

	db := database.CreateDynamoDB()

	utils.CheckRequestAllowed(db, utils.ConcurrencyDisablingLambda)

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(utils.FEATURE_FLAG_USER_MAPPING_TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			utils.UserId: { // partition key
				S: aws.String(userId),
			},
			utils.FlagId: { // sort key
				S: aws.String(flagId),
			},
		},
		UpdateExpression: aws.String("set #status = :status, #updatedBy = :updatedBy, #updatedAt = :updatedAt"),
		ExpressionAttributeNames: map[string]*string{
			"#status":    aws.String("status"),
			"#updatedBy": aws.String("updatedBy"),
			"#updatedAt": aws.String("updatedAt"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":status": {
				S: aws.String(strings.ToUpper(requestBody.Status)),
			},
			":updatedBy": {
				S: aws.String(requestBody.UserId),
			},
			":updatedAt": {
				N: aws.String(strconv.FormatInt(time.Now().Unix(), 10)),
			},
		},
		ConditionExpression: aws.String("attribute_exists(userId)"),
		ReturnValues:        aws.String("ALL_NEW"),
	}

	result, err := db.UpdateItem(input)

	if err != nil {
		utils.DdbError(err)
		return nil, err
	}

	featureFlagUserMapping := new(utils.FeatureFlagUserMappingResponse)
	err = dynamodbattribute.UnmarshalMap(result.Attributes, &featureFlagUserMapping)

	if err != nil {
		return nil, err
	}
	return featureFlagUserMapping, nil
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

	var requestBody utils.UpdateFeatureFlagUserMappingRequest
	err = json.Unmarshal([]byte(req.Body), &requestBody)
	if err != nil {
		log.Printf("Error unmarshal request body: \n %v", err)
		return utils.ClientError(http.StatusUnprocessableEntity, "Error unmarshalling request body")
	}

	if err := validate.Struct(&requestBody); err != nil {
		errorMessage := "Check the request body passed status and userId are required."
		response := events.APIGatewayProxyResponse{
			Body:       errorMessage,
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

	result, err := processUpdateByIds(userId, flagId, requestBody)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
				return utils.ClientError(http.StatusNotFound, "Mapping of User Id and Flag Id does not exist")
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
