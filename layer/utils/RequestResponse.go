package utils

type UpdateFeatureFlagRequest struct {
	Status string `json:"status" validate:"required"`
	UserId string `json:"userId"` // Optional - will be set from authenticated user context
}

type CreateFeatureFlagRequest struct {
	FlagName    string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
	UserId      string `json:"userId"` // Optional - will be set from authenticated user context
}

type FeatureFlagResponse struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   int64  `json:"createdAt"`
	CreatedBy   string `json:"createdBy"`
	UpdatedAt   int64  `json:"updatedAt"`
	UpdatedBy   string `json:"updatedBy"`
}

type CreateFeatureFlagUserMappingRequest struct {
	Status string `json:"status" validate:"required"`
	UserId string `json:"userId"` // Optional - will be set from authenticated user context
}

type UpdateFeatureFlagUserMappingRequest struct {
	Status string `json:"status" validate:"required"`
	UserId string `json:"userId"` // Optional - will be set from authenticated user context
}

type FeatureFlagUserMappingResponse struct {
	UserId    string `json:"userId"`
	FlagId    string `json:"flagId"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"createdAt"`
	CreatedBy string `json:"createdBy"`
	UpdatedAt int64  `json:"updatedAt"`
	UpdatedBy string `json:"updatedBy"`
}

type RegisterUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role" validate:"omitempty,oneof=ADMIN DEVELOPER VIEWER"`
}

type LoginUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	Id        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsActive  bool   `json:"isActive"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  UserResponse `json:"user"`
}
