.PHONY: build

build:
	sam build

dev_db_setup:
	java -D"java.library.path=$(DYNAMODB_LOCAL_ROOT_FOLDER_PATH)DynamoDBLocal_lib" -jar $(DYNAMODB_LOCAL_ROOT_FOLDER_PATH)DynamoDBLocal.jar -sharedDb

test_db_setup:
	java -D"java.library.path=$(DYNAMODB_LOCAL_ROOT_FOLDER_PATH)DynamoDBLocal_lib" -jar $(DYNAMODB_LOCAL_ROOT_FOLDER_PATH)DynamoDBLocal.jar -inMemory