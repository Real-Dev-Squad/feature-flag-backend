package main

import (
	"log"
	"os"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
)


var Constants map[string]string  = map[string]string {
	"ENV" : 				"ENVIRONMENT",
	"ENV_DEVELOPMENT": 		"DEVELOPMENT",
	"ENV_PRODUCTION":		"PRODUCTION",
	"REGION": 				"AWS_REGION",
	"ACCESS_KEY":			"AWS_ACCESS_KEY",
	"SECRET_KEY":			"AWS_SECRET_KEY",
}

type AWSCredentials struct{
	AccessKey string
	SecretKey string
	Region string
}

var db *dynamodb.DynamoDB

func init(){

	env, ok := os.LookupEnv(Constants["ENV"])
	if !ok {
		// load the env values using the `setUpEnv` function.
		utils.setUpEnv()
	} 
	if (env == Constants["ENV_PRODUCTION"]){
		// to be written
	}
}

func getAWSCredentials() *AWSCredentials{

	awsCredentials := new(AWSCredentials)
	awsCredentials.Region = os.Getenv(Constants["REGION"])
	awsCredentials.AccessKey = os.Getenv(Constants["ACCES_KEY"])
	awsCredentials.SecretKey = os.Getenv(Constants["SECRET_KEY"])
	
	return awsCredentials
}

func createDynamoDB() *dynamodb.DynamoDB {
	awsCredentials := getAWSCredentials()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsCredentials.Region),
		Credentials: credentials.NewStaticCredentials(awsCredentials.AccessKey,awsCredentials.SecretKey,""),
	})

	if err != nil {
		log.Panic("Error while creating a session")
	}
	db = dynamodb.New(sess)
	return db
}
