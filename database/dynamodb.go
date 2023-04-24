package main

import (
	"log"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"

)


type AWSCredentials struct{
	AccessKey string
	SecretKey string
	Region string
}

var db *dynamodb.DynamoDB

func getAWSCredentials() *AWSCredentials{

	awsCredentials := new(AWSCredentials)
	awsCredentials.Region = utils.GoDotEnvVariable("AWS_REGION")
	awsCredentials.AccessKey = utils.GoDotEnvVariable(("AWS_ACCESS_KEY"))
	awsCredentials.SecretKey = utils.GoDotEnvVariable(("AWS_SECRET_KEY"))
	
	return awsCredentials
}

func createDynamoDB() *dynamodb.DynamoDB {
	awsCredentials := getAWSCredentials()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsCredentials.Region),
		Credentials: credentials.NewStaticCredentials(awsCredentials.AccessKey,awsCredentials.SecretKey,""),
	})

	if err != nil {
		log.Fatal("Error while creating a session")
	}
	db = dynamodb.New(sess)
	return db
}
