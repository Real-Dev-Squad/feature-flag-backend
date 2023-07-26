package main

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Real-Dev-Squad/feature-flag-backend/database"
	"github.com/Real-Dev-Squad/feature-flag-backend/models"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	os.Setenv(utils.ENV, utils.TEST)
	os.Setenv("AWS_REGION", "fakeRegion")
	defer os.Clearenv()

	createdByUserId := uuid.New().String()
	userId := uuid.New().String()
	flagId := uuid.New().String()

	featureFlagUserMappings := []models.FeatureFlagUserMapping{
		{
			UserId:    userId,
			FlagId:    flagId,
			Status:    utils.ENABLED,
			CreatedBy: createdByUserId,
			UpdatedBy: createdByUserId,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		},
	}

	result, err := database.AddUserFeatureFlagMapping(featureFlagUserMappings)
	if err != nil {
		t.Fatalf("Failed to add user feature flag mapping: %v", err)
	}
	resultJson, err := json.Marshal(result[0])
	if err != nil {
		t.Fatalf("Error converting addUserFeatureFlagMapping result to JSON")
	}

	tests := []struct {
		request      events.APIGatewayProxyRequest
		responseBody string
		statusCode   int
		err          error
	}{
		{
			request:      events.APIGatewayProxyRequest{PathParameters: map[string]string{"userId": uuid.New().String(), "flagId": uuid.New().String()}},
			responseBody: "User feature flag not found",
			statusCode:   http.StatusNotFound,
			err:          nil,
		},
		{
			request:      events.APIGatewayProxyRequest{PathParameters: map[string]string{"userId": userId, "flagId": flagId}},
			responseBody: string(resultJson),
			statusCode:   http.StatusOK,
			err:          nil,
		},
	}

	for _, test := range tests {
		response, err := handler(test.request)

		assert.IsType(t, test.err, err)
		assert.Equal(t, test.statusCode, response.StatusCode)
		assert.Equal(t, test.responseBody, response.Body)
	}
}
