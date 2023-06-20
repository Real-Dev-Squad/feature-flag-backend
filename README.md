# Feature Flag Backend

The Feature Flag Backend service helps manage feature flags for your applications in Real Dev Squad. With feature flags, we can easily enable or disable specific features for different groups of users. It provides APIs for creating, updating, and retrieving feature flags and user mappings. The backend is built using Golang and is deployed using a serverless architecture on AWS Lambda. It uses DynamoDB as the database. Whenever we want to roll out new features gradually or experiment with different options, the Feature Flag Backend simplifies the process and gives us full control over your features.

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
- [SAM-CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)
- [Golang](https://go.dev/)
- [Docker](https://www.docker.com/)

**Open the terminal or command prompt:** Depending on your operating system, open the terminal or command prompt to begin the cloning process.

**Navigate to your desired local directory:** Use the cd command to navigate to the directory where you want to store the cloned repository.

**Clone the repository:** Use the following command to clone the repository :

```
git clone https://github.com/Real-Dev-Squad/feature-flag-backend.git
```

## Run

//TODO

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
- Id (string) **Partition key**
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
