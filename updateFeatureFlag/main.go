package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func updateFeatureFlag(ctx context.Context, flagId string, updateFeatureFlagRequest utils.UpdateFeatureFlagRequest) (events.APIGatewayProxyResponse, error) {
	db := database.CreateDynamoDB()

	utils.CheckRequestAllowed(ctx, db, utils.ConcurrencyDisablingLambda)

	input := &dynamodb.UpdateItemInput{
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{
				Value: flagId,
			},
		},
		TableName: aws.String(utils.FEATURE_FLAG_TABLE_NAME),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{
				Value: updateFeatureFlagRequest.Status,
			},
			":updatedAt": &types.AttributeValueMemberN{
				Value: strconv.Itoa(int(time.Now().Unix())),
			},
			":updatedBy": &types.AttributeValueMemberS{
				Value: updateFeatureFlagRequest.UserId,
			},
		},
		UpdateExpression: aws.String("set #status = :status, #updatedAt = :updatedAt, #updatedBy = :updatedBy"),
		ExpressionAttributeNames: map[string]string{
			"#status":    "status",
			"#updatedAt": "updatedAt",
			"#updatedBy": "updatedBy",
		},
		ReturnValues:        types.ReturnValueAllNew,
		ConditionExpression: aws.String("attribute_exists(id)"),
	}

	result, err := db.UpdateItem(ctx, input)

	//throw the response on conditional check failed exception
	if err != nil {
		var conditionalCheckErr *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckErr) {
			return utils.ClientError(http.StatusNotFound, "Feature flag with given flagId doesn't exists")
		}
		utils.ServerError(err)
	}

	featureFlag := new(utils.FeatureFlagResponse)
	err = database.UnmarshalMap(result.Attributes, featureFlag)

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

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id, _ := request.PathParameters["flagId"]

	corsResponse, err, passed := middleware.HandleCORS(request)
	if !passed {
		return corsResponse, err
	}

	// Use enhanced middleware with user verification and RBAC (Week 3)
	jwtResponse, userContext, err := jwt.JWTMiddlewareWithUserVerification()(request)
	if err != nil || jwtResponse.StatusCode != http.StatusOK {
		return jwtResponse, err
	}

	if userContext == nil {
		return utils.ClientError(http.StatusUnauthorized, "User context not available")
	}

	// Check permission: UPDATE_FEATURE_FLAG (Week 3 RBAC)
	permResponse, err := utils.RequirePermission(userContext, utils.PermissionUpdateFeatureFlag)
	if err != nil || permResponse.StatusCode != http.StatusOK {
		permResponse.Headers = middleware.GetCORSHeadersV1(request.Headers)
		return permResponse, err
	}

	corsHeaders := middleware.GetCORSHeadersV1(request.Headers)

	updateFeatureFlagRequest := utils.UpdateFeatureFlagRequest{}

	//marshal to updateFeatureFlag
	bytes := []byte(request.Body)
	err = json.Unmarshal(bytes, &updateFeatureFlagRequest)
	if err != nil {
		log.Printf("Error in reading input \n %v", err)
		return utils.ClientError(http.StatusBadRequest, "Error in reading input")
	}

	if err := validate.Struct(&updateFeatureFlagRequest); err != nil {
		errorMessage := "Check the request body passed status is required."
		response := events.APIGatewayProxyResponse{
			Body:       errorMessage,
			StatusCode: http.StatusBadRequest,
			Headers:    corsHeaders,
		}
		return response, nil
	}

	// Use userId from authenticated user context (Week 2 migration)
	updateFeatureFlagRequest.UserId = userContext.UserId

	found := utils.ValidateFeatureFlagStatus(updateFeatureFlagRequest.Status)
	if !found {
		response := events.APIGatewayProxyResponse{
			Body:       "Allowed values of Status are ENABLED, DISABLED",
			StatusCode: http.StatusBadRequest,
			Headers:    corsHeaders,
		}
		return response, nil
	}

	response, err := updateFeatureFlag(ctx, id, updateFeatureFlagRequest)
	if err != nil {
		return response, err
	}
	response.Headers = corsHeaders

	return response, nil
}

func main() {
	lambda.Start(handler)
}
