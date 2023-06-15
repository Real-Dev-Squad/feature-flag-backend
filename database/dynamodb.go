package database

import (
	"errors"
	"log"
	"os"

	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var db *dynamodb.DynamoDB

func init() {
	env := os.Getenv(utils.ENV)
	log.Println("ENV=", env)
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
		sess, err = session.NewSession(&aws.Config{
			Region:      aws.String(os.Getenv("AWS_REGION")),
			Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY"), os.Getenv("AWS_SECRET_KEY"), ""),
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
		TableName: aws.String(utils.FEATURE_FLAG_TABLE_NAME),
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
