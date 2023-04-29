package main

import (
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
	"os"
)

var Constants map[string]string = map[string]string{
	"ENV":        "ENVIRONMENT",
	"DEV":        "DEVELOPMENT",
	"PROD":       "PRODUCTION",
	"REGION":     "AWS_REGION",
	"ACCESS_KEY": "AWS_ACCESS_KEY",
	"SECRET_KEY": "AWS_SECRET_KEY",
}

type AWSCredentials struct {
	AccessKey string
	SecretKey string
	Region    string
}

var db *dynamodb.DynamoDB

func init() {

	env, found := os.LookupEnv(Constants["ENV"])
	if !found {
		// load the env values using the `setUpEnv` function.
		utils.SetUpEnv()
	}

}

func getAWSCredentials() *AWSCredentials {

	awsCredentials := new(AWSCredentials)
	awsCredentials.Region = os.Getenv(Constants["REGION"])
	awsCredentials.AccessKey = os.Getenv(Constants["ACCES_KEY"])
	awsCredentials.SecretKey = os.Getenv(Constants["SECRET_KEY"])

	return awsCredentials
}

func createDynamoDB() *dynamodb.DynamoDB {
	awsCredentials := getAWSCredentials()

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsCredentials.Region),
		Credentials: credentials.NewStaticCredentials(awsCredentials.AccessKey, awsCredentials.SecretKey, ""),
	})

	if err != nil {
		log.Panic("Error while creating a session")
	}
	db = dynamodb.New(sess)
	return db
}
