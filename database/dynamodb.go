package database

import (
	"fmt"
	"log"
	"os"

	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type AWSCredentials struct {
	AccessKey string
	SecretKey string
	Region    string
}

var db *dynamodb.DynamoDB

func init() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	env := os.Getenv(utils.ENV)
	fmt.Print(env)
	if env == utils.PROD {
		log.Println(env, " is the env")
	}

}

func getAWSCredentials() *AWSCredentials {

	var found bool
	awsCredentials := new(AWSCredentials)

	awsCredentials.Region, found = os.LookupEnv(utils.REGION)
	if !found {
		log.Panic("AWS Region not found ")
	}
	awsCredentials.AccessKey, found = os.LookupEnv(utils.ACCESS_KEY)
	if !found {
		log.Panic("AWS access key not found")
	}
	awsCredentials.SecretKey, found = os.LookupEnv(utils.SECRET_KEY)
	if !found {
		log.Panic("AWS secret key not found")
	}

	return awsCredentials
}

func GetTableName(envVarName string) string {
	tableName, found := os.LookupEnv(envVarName)
	if !found {
		log.Panicf("%v is not set in env", envVarName)
	}
	return tableName
}

func CreateDynamoDB() *dynamodb.DynamoDB {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	awsCredentials := getAWSCredentials()
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
