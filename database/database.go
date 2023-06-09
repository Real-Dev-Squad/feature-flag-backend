package database

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type AWSCredentials struct {
	AccessKey string
	SecretKey string
	Region    string
}

var db *dynamodb.DynamoDB

func init() {
	// os.Setenv(utils.ENV, utils.PROD)
	env := os.Getenv(utils.ENV)
	if env == utils.PROD {
		log.Println(env, " is the env")

	} else {
		// env is DEVELOPMENT
		utils.SetUpEnv() // to load the .env file
	}
}

func getAWSCredentials() *AWSCredentials {

	var found bool
	awsCredentials := new(AWSCredentials)

	awsCredentials.Region, found = os.LookupEnv(utils.REGION)
	if !found {
		log.Println("AWS region not stored, please store it.")

		utils.ServerError(errors.New("AWS region not stored in ENV var"))
	}

	awsCredentials.AccessKey, found = os.LookupEnv(utils.ACCESS_KEY)
	if !found {
		log.Println("AWS Access key not stored, please store it.")

		utils.ServerError(errors.New("AWS Access key not stored in ENV var"))
	}

	awsCredentials.SecretKey, found = os.LookupEnv(utils.SECRET_KEY)
	if !found {
		log.Println("AWS Secret key not stored, please store it.")

		utils.ServerError(errors.New("AWS Secret key not stored, please store it."))
	}

	return awsCredentials
}

func GetTableName(envVarName string) string {
	tableName, found := os.LookupEnv(envVarName)
	if !found {
		errorMessage := fmt.Sprintf("%v is not set in env. \n", envVarName)

		log.Printf(errorMessage)

		utils.ServerError(errors.New(errorMessage))
	}
	return tableName
}

func CreateDynamoDB() *dynamodb.DynamoDB {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Error is \n %v", err)
		}
	}()

	env, found := os.LookupEnv(utils.ENV)
	if !found {
		log.Println("ENV is not set, please store it.")

		utils.ServerError(errors.New("Env is not set, please set DEVELOPMENT or PRODUCTION"))
	}

	var sess *session.Session
	var err error

	if env == utils.DEV {
		awsCredentials := getAWSCredentials()

		sess, err = session.NewSession(&aws.Config{
			Region:      aws.String(awsCredentials.Region),
			Credentials: credentials.NewStaticCredentials(awsCredentials.AccessKey, awsCredentials.SecretKey, ""),
		})

		if err != nil {
			log.Printf("Error creating the dynamodb session in DEV env \n %v", err)
			utils.ServerError(errors.New("Error creating dynamodb session in PROD env"))
		}

	} else { //ENV = PROD

		sess, err = session.NewSession()

		if err != nil {
			log.Printf("Error creating the dynamodb session in PROD env \n %v", err)
			utils.ServerError(errors.New("Error creating dynamodb session in PROD env"))
		}

	}
	db = dynamodb.New(sess)
	return db
}

func ProcessGetFeatureFlagByHashKey(attributeName string, attributeValue string) (*utils.FeatureFlagResponse, error) {

	db := CreateDynamoDB()

	input := &dynamodb.GetItemInput{
		TableName: aws.String(GetTableName(utils.FF_TABLE_NAME)),
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

	if len(result.Item) == 0 {
		return nil, nil
	}

	featureFlagResponse := new(utils.FeatureFlagResponse)
	err = dynamodbattribute.UnmarshalMap(result.Item, &featureFlagResponse)

	if err != nil {
		log.Println(err, " is the error while converting to ddb object")
		return nil, err
	}
	return featureFlagResponse, nil
}
