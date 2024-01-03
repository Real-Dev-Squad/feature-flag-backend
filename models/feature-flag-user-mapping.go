package models

type FeatureFlagUserMapping struct {
	UserId    string `json:"userId" dynamodbav:"userId"`
	FlagId    string `json:"flagId" dynamodbav:"flagId"`
	Status    string `json:"status" dynamodbav:"status"`
	CreatedAt int64  `json:"createdAt" dynamodbav:"createdAt"`
	CreatedBy string `json:"createdBy" dynamodbav:"createdBy"`
	UpdatedAt int64  `json:"updatedAt" dynamodbav:"updatedAt"`
	UpdatedBy string `json:"updatedBy" dynamodbav:"updatedBy"`
}
