package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"feature-flag-backend/layer/database"
	"feature-flag-backend/layer/jwt"
	middleware "feature-flag-backend/layer/middlewares"
	"feature-flag-backend/layer/models"
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

func processPutById(ctx context.Context, userId string, flagId string, featureFlagUserMapping models.FeatureFlagUserMapping) (*models.FeatureFlagUserMapping, error) {

	db := database.CreateDynamoDB()

	utils.CheckRequestAllowed(ctx, db, utils.ConcurrencyDisablingLambda)

	item, err := database.MarshalMap(featureFlagUserMapping)
	if err != nil {
		return nil, err
	}

	input := &dynamodb.PutItemInput{
		TableName:           aws.String(utils.FEATURE_FLAG_USER_MAPPING_TABLE_NAME),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(userId)"),
	}

	_, err = db.PutItem(ctx, input)
	if err != nil {
		utils.DdbError(err)
		return nil, err
	}

	return &featureFlagUserMapping, nil
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userId := req.PathParameters["userId"]
	flagId := req.PathParameters["flagId"]

	corsResponse, err, passed := middleware.HandleCORS(req)
	if !passed {
		return corsResponse, err
	}

	// Use enhanced middleware with user verification and RBAC (Week 3)
	jwtResponse, userContext, err := jwt.JWTMiddlewareWithUserVerification()(req)
	if err != nil || jwtResponse.StatusCode != http.StatusOK {
		return jwtResponse, err
	}

	if userContext == nil {
		return utils.ClientError(http.StatusUnauthorized, "User context not available")
	}

	// Check permission: CREATE_USER_MAPPING (Week 3 RBAC)
	permResponse, err := utils.RequirePermission(userContext, utils.PermissionCreateUserMapping)
	if err != nil || permResponse.StatusCode != http.StatusOK {
		permResponse.Headers = middleware.GetCORSHeadersV1(req.Headers)
		return permResponse, err
	}

	corsHeaders := middleware.GetCORSHeadersV1(req.Headers)

	// Check if user can access this resource (own resources or ADMIN)
	if !utils.CanAccessUserResource(userContext, userId) {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusForbidden,
			Body:       "You can only manage your own feature flag mappings",
			Headers:    corsHeaders,
		}, nil
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

	result, err := processPutById(ctx, userId, flagId, featureFlagUserMapping)
	if err != nil {
		var conditionalCheckErr *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckErr) {
			return utils.ClientError(http.StatusNotFound, "Mapping of User Id and Flag Id already exists")
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
