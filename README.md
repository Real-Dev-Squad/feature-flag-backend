# Feature Flag Backend

The Feature Flag Backend service helps manage feature flags for your applications in Real Dev Squad. With feature flags, we can easily enable or disable specific features for different groups of users. It provides APIs for creating, updating, and retrieving feature flags and user mappings. The backend is built using Golang and is deployed using a serverless architecture on [AWS Lambda](https://aws.amazon.com/lambda/). It uses DynamoDB as the database. Whenever we want to roll out new features gradually or experiment with different options, the Feature Flag Backend simplifies the process and gives us full control over our features.

## Table of Contents

-   [Installation](#installation)
-   [Run](#run)
-   [Usage](#usage)
-   [Features](#features)
-   [API Endpoints](#api-endpoints)
-   [Data Model](#data-model)
-   [Contributing](#contributing)

## Installation
You should have some things pre-installed :
- [VS Code](https://code.visualstudio.com/) or any other IDE
- [Git](https://git-scm.com/)
- [Golang](https://go.dev/)(version 1.20 or later)
- [Docker](https://www.docker.com/)


1. **Clone the repository**

    - **Open the terminal or command prompt:** Depending on your operating system, open the terminal or command prompt to begin the cloning process. 
    - **Navigate to your desired local directory:** Use the cd command to navigate to the directory where you want to store the cloned repository.
    - **Clone the repository:** Use the following command to clone the repository :
        ```
        git clone https://github.com/Real-Dev-Squad/feature-flag-backend.git
        ```

2. **AWS CLI**
    - Follow all the steps mentioned in [AWS SAM prerequisites](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/prerequisites.html) document guide to install and setup AWS CLI
    
    > **Note**
    > This step will not be required once the support for local DynamoDB setup is added. To know more read [this](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.html)
  
3. **AWS SAM CLI**
    - Follow all the steps mentioned for your local OS in [Installing the AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/prerequisites.html) document guide to install and setup AWS SAM CLI

4. **Add tables in DynamoDB**
    - Follow steps 1 to 5 (ignore the last step to add backup) mentioned in [Create a table](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/getting-started-step-1.html) development guide under AWS Management Console section
   - For table names and thier respective keys information refer to the [Data Model](#data-model) section
  
<details>
    <Summary>
        Code snippet for creating tables in local Dynamodb  - This below code needs to be pasted below this line.
https://github.com/Real-Dev-Squad/feature-flag-backend/blob/1ab29298fc371daffb747752d88f7f23fffe218c/database/dynamodb.go#L61
    </Summary>

```
    if env == utils.DEV || env == utils.TEST {
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
				log.Printf("Error setting up local dynamodb in %v env \n %v", env, err)
				utils.ServerError(errors.New("Error setting up local dynamodb in DEV env"))
			}
		}
	}
```
</details>

## Run

1. Navigate to the directory where [template.yaml](./template.yaml) is present
2. Run this command to build the backend
  ```
  sam build
  ```
3. Run this command to start the backend in development mode
  ```
  sam local start-api
  ```
  By default port 3000 is used, if you need to change add the `--port {port_number}` options at the end of the above command.
  If you have multiple AWS profiles use the `--profile {profile_name}` option at the end of the above command.
  For more options refer [here](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/sam-cli-command-reference-sam-local-start-api.html) 

## Usage

The Feature Flag Backend is used to manage feature flags for enabling or disabling features for a set of users. It provides API endpoints for creating, updating, and retrieving feature flags and user mappings.

## Features

The Feature Flag Backend project is built using the following technologies and programming languages:

-   Programming Language: Golang
-   Database: DynamoDB
-   Deployment: Serverless (AWS Lambda)
-   Deployment Automation: GitHub Actions

## API Endpoints

The API endpoints available in the Feature Flag Backend project are as follows:

-   GET `/feature-flags` to get all the feature flags
-   POST `/feature-flags` to create a feature flag
-   GET `/feature-flags/{flagId}` to get the feature flag with an ID
-   PATCH `/feature-flags/{flagId}` to update a feature flag
-   GET `/users/{userId}/feature-flags/{flagId}` to get a feature flag details for a user
-   GET `/users/{userId}/feature-flags/` to get all feature flag details for a user
-   POST `/users/{userId}/feature-flags/{flagId}` to create a feature flag for a user
-   PATCH `/users/{userId}/feature-flags/{flagId}` to update a feature flag for a user

For more detailed information about the API contracts, please refer to the [API contract](./openapi.yaml).

## Data Model

The Feature Flag Backend project uses DynamoDB as the database. The data model consists of two main entities:

### FeatureFlag 
- id (string) **Partition key**
- name (string) (GSI)
- description (string)
- createdAt (number)
- createdBy (string)
- updatedAt (number)
- updatedBy (string)
- status (string)


### FeatureFlagUserMapping
- userId (string) **Global Secondary Index**
- flagId (string) **Partition key** 
- status (string)
- createdAt (number)
- createdBy (string)
- updatedAt (number)
- updatedBy (string)

For a visual representation of the data model, refer to the [ER diagram](./ER%20diagram.jpg).

## Contributing

Wish to contribute? You can find a detailed guide [here](./CONTRIBUTING.md)
