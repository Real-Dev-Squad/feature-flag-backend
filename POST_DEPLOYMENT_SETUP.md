# Post-Deployment Setup Guide

After deploying with `sam deploy`, you need to set up the following resources:

## 🔴 Critical: DynamoDB Tables

The error you're seeing is because **DynamoDB tables don't exist**. You need to create 3 tables:

### Quick Setup (Using Script)

```bash
# Make script executable
chmod +x setup-dynamodb-tables.sh

# Run the script (defaults to us-east-1)
./setup-dynamodb-tables.sh

# OR specify a region
./setup-dynamodb-tables.sh us-east-1
```

### Manual Setup

#### 1. Create `featureFlag` Table

```bash
aws dynamodb create-table \
  --table-name featureFlag \
  --attribute-definitions AttributeName=id,AttributeType=S \
  --key-schema AttributeName=id,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region us-east-1
```

#### 2. Create `featureFlagUserMapping` Table

```bash
aws dynamodb create-table \
  --table-name featureFlagUserMapping \
  --attribute-definitions \
    AttributeName=userId,AttributeType=S \
    AttributeName=flagId,AttributeType=S \
  --key-schema \
    AttributeName=userId,KeyType=HASH \
    AttributeName=flagId,KeyType=RANGE \
  --billing-mode PAY_PER_REQUEST \
  --region us-east-1
```

#### 3. Create `requestLimit` Table (This is the missing one causing your error!)

```bash
aws dynamodb create-table \
  --table-name requestLimit \
  --attribute-definitions AttributeName=limitType,AttributeType=S \
  --key-schema AttributeName=limitType,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region us-east-1
```

#### 4. Initialize `requestLimit` Table

After creating the table, you need to add an initial value:

```bash
aws dynamodb put-item \
  --table-name requestLimit \
  --item '{
    "limitType": {"S": "pendingLimit"},
    "limitValue": {"N": "1000"}
  }' \
  --region us-east-1
```

---

## ✅ Verify Tables Exist

```bash
# List all tables
aws dynamodb list-tables --region us-east-1

# Check specific table
aws dynamodb describe-table --table-name requestLimit --region us-east-1
```

You should see:
- `featureFlag`
- `featureFlagUserMapping`
- `requestLimit`

---

## 🔑 SSM Parameter for JWT Public Key

Make sure the JWT public key exists in AWS Systems Manager Parameter Store:

### For PRODUCTION:

```bash
# Check if parameter exists
aws ssm get-parameter \
  --name PROD_RDS_BACKEND_PUBLIC_KEY \
  --with-decryption \
  --region us-east-1

# If it doesn't exist, create it:
aws ssm put-parameter \
  --name PROD_RDS_BACKEND_PUBLIC_KEY \
  --value "-----BEGIN PUBLIC KEY-----
YOUR_PUBLIC_KEY_HERE
-----END PUBLIC KEY-----" \
  --type SecureString \
  --region us-east-1
```

### For DEVELOPMENT:

```bash
# Check if parameter exists
aws ssm get-parameter \
  --name STAGING_RDS_BACKEND_PUBLIC_KEY \
  --with-decryption \
  --region us-east-1

# If it doesn't exist, create it:
aws ssm put-parameter \
  --name STAGING_RDS_BACKEND_PUBLIC_KEY \
  --value "-----BEGIN PUBLIC KEY-----
YOUR_PUBLIC_KEY_HERE
-----END PUBLIC KEY-----" \
  --type SecureString \
  --region us-east-1
```

---

## 🧪 Test After Setup

After creating the tables, test your API:

**Note:** Replace `{YOUR_API_GATEWAY_URL}` with your actual API Gateway endpoint URL. You can find it in the SAM deployment output or AWS Console.

```bash
# 1. Health check (should work)
curl https://{YOUR_API_GATEWAY_URL}/Prod/health-check

# 2. Get feature flags (should work now)
curl -X GET "https://{YOUR_API_GATEWAY_URL}/Prod/feature-flags/" \
  -H "Cookie: rds-session-staging=YOUR_JWT_TOKEN" \
  -H "Origin: https://test.realdevsquad.com"
```

---

## 📋 Complete Checklist

- [ ] **DynamoDB Tables Created:**
  - [ ] `featureFlag` table exists
  - [ ] `featureFlagUserMapping` table exists
  - [ ] `requestLimit` table exists
  - [ ] `requestLimit` table has initial item with `limitType: "pendingLimit"` and `limitValue: 1000`

- [ ] **SSM Parameter:**
  - [ ] `PROD_RDS_BACKEND_PUBLIC_KEY` exists (for PRODUCTION)
  - [ ] OR `STAGING_RDS_BACKEND_PUBLIC_KEY` exists (for DEVELOPMENT)

- [ ] **Testing:**
  - [ ] Health check endpoint works
  - [ ] Feature flags endpoints work with JWT token

---

## 🔍 Troubleshooting

### Error: "ResourceNotFoundException: Requested resource not found"

**Cause:** DynamoDB table doesn't exist

**Solution:** Create the missing table using the commands above

### Error: "invalid memory address or nil pointer dereference"

**Cause:** Code is trying to unmarshal a nil response from DynamoDB (table doesn't exist or item doesn't exist)

**Solution:**
1. Create the `requestLimit` table
2. Initialize it with the default value (see step 4 above)

### Error: "ParameterNotFound" when calling API

**Cause:** SSM Parameter for JWT public key doesn't exist

**Solution:** Create the SSM parameter with the public key (see SSM Parameter section above)

### Error: "AccessDeniedException" when creating tables

**Cause:** Your AWS credentials don't have DynamoDB permissions

**Solution:** Ensure your AWS user/role has:
- `dynamodb:CreateTable`
- `dynamodb:PutItem`
- `dynamodb:DescribeTable`
- `dynamodb:ListTables`

---

## 🚀 Quick Setup Command

Run this to set up everything:

```bash
# 1. Create all tables
./setup-dynamodb-tables.sh us-east-1

# 2. Verify tables
aws dynamodb list-tables --region us-east-1

# 3. Test API (replace {YOUR_API_GATEWAY_URL} with your actual endpoint)
curl https://{YOUR_API_GATEWAY_URL}/Prod/health-check
```

---

## 📝 Notes

- Tables are created with **PAY_PER_REQUEST** billing mode (no capacity planning needed)
- The `requestLimit` table is initialized with a default value of 1000
- You can adjust the initial `limitValue` based on your needs
- All tables are created in the same region as your Lambda functions
