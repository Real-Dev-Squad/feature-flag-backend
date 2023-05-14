package database

import (
	"fmt"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/Real-Dev-Squad/feature-flag-backend/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
	"os"
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

func GetFeatureFlagTableName() string {
	tableName, found := os.LookupEnv(utils.FF_TABLE_NAME)
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

func ProcessGetFeatureFlagByHashKey(attributeName string, attributeValue string) (*models.FeatureFlag, error) {

	db := CreateDynamoDB()

	input := &dynamodb.GetItemInput{
		TableName: aws.String(GetFeatureFlagTableName()),
		Key: map[string]*dynamodb.AttributeValue{
			attributeName: { 
				S: aws.String(attributeValue),
			},
		},
	}

	result, err := db.GetItem(input)

	if err != nil {
		utils.DdbError(err)
		return nil, err
	}

	if len(result.Item) == 0{
		return nil, nil
	}

	featureFlag := new(models.FeatureFlag)
	err = dynamodbattribute.UnmarshalMap(result.Item, &featureFlag)

	if err != nil {
		log.Println(err ," is the error while converting to ddb object")
		return nil, err
	}
	return featureFlag, nil
}