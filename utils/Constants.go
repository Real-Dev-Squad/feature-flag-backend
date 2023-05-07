package utils

var Constants map[string]string = map[string]string{
	"ENV":           "ENVIRONMENT",
	"DEV":           "DEVELOPMENT",
	"PROD":          "PRODUCTION",
	"REGION":        "AWS_REGION",
	"ACCESS_KEY":    "AWS_ACCESS_KEY",
	"SECRET_KEY":    "AWS_SECRET_KEY",
	"FF_TABLE_NAME": "FEATURE_FLAG_TABLE_NAME",
}

var FeatureFlagStatus map[string]string = map[string]string{
	"ENABLED":  "ENABLED",
	"DISABLED": "DISABLED",
}

var FeatureFlagAttributes map[string]string = map[string]string{
	"Id":          "Id",
	"name":        "name",
	"description": "description",
	"createdAt":   "createdAt",
	"createdBy":   "createdBy",
	"updatedAt":   "updatedAt",
	"updatedBy":   "updatedBy",
	"status":      "status",
}
