package database

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/Real-Dev-Squad/feature-flag-backend/models"
	"github.com/Real-Dev-Squad/feature-flag-backend/utils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var db *dynamodb.Client
var cfg aws.Config
var marshalMapFunction = attributevalue.MarshalMap
var unmarshalMapFunction = attributevalue.UnmarshalMap

func init() {
	env := os.Getenv(utils.ENV)
	log.Println("ENV=", env)
}

func CreateDynamoDB() *dynamodb.Client {
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

	var err error

	if db == nil {
		ctx := context.TODO()
		if env == utils.DEV || env == utils.TEST {
			cfg, err = config.LoadDefaultConfig(ctx,
				config.WithRegion(os.Getenv("AWS_REGION")),
			)

			if err != nil {
				log.Printf("Error creating the dynamodb config in DEV env \n %v", err)
				utils.ServerError(errors.New("Error creating dynamodb config in DEV env"))
			}

		} else { //ENV = PROD

			cfg, err = config.LoadDefaultConfig(ctx)

			if err != nil {
				log.Printf("Error creating the dynamodb config in PROD env \n %v", err)
				utils.ServerError(errors.New("Error creating dynamodb config in PROD env"))
			}
		}
		// Create DynamoDB client with custom endpoint for local development
		if env == utils.DEV || env == utils.TEST {
			db = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
				o.BaseEndpoint = aws.String("http://host.docker.internal:8000")
			})
		} else {
			db = dynamodb.NewFromConfig(cfg)
		}
	}

	return db
}

func MarshalMap(input interface{}) (map[string]types.AttributeValue, error) {
	item, err := marshalMapFunction(input)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func UnmarshalMap(input map[string]types.AttributeValue, targetStruct interface{}) error {
	err := unmarshalMapFunction(input, &targetStruct)
	if err != nil {
		return err
	}
	return nil
}

func createTables(db *dynamodb.Client, schemas []dynamodb.CreateTableInput) error {
	ctx := context.TODO()
	for _, schema := range schemas {
		input := &dynamodb.CreateTableInput{
			TableName:             schema.TableName,
			KeySchema:             schema.KeySchema,
			AttributeDefinitions:  schema.AttributeDefinitions,
			ProvisionedThroughput: schema.ProvisionedThroughput,
		}

		_, err := db.CreateTable(ctx, input)
		if err != nil {
			return err
		}
	}

	return nil
}

func ProcessGetFeatureFlagByHashKey(attributeName string, attributeValue string) (*utils.FeatureFlagResponse, error) {
	ctx := context.TODO()
	db := CreateDynamoDB()

	utils.CheckRequestAllowed(ctx, db, utils.ConcurrencyDisablingLambda)

	input := &dynamodb.GetItemInput{
		TableName: aws.String(utils.FEATURE_FLAG_TABLE_NAME),
		Key: map[string]types.AttributeValue{
			attributeName: &types.AttributeValueMemberS{
				Value: attributeValue,
			},
		},
	}

	result, err := db.GetItem(ctx, input)

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
	ctx := context.TODO()
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

		_, err = db.PutItem(ctx, input)
		if err != nil {
			utils.DdbError(err)
			return nil, err
		}
	}
	return featureFlagUserMappings, nil
}

