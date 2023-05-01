package database

import (
	"fmt"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
	"os"
)

type AWSCredentials struct {
	AccessKey string
	SecretKey string
	Region    string
}

type GetItemResult struct {
	Item map[string]*dynamodb.AttributeValue
}

var db *dynamodb.DynamoDB

func init() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	env := os.Getenv(utils.Constants["ENV"])
	if env == utils.Constants["PROD"] {
		log.Println(env, " is the env")
	}

}

func getAWSCredentials() *AWSCredentials {

	var found bool
	awsCredentials := new(AWSCredentials)

	awsCredentials.Region, found = os.LookupEnv(utils.Constants["REGION"])
	if !found {
		log.Panic("AWS Region not found ")
	}
	awsCredentials.AccessKey, found = os.LookupEnv(utils.Constants["ACCESS_KEY"])
	if !found {
		log.Panic("AWS access key not found")
	}
	awsCredentials.SecretKey, found = os.LookupEnv(utils.Constants["SECRET_KEY"])
	if !found {
		log.Panic("AWS secret key not found")
	}

	return awsCredentials
}

func GetFeatureFlagTableName() string {
	tableName, found := os.LookupEnv(utils.Constants["FF_TABLE_NAME"])
	if !found {
		log.Panic("Feature Flag table name is not set in env.")
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
