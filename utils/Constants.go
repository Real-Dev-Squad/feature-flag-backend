package utils

const (
	ENV                                  = "ENVIRONMENT"
	DEV                                  = "DEVELOPMENT"
	PROD                                 = "PRODUCTION"
	TEST                                 = "TESTING"
	FEATURE_FLAG_TABLE_NAME              = "featureFlag"
	FEATURE_FLAG_USER_MAPPING_TABLE_NAME = "featureFlagUserMapping"
	ENABLED                              = "ENABLED"
	DISABLED                             = "DISABLED"
	Id                                   = "id"
	Name                                 = "name"
	Description                          = "description"
	CreatedAt                            = "createdAt"
	CreatedBy                            = "createdBy"
	Status                               = "status"
	UpdatedBy                            = "updatedBy"
	UpdatedAt                            = "updatedAt"
	UserId                               = "userId"
	FlagId                               = "flagId"
	ConcurrencyDisablingLambda           = 0
	SESSION_COOKIE_NAME_PROD             = "rds-session"
	SESSION_COOKIE_NAME_DEV              = "rds-session-staging"
	SESSION_COOKIE_NAME_LOCAL            = "rds-session-development"
	RDS_BACKEND_PUBLIC_KEY_NAME_DEV      = "rdsbackendpublickeystaging"
	RDS_BACKEND_PUBLIC_KEY_NAME_PROD     = "rdsbackendpublickeyprod"
	RDS_BACKEND_PUBLIC_KEY_NAME_LOCAL    = "publickey"
)
