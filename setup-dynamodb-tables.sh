#!/bin/bash

# Script to create all required DynamoDB tables for feature-flag-backend
# Usage: ./setup-dynamodb-tables.sh [region]

set -e

REGION=${1:-us-east-1}

# Helper function to check if a table exists
table_exists() {
  local table_name="$1"
  aws dynamodb describe-table \
    --table-name "$table_name" \
    --region "$REGION" \
    --no-cli-pager >/dev/null 2>&1
}

echo "🚀 Setting up DynamoDB tables in region: $REGION"
echo ""

# Table 1: featureFlag
if ! table_exists "featureFlag"; then
  echo "Creating table: featureFlag"
  aws dynamodb create-table \
    --table-name featureFlag \
    --attribute-definitions \
      AttributeName=id,AttributeType=S \
    --key-schema \
      AttributeName=id,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST \
    --region "$REGION" \
    --no-cli-pager
  echo "✅ Created featureFlag table"
else
  echo "⏭️  Table featureFlag already exists, skipping creation"
fi
echo ""

# Table 2: featureFlagUserMapping
if ! table_exists "featureFlagUserMapping"; then
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
    --region "$REGION" \
    --no-cli-pager
  echo "✅ Created featureFlagUserMapping table"
else
  echo "⏭️  Table featureFlagUserMapping already exists, skipping creation"
fi
echo ""

# Table 3: user
if ! table_exists "user"; then
  echo "Creating table: user"
  aws dynamodb create-table \
    --table-name user \
    --attribute-definitions \
      AttributeName=id,AttributeType=S \
      AttributeName=email,AttributeType=S \
    --key-schema \
      AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
      "[{\"IndexName\": \"email-index\", \"KeySchema\": [{\"AttributeName\": \"email\", \"KeyType\": \"HASH\"}], \"Projection\": {\"ProjectionType\": \"ALL\"}}]" \
    --billing-mode PAY_PER_REQUEST \
    --region "$REGION" \
    --no-cli-pager
  echo "✅ Created user table"
else
  echo "⏭️  Table user already exists, skipping creation"
fi
echo ""

# Table 4: requestLimit
if ! table_exists "requestLimit"; then
  echo "Creating table: requestLimit"
  aws dynamodb create-table \
    --table-name requestLimit \
    --attribute-definitions \
      AttributeName=limitType,AttributeType=S \
    --key-schema \
      AttributeName=limitType,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST \
    --region "$REGION" \
    --no-cli-pager
  echo "✅ Created requestLimit table"
else
  echo "⏭️  Table requestLimit already exists, skipping creation"
fi
echo ""

# Wait for tables to be active
echo "⏳ Waiting for tables to become active..."
aws dynamodb wait table-exists \
  --table-name featureFlag \
  --region "$REGION"

aws dynamodb wait table-exists \
  --table-name featureFlagUserMapping \
  --region "$REGION"

aws dynamodb wait table-exists \
  --table-name user \
  --region "$REGION"

aws dynamodb wait table-exists \
  --table-name requestLimit \
  --region "$REGION"

echo "✅ All tables are active!"
echo ""

# Initialize requestLimit table with default value (idempotent)
echo "Initializing requestLimit table with default value..."
# Check if item already exists
if ! aws dynamodb get-item \
  --table-name requestLimit \
  --key '{"limitType": {"S": "pendingLimit"}}' \
  --region "$REGION" \
  --no-cli-pager | grep -q "Item"; then
  aws dynamodb put-item \
    --table-name requestLimit \
    --item '{
      "limitType": {"S": "pendingLimit"},
      "limitValue": {"N": "1000"}
    }' \
    --region "$REGION" \
    --no-cli-pager
  echo "✅ Initialized requestLimit with default value (1000)"
else
  echo "⏭️  requestLimit already initialized, skipping"
fi
echo ""

echo "🎉 All DynamoDB tables created and initialized successfully!"
echo ""
echo "Tables created:"
echo "  - featureFlag"
echo "  - featureFlagUserMapping"
echo "  - user"
echo "  - requestLimit (with initial value)"
echo ""
echo "You can now test your API endpoints!"
