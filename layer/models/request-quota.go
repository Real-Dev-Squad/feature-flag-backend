package models

type RequestLimit struct {
	LimitType string `json:"limitType" dynamodbav:"limitType"`
	LimitValue int16 `json:"limitValue" dynamodbav:"limitValue"`
}