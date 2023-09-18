package models

type FeatureFlag struct {
	Id          string `json:"id" dynamodbav:"id"`
	Name        string `json:"name" dynamodbav:"name"`
	Description string `json:"description" dynamodbav:"description"`
	CreatedAt   int64  `json:"createdAt" dynamodbav:"createdAt"` //was getting error on unmarshal ddb result with time.Time
	CreatedBy   string `json:"createdBy" dynamodbav:"createdBy"`
	UpdatedAt   int64  `json:"updatedAt" dynamodbav:"updatedAt"`
	UpdatedBy   string `json:"updatedBy" dynamodbav:"updatedBy"`
	Status      string `json:"status" dynamodbav:"status"`
}
