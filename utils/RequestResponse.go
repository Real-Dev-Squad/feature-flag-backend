package utils

type UpdateFeatureFlagRequest struct {
	Status string `json:"Status" validate:"required"`
	UserId string `json:"UserId" validate:"required"`
}

type CreateFeatureFlagRequest struct {
	FlagName    string `json:"Name" validate:"required,unique"`
	Description string `json:"Description" validate:"required"`
	UserId      string `json:"UserId" validate:"required"`
}

type FeatureFlagResponse struct {
	Id          string `json:"Id"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Status      string `json:"Status"`
}
