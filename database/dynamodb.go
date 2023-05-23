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
	"errors"
	"os"
)

type AWSCredentials struct {
	AccessKey string
	SecretKey string
	Region    string
}

var db *dynamodb.DynamoDB

func init() {

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
		log.Println("AWS region not stored, please store it.")

		// Here the error is passed for debugging to the user we show "Something went wrong"
		utils.ServerError(errors.New("AWS region not stored in ENV var")) 
	}

	awsCredentials.AccessKey, found = os.LookupEnv(utils.ACCESS_KEY)
	if !found {
		log.Println("AWS Access key not stored, please store it.")
		
		// Here the error is passed for debugging to the user we show "Something went wrong"
		utils.ServerError(errors.New("AWS Access key not stored in ENV var"))
	}

	awsCredentials.SecretKey, found = os.LookupEnv(utils.SECRET_KEY)
	if !found {
		log.Println("AWS Secret key not stored, please store it.")

		// Here the error is passed for debugging to the user we show "Something went wrong
		utils.ServerError(errors.New("AWS Secret key not stored, please store it."))
	}

	return awsCredentials
}

func GetFeatureFlagTableName() string {
	tableName, found := os.LookupEnv(utils.FF_TABLE_NAME)
	if !found {
		log.Println("Feature Flag table name is not set in env.")

		// Here the error is passed for debugging to the user we show "Something went wrong
		utils.ServerError(errors.New("Feature flag table name not stored in the env."))
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
		log.Println("Error creating the dynamodb session.")
		utils.ServerError(errors.New("Error creating dynamodb session"))
	}
	db = dynamodb.New(sess)
	return db
}

func ProcessGetFeatureFlagByHashKey(attributeName string, attributeValue string) (*models.FeatureFlagResponse, error) {

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

	featureFlagResponse := new(models.FeatureFlagResponse)
	err = dynamodbattribute.UnmarshalMap(result.Item, &featureFlagResponse)

	if err != nil {
		log.Println(err ," is the error while converting to ddb object")
		return nil, err
	}
	return featureFlagResponse, nil
}