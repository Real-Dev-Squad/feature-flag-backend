package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"feature-flag-backend/layer/database"
	"feature-flag-backend/layer/jwt"
	middleware "feature-flag-backend/layer/middlewares"
	"feature-flag-backend/layer/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func processUpdateByIds(ctx context.Context, userId string, flagId string, requestBody utils.UpdateFeatureFlagUserMappingRequest) (*utils.FeatureFlagUserMappingResponse, error) {

	db := database.CreateDynamoDB()

	utils.CheckRequestAllowed(ctx, db, utils.ConcurrencyDisablingLambda)

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(utils.FEATURE_FLAG_USER_MAPPING_TABLE_NAME),
		Key: map[string]types.AttributeValue{
			utils.UserId: &types.AttributeValueMemberS{ // partition key
				Value: userId,
			},
			utils.FlagId: &types.AttributeValueMemberS{ // sort key
				Value: flagId,
			},
		},
		UpdateExpression: aws.String("set #status = :status, #updatedBy = :updatedBy, #updatedAt = :updatedAt"),
		ExpressionAttributeNames: map[string]string{
			"#status":    "status",
			"#updatedBy": "updatedBy",
			"#updatedAt": "updatedAt",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{
				Value: strings.ToUpper(requestBody.Status),
			},
			":updatedBy": &types.AttributeValueMemberS{
				Value: requestBody.UserId,
			},
			":updatedAt": &types.AttributeValueMemberN{
				Value: strconv.FormatInt(time.Now().Unix(), 10),
			},
		},
		ConditionExpression: aws.String("attribute_exists(userId)"),
		ReturnValues:        types.ReturnValueAllNew,
	}

	result, err := db.UpdateItem(ctx, input)

	if err != nil {
		utils.DdbError(err)
		return nil, err
	}

	featureFlagUserMapping := new(utils.FeatureFlagUserMappingResponse)
	err = attributevalue.UnmarshalMap(result.Attributes, featureFlagUserMapping)

	if err != nil {
		return nil, err
	}
	return featureFlagUserMapping, nil
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userId := req.PathParameters["userId"]
	flagId := req.PathParameters["flagId"]

	corsResponse, err, passed := middleware.HandleCORS(req)
	if !passed {
		return corsResponse, err
	}

	jwtResponse, _, err := jwt.JWTMiddleware()(req)
	if err != nil || jwtResponse.StatusCode != http.StatusOK {
		return jwtResponse, err
	}

	corsHeaders := middleware.GetCORSHeadersV1(req.Headers)

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

	result, err := processUpdateByIds(ctx, userId, flagId, requestBody)
	if err != nil {
		var conditionalCheckErr *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckErr) {
			return utils.ClientError(http.StatusNotFound, "Mapping of User Id and Flag Id does not exist")
		}
		return utils.ServerError(err)
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		log.Println("Error converting featureFlagUserMapping to JSON")
		return utils.ServerError(err)
	}

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
