package models

type User struct {
	Id          string `json:"id" dynamodbav:"id"`
	Email       string `json:"email" dynamodbav:"email"`
	PasswordHash string `json:"passwordHash" dynamodbav:"passwordHash"`
	Role        string `json:"role" dynamodbav:"role"` // ADMIN, DEVELOPER, VIEWER
	CreatedAt   int64  `json:"createdAt" dynamodbav:"createdAt"`
	CreatedBy   string `json:"createdBy" dynamodbav:"createdBy"`
	UpdatedAt   int64  `json:"updatedAt" dynamodbav:"updatedAt"`
	UpdatedBy   string `json:"updatedBy" dynamodbav:"updatedBy"`
	IsActive    bool   `json:"isActive" dynamodbav:"isActive"`
}

