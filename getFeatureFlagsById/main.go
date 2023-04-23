package main

import (
	"os"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"

)


type AWSCredentials struct{
	AccessKey string
	SecretKey string
	Region string
}

var db *dynamodb.DynamoDB

func getAWSCredentials() *AWSCredentials{
	err := godotenv.Load("../.env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	awsCredentials := new(AWSCredentials)
	if os.Getenv(("ENVIRONMENT")) == "DEVELOPMENT" {
		awsCredentials.Region = os.Getenv("AWS_REGION")
		awsCredentials.AccessKey = os.Getenv(("AWS_ACCESS_KEY"))
		awsCredentials.SecretKey = os.Getenv(("AWS_SECRET_KEY"))
	}
	
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

func main() { 
	
	db := createDynamoDB()
	
	// input := &dynamodb.ListTablesInput{}

	// fmt.Println("Tables: \n")

		
	// params := &dynamodb.QueryInput{
	// 	TableName: aws.String(os.Getenv("FEATURE_FLAG_TABLE_NAME")),
	// 	KeyConditionExpression: aws.String(
	// 		"Id = :pk"),
	// 	ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
	// 		":pk" :{
	// 			S: aws.String("flag1"),
	// 		},
	// 	},
	// }
	// result, err := db.Query(params)
	// if err != nil{
	// 	fmt.Println("failed to query data", err)
	// 	return
	// }

	// fmt.Println(result," is the result")
}