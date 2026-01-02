#!/bin/bash

# Script to create all required DynamoDB tables for feature-flag-backend
# Usage: ./setup-dynamodb-tables.sh [region]

set -e

REGION=${1:-us-east-1}

echo "🚀 Setting up DynamoDB tables in region: $REGION"
echo ""

# Table 1: featureFlag
echo "Creating table: featureFlag"
aws dynamodb create-table \
  --table-name featureFlag \
  --attribute-definitions \
    AttributeName=id,AttributeType=S \
  --key-schema \
    AttributeName=id,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region $REGION \
  --no-cli-pager

echo "✅ Created featureFlag table"
echo ""

# Table 2: featureFlagUserMapping
echo "Creating table: featureFlagUserMapping"
aws dynamodb create-table \
  --table-name featureFlagUserMapping \
  --attribute-definitions \
    AttributeName=userId,AttributeType=S \
    AttributeName=flagId,AttributeType=S \
  --key-schema \
    AttributeName=userId,KeyType=HASH \
    AttributeName=flagId,KeyType=RANGE \
  --billing-mode PAY_PER_REQUEST \
  --region $REGION \
  --no-cli-pager

echo "✅ Created featureFlagUserMapping table"
echo ""

# Table 3: requestLimit
echo "Creating table: requestLimit"
aws dynamodb create-table \
  --table-name requestLimit \
  --attribute-definitions \
    AttributeName=limitType,AttributeType=S \
  --key-schema \
    AttributeName=limitType,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region $REGION \
  --no-cli-pager

echo "✅ Created requestLimit table"
echo ""

# Wait for tables to be active
echo "⏳ Waiting for tables to become active..."
aws dynamodb wait table-exists \
  --table-name featureFlag \
  --region $REGION

aws dynamodb wait table-exists \
  --table-name featureFlagUserMapping \
  --region $REGION

aws dynamodb wait table-exists \
  --table-name requestLimit \
  --region $REGION

echo "✅ All tables are active!"
echo ""

# Initialize requestLimit table with default value
echo "Initializing requestLimit table with default value..."
aws dynamodb put-item \
  --table-name requestLimit \
  --item '{
    "limitType": {"S": "pendingLimit"},
    "limitValue": {"N": "1000"}
  }' \
  --region $REGION \
  --no-cli-pager

echo "✅ Initialized requestLimit with default value (1000)"
echo ""

echo "🎉 All DynamoDB tables created and initialized successfully!"
echo ""
echo "Tables created:"
echo "  - featureFlag"
echo "  - featureFlagUserMapping"
echo "  - requestLimit (with initial value)"
echo ""
echo "You can now test your API endpoints!"
