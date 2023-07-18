package database

import (
	"errors"
	"log"
	"os"

	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-sdk-go/aws"
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

	if env == utils.DEV {
		input := &dynamodb.ListTablesInput{}
		result, err := db.ListTables(input)
		if err != nil {
			log.Printf("Error listing tables \n %v", err)
			utils.ServerError(errors.New("Error listing tables"))
		}

		if len(result.TableNames) == 0 {
			tableSchemas := []dynamodb.CreateTableInput{
				{
					TableName: aws.String(utils.FEATURE_FLAG_USER_MAPPING_TABLE_NAME),
					KeySchema: []*dynamodb.KeySchemaElement{
						{
							AttributeName: aws.String(utils.UserId),
							KeyType:       aws.String("HASH"),
						},
						{
							AttributeName: aws.String(utils.FlagId),
							KeyType:       aws.String("RANGE"),
						},
					},
					AttributeDefinitions: []*dynamodb.AttributeDefinition{
						{
							AttributeName: aws.String(utils.UserId),
							AttributeType: aws.String("S"),
						},
						{
							AttributeName: aws.String(utils.FlagId),
							AttributeType: aws.String("S"),
						},
					},
					ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(5),
						WriteCapacityUnits: aws.Int64(5),
					},
				},
				{
					TableName: aws.String(utils.FEATURE_FLAG_TABLE_NAME),
					KeySchema: []*dynamodb.KeySchemaElement{
						{
							AttributeName: aws.String(utils.Id),
							KeyType:       aws.String("HASH"),
						},
					},
					AttributeDefinitions: []*dynamodb.AttributeDefinition{
						{
							AttributeName: aws.String(utils.Id),
							AttributeType: aws.String("S"),
						},
					},
					ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(5),
						WriteCapacityUnits: aws.Int64(5),
					},
				},
			}
			err := createTables(db, tableSchemas)
			if err != nil {
				log.Printf("Error setting up local dynamodb in DEV env \n %v", err)
				utils.ServerError(errors.New("Error setting up local dynamodb in DEV env"))
			}
		}
	}

	return db
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
