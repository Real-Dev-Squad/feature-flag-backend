package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Real-Dev-Squad/feature-flag-backend/database"
	"github.com/Real-Dev-Squad/feature-flag-backend/jwt"
	middleware "github.com/Real-Dev-Squad/feature-flag-backend/middlewares"
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

func updateFeatureFlag(flagId string, updateFeatureFlagRequest utils.UpdateFeatureFlagRequest) (events.APIGatewayProxyResponse, error) {
	db := database.CreateDynamoDB()

	utils.CheckRequestAllowed(db, utils.ConcurrencyDisablingLambda)

	input := &dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(flagId),
			},
		},
		TableName: aws.String(utils.FEATURE_FLAG_TABLE_NAME),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":status": {
				S: aws.String(updateFeatureFlagRequest.Status),
			},
			":updatedAt": {
				N: aws.String(strconv.Itoa(int(time.Now().Unix()))),
			},
			":updatedBy": {
				S: aws.String(updateFeatureFlagRequest.UserId),
			},
		},
		UpdateExpression: aws.String("set #status = :status, #updatedAt = :updatedAt, #updatedBy = :updatedBy"),
		ExpressionAttributeNames: map[string]*string{
			"#status":    aws.String("status"),
			"#updatedAt": aws.String("updatedAt"),
			"#updatedBy": aws.String("updatedBy"),
		},
		ReturnValues:        aws.String("ALL_NEW"),
		ConditionExpression: aws.String("attribute_exists(Id)"),
	}

	result, err := db.UpdateItem(input)

	//throw the response on conditional check failed exception
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
				return utils.ClientError(http.StatusNotFound, "Feature flag with given flagId doesn't exists")
			}
		}
		utils.ServerError(err)
	}

	featureFlag := new(utils.FeatureFlagResponse)
	err = database.UnmarshalMap(result.Attributes, &featureFlag)

	if err != nil {
		log.Printf("Error is %v", err)
		utils.ServerError(err)
	}

	//marshal to JSON
	resultJson, err := json.Marshal(featureFlag)
	if err != nil {
		log.Printf("Unable to marshal to JSON \n %v", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(resultJson),
	}, nil
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id, _ := request.PathParameters["flagId"]

	corsResponse, err, passed := middleware.HandleCORS(request)
	if !passed {
		return corsResponse, err
	}

	jwtResponse, _, err := jwt.JWTMiddleware()(request)
	if err != nil || jwtResponse.StatusCode != http.StatusOK {
		return jwtResponse, err
	}

	updateFeatureFlagRequest := utils.UpdateFeatureFlagRequest{}

	//marshal to updateFeatureFlag
	bytes := []byte(request.Body)
	err = json.Unmarshal(bytes, &updateFeatureFlagRequest)
	if err != nil {
		log.Printf("Error in reading input \n %v", err)
		return utils.ClientError(http.StatusBadRequest, "Error in reading input")
	}

	if err := validate.Struct(&updateFeatureFlagRequest); err != nil {
		errorMessage := "Check the request body passed status and userId are required."
		response := events.APIGatewayProxyResponse{
			Body:       errorMessage,
			StatusCode: http.StatusBadRequest,
		}
		return response, nil
	}

	origin := request.Headers["Origin"]
	corsHeaders := middleware.GetCORSHeaders(origin)

	found := utils.ValidateFeatureFlagStatus(updateFeatureFlagRequest.Status)
	if !found {
		response := events.APIGatewayProxyResponse{
			Body:       "Allowed values of Status are ENABLED, DISABLED",
			StatusCode: http.StatusBadRequest,
			Headers:    corsHeaders,
		}
		return response, nil
	}

	response, err := updateFeatureFlag(id, updateFeatureFlagRequest)
	if err != nil {
		return response, err
	}
	response.Headers = corsHeaders

	return response, nil
}

func main() {
	lambda.Start(handler)
}
