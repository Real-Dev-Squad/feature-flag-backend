package utils

type UpdateFeatureFlagRequest struct {
	Status string `json:"status" validate:"required"`
	UserId string `json:"userId" validate:"required"`
}

type CreateFeatureFlagRequest struct {
	FlagName    string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
	UserId      string `json:"userId" validate:"required"`
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
	UserId string `json:"userId" validate:"required"`
}

type UpdateFeatureFlagUserMappingRequest struct {
	Status string `json:"status" validate:"required"`
	UserId string `json:"userId" validate:"required"`
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
