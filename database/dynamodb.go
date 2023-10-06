package database

import (
	"errors"
	"log"
	"os"

	"github.com/Real-Dev-Squad/feature-flag-backend/models"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var db *dynamodb.DynamoDB
var marshalMapFunction = dynamodbattribute.MarshalMap
var unmarshalMapFunction = dynamodbattribute.UnmarshalMap

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

	if db == nil {
		if env == utils.DEV || env == utils.TEST {
			sess, err = session.NewSession(&aws.Config{
				Region:   aws.String(os.Getenv("AWS_REGION")),
				Endpoint: aws.String("http://host.docker.internal:8000"),
			})

			if err != nil {
				log.Printf("Error creating the dynamodb session in DEV env \n %v", err)
				utils.ServerError(errors.New("Error creating dynamodb session in DEV env"))
			}

		} else { //ENV = PROD

			sess, err = session.NewSession()

			if err != nil {
				log.Printf("Error creating the dynamodb session in PROD env \n %v", err)
				utils.ServerError(errors.New("Error creating dynamodb session in PROD env"))
			}
		}
		db = dynamodb.New(sess)
	}

	return db
}

func MarshalMap(input interface{}) (map[string]*dynamodb.AttributeValue, error) {
	item, err := marshalMapFunction(input)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func UnmarshalMap(input map[string]*dynamodb.AttributeValue, targetStruct interface{}) error {
	err := unmarshalMapFunction(input, &targetStruct)
	if err != nil {
		return err
	}
	return nil
}

func createTables(db *dynamodb.DynamoDB, schemas []dynamodb.CreateTableInput) error {
	for _, schema := range schemas {
		input := &dynamodb.CreateTableInput{
			TableName:             schema.TableName,
			KeySchema:             schema.KeySchema,
			AttributeDefinitions:  schema.AttributeDefinitions,
			ProvisionedThroughput: schema.ProvisionedThroughput,
		}

		_, err := db.CreateTable(input)
		if err != nil {
			return err
		}
	}

	return nil
}

func ProcessGetFeatureFlagByHashKey(attributeName string, attributeValue string) (*utils.FeatureFlagResponse, error) {

	db := CreateDynamoDB()

	utils.CheckRequestAllowed(db, utils.ConcurrencyDisablingLambda)

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
	err = UnmarshalMap(result.Item, &featureFlagResponse)

	if err != nil {
		log.Println(err, " is the error while converting to ddb object")
		return nil, err
	}
	return featureFlagResponse, nil
}

func AddUserFeatureFlagMapping(featureFlagUserMappings []models.FeatureFlagUserMapping) ([]models.FeatureFlagUserMapping, error) {
	db := CreateDynamoDB()

	for _, featureFlagUserMapping := range featureFlagUserMappings {
		item, err := MarshalMap(featureFlagUserMapping)
		if err != nil {
			return nil, err
		}

		input := &dynamodb.PutItemInput{
			TableName:           aws.String(utils.FEATURE_FLAG_USER_MAPPING_TABLE_NAME),
			Item:                item,
			ConditionExpression: aws.String("attribute_not_exists(userId)"),
		}

		_, err = db.PutItem(input)
		if err != nil {
			utils.DdbError(err)
			return nil, err
		}
	}
	return featureFlagUserMappings, nil
}
