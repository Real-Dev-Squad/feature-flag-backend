package models

type FeatureFlag struct {
	Id          string `json:"id" dynamodbav:"Id"`
	Name        string `json:"name" dynamodbav:"name"`
	Description string `json:"description" dynamodbav:"description"`
	CreatedAt   int64  `json:"createdAt" dynamodbav:"createdAt"` //was getting error on unmarshal ddb result with time.Time
	CreatedBy   string `json:"createdBy" dynamodbav:"createdBy"`
	UpdatedAt   int64  `json:"updatedAt" dynamodbav:"updatedAt"`
	UpdatedBy   string `json:"updatedBy" dynamodbav:"updatedBy"`
	Status      string `json:"status" dynamodbav:"status"`
}

type UpdateFeatureFlagRequest struct {
	Id          string `json:"Id" validate:"required"`
	Status      string `json:"Status" validate:"required"`
	UserId      string `json:"UserId" validate:"required"`
}

type CreateFeatureFlagRequest struct {
	FlagName    string `json:"Name" validate:"required"`
	Description string `json:"Description" validate:"required"`
	UserId      string `json:"UserId" validate:"required"`
}

type FeatureFlagResponse struct{
	Id string `json:"Id"`
	Name string `json:"Name"`
	Description string `json:"Description"`
	Status string `json:"Status"`
}