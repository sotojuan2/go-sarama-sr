# Confluent Cloud ACLs and Roles

This documentation explains the roles and ACLs (Access Control Lists) required to work with Confluent Cloud, both for Kafka and Schema Registry.

## Table of Contents

- [Basic Concepts](#basic-concepts)
- [Required Service Accounts](#required-service-accounts)
- [ACLs for Basic Cluster](#acls-for-basic-cluster)
- [RBAC Roles for Dedicated Cluster](#rbac-roles-for-dedicated-cluster)
- [Configuration in .env](#configuration-in-env)
- [Configuration Scripts](#configuration-scripts)
- [Verification and Troubleshooting](#verification-and-troubleshooting)

## Basic Concepts

### Differences between ACLs and RBAC

- **ACLs (Access Control Lists)**: Used in **basic** Confluent Cloud clusters
- **RBAC (Role-Based Access Control)**: Used in **dedicated** and **standard** clusters

### Service Accounts vs User Accounts

- **Service Accounts**: For automated applications (recommended for production)
- **User Accounts**: For individual developers

## Required Service Accounts

For this project you need **two service accounts** with their respective API keys:

### 1. Service Account for Kafka
```bash
# Create service account for Kafka
confluent iam service-account create "go-sarama-kafka" --description "Service account for Kafka producers/consumers"

# Example response:
# +-------------+------------------+
# | ID          | sa-pg9nnk5       |
# | Name        | go-sarama-kafka  |
# | Description | Service account for Kafka producers/consumers |
# +-------------+------------------+
```

### 2. Service Account for Schema Registry
```bash
# Create service account for Schema Registry
confluent iam service-account create "go-sarama-sr" --description "Service account for Schema Registry"

# Example response:
# +-------------+------------------+
# | ID          | sa-xq7ookz       |
# | Name        | go-sarama-sr     |
# | Description | Service account for Schema Registry |
# +-------------+------------------+
```

## ACLs for Basic Cluster

### Current ACLs Visualization

Based on the provided image, the configured ACLs are:

```
Principal         | Permission | Operation        | Resource Type | Resource Name | Pattern Type
------------------|------------|------------------|---------------|---------------|-------------
User:sa-pg9nnk5   | ALLOW      | IDEMPOTENT_WRITE | CLUSTER       | kafka-cluster | LITERAL
User:sa-pg9nnk5   | ALLOW      | READ             | TOPIC         | js            | PREFIXED
User:sa-pg9nnk5   | ALLOW      | WRITE            | TOPIC         | js            | PREFIXED
```

### Commands to Configure ACLs in Basic Cluster

#### 1. ACLs for Kafka (Service Account: sa-pg9nnk5)

```bash
# List existing ACLs
confluent kafka acl list --service-account sa-pg9nnk5

# IDEMPOTENT_WRITE on cluster (for idempotent producers)
confluent kafka acl create \
  --allow \
  --service-account sa-pg9nnk5 \
  --operation IDEMPOTENT_WRITE \
  --cluster-scope

# READ on topics with prefix "js" (for consumers)
confluent kafka acl create \
  --allow \
  --service-account sa-pg9nnk5 \
  --operation READ \
  --topic js \
  --prefix

# WRITE on topics with prefix "js" (for producers)
confluent kafka acl create \
  --allow \
  --service-account sa-pg9nnk5 \
  --operation WRITE \
  --topic js \
  --prefix

# READ on consumer groups (required for consumers)
confluent kafka acl create \
  --allow \
  --service-account sa-pg9nnk5 \
  --operation READ \
  --consumer-group go-sarama-consumer \
  --prefix
```

#### 2. API Keys for Kafka

```bash
# Create API key for Kafka service account
confluent api-key create --resource <KAFKA_CLUSTER_ID> --service-account sa-pg9nnk5

# Example output:
# +------------+------------------------------------------------------------------+
# | API Key    | EW6SNF25TI73HOU7                                                 |
# | API Secret | cflt+dflBDGCBnaVcLe7DcGa9bYtdQtWWLhji/lH5YCYF1MKjnehgvm2sVmNLmGQ |
# +------------+------------------------------------------------------------------+
```

#### 3. ACLs for Schema Registry

For Schema Registry in basic cluster, ACLs are also used:

```bash
# Create API key for Schema Registry
confluent api-key create --resource <SCHEMA_REGISTRY_CLUSTER_ID> --service-account sa-xq7ookz

# Example output:
# +------------+------------------------------------------------------------------+
# | API Key    | CCCB5CHQBSL7FPUK                                                 |
# | API Secret | cflt/NDN30Aa/Eqq4shUBhf2O7GLU5RPsvEpxpqWzSg5yRAvBieeLKa99dt/16PA |
# +------------+------------------------------------------------------------------+
```

## RBAC Roles for Dedicated Cluster

### Schema Registry RBAC (Dedicated/Standard Clusters)

If you use a dedicated or standard cluster, you need to configure RBAC roles:

```bash
# ResourceOwner for all subjects in Schema Registry
confluent iam rbac role-binding create \
  --principal User:sa-xq7ookz \
  --role ResourceOwner \
  --environment env-9zj7y5 \
  --schema-registry-cluster lsrc-9380d5 \
  --resource "Subject:*"

# DeveloperRead to read schemas (more restrictive alternative)
confluent iam rbac role-binding create \
  --principal User:sa-xq7ookz \
  --role DeveloperRead \
  --environment env-9zj7y5 \
  --schema-registry-cluster lsrc-9380d5 \
  --resource "Subject:js_shoe-value"

# DeveloperWrite to write schemas
confluent iam rbac role-binding create \
  --principal User:sa-xq7ookz \
  --role DeveloperWrite \
  --environment env-9zj7y5 \
  --schema-registry-cluster lsrc-9380d5 \
  --resource "Subject:js_shoe-value"
```

### Kafka RBAC (Dedicated/Standard Clusters)

```bash
# DeveloperWrite for topics
confluent iam rbac role-binding create \
  --principal User:sa-pg9nnk5 \
  --role DeveloperWrite \
  --environment env-9zj7y5 \
  --kafka-cluster lkc-xxxxx \
  --resource "Topic:js_shoe"

# DeveloperRead for consumer groups
confluent iam rbac role-binding create \
  --principal User:sa-pg9nnk5 \
  --role DeveloperRead \
  --environment env-9zj7y5 \
  --kafka-cluster lkc-xxxxx \
  --resource "Group:go-sarama-consumer"
```

## Configuration in .env

The `.env` file must contain credentials for both service accounts:

```env
# Kafka Cluster Configuration
KAFKA_BOOTSTRAP_SERVERS=pkc-mxqvx.europe-southwest1.gcp.confluent.cloud:9092
KAFKA_API_KEY=EW6SNF25TI73HOU7
KAFKA_API_SECRET=cflt+dflBDGCBnaVcLe7DcGa9bYtdQtWWLhji/lH5YCYF1MKjnehgvm2sVmNLmGQ

# Topic Configuration
KAFKA_TOPIC=js_shoe

# Schema Registry Configuration
SCHEMA_REGISTRY_URL=https://psrc-8qmnr.eu-west-2.aws.confluent.cloud
SCHEMA_REGISTRY_API_KEY=CCCB5CHQBSL7FPUK
SCHEMA_REGISTRY_API_SECRET=cflt/NDN30Aa/Eqq4shUBhf2O7GLU5RPsvEpxpqWzSg5yRAvBieeLKa99dt/16PA
```

## Configuration Scripts

### Script to Generate ACLs Automatically

```bash
#!/bin/bash
# scripts/setup-acls.sh

set -e

SERVICE_ACCOUNT_KAFKA="sa-pg9nnk5"
SERVICE_ACCOUNT_SR="sa-xq7ookz"
TOPIC_PREFIX="js"
CONSUMER_GROUP_PREFIX="go-sarama-consumer"

echo "🔧 Setting up ACLs for Kafka..."

# ACLs for Kafka
confluent kafka acl create --allow --service-account $SERVICE_ACCOUNT_KAFKA --operation IDEMPOTENT_WRITE --cluster-scope
confluent kafka acl create --allow --service-account $SERVICE_ACCOUNT_KAFKA --operation READ --topic $TOPIC_PREFIX --prefix
confluent kafka acl create --allow --service-account $SERVICE_ACCOUNT_KAFKA --operation WRITE --topic $TOPIC_PREFIX --prefix
confluent kafka acl create --allow --service-account $SERVICE_ACCOUNT_KAFKA --operation READ --consumer-group $CONSUMER_GROUP_PREFIX --prefix

echo "✅ ACLs configured successfully"
```

## Using ACLs with Consumer Scripts

### Script for Magic Byte and Schema ID Verification

The `confluent_cli/schema_registry_consumer.sh` script uses the configured **READ** ACLs:

```bash
# READ ACLs allow:
# 1. Read messages from js_shoe topic
# 2. Decode magic byte (0x00)
# 3. Extract schema ID
# 4. Query schema in Schema Registry

./confluent_cli/schema_registry_consumer.sh
```

### Required Permissions for the Script

```
┌─────────────────────────────────────────────────────────────────┐
│ Permission Flow for schema_registry_consumer.sh                 │
├─────────────────────────────────────────────────────────────────┤
│ 1. READ on topic js_shoe (Kafka ACL)                           │
│    └─> Read binary messages from topic                         │
│                                                                 │
│ 2. Local message decoding                                       │
│    ├─> Magic byte (0x00)                                       │
│    ├─> Schema ID (4 bytes)                                     │
│    └─> Avro payload                                            │
│                                                                 │
│ 3. READ on Schema Registry (SR API Key)                        │
│    └─> Query schema by ID                                      │
│                                                                 │
│ 4. Display formatted information                               │
│    ├─> Offset, partition, timestamp                            │
│    ├─> Schema ID and details                                   │
│    └─> Decoded payload                                         │
└─────────────────────────────────────────────────────────────────┘
```

## Verification and Troubleshooting

### Verify Configured ACLs

```bash
# List all ACLs for the service account
confluent kafka acl list --service-account sa-pg9nnk5

# Verify Kafka connectivity
./confluent_cli/test_consumer.sh

# Verify Schema Registry connectivity
curl -u "${SCHEMA_REGISTRY_API_KEY}:${SCHEMA_REGISTRY_API_SECRET}" \
  "${SCHEMA_REGISTRY_URL}/subjects"
```

### Common Issues

#### 1. Kafka Authorization Error
```
Error: Topic authorization failed
```
**Solution**: Verify that the service account has READ/WRITE permissions on the topic

#### 2. Schema Registry Error
```
Error: 401 Unauthorized
```
**Solution**: Verify that Schema Registry credentials are correct

#### 3. Consumer Group Permissions
```
Error: Group authorization failed
```
**Solution**: Add READ permissions for consumer groups:
```bash
confluent kafka acl create --allow --service-account sa-pg9nnk5 --operation READ --consumer-group go-sarama-consumer --prefix
```

### Diagnostic Commands

```bash
# Verify service accounts
confluent iam service-account list

# Verify API keys
confluent api-key list

# Verify available clusters
confluent kafka cluster list

# Verify Schema Registry
confluent schema-registry cluster list
```

## Security and Best Practices

### 1. Principle of Least Privilege
- Use specific permissions instead of wildcards when possible
- Separate service accounts by function (Kafka vs Schema Registry)

### 2. Credential Rotation
```bash
# Create new API key
confluent api-key create --resource <CLUSTER_ID> --service-account <SA_ID>

# Delete old API key
confluent api-key delete <OLD_API_KEY>
```

### 3. Access Monitoring
- Regularly review audit logs
- Use monitoring tools to detect anomalous access

### 4. Environment Variables
- Never hardcode credentials in code
- Use local `.env` files (not versioned)
- In production, use secret management systems

## References

- [Confluent Cloud ACL Documentation](https://docs.confluent.io/cloud/current/access-management/acl.html)
- [Confluent Cloud RBAC Documentation](https://docs.confluent.io/cloud/current/access-management/rbac.html)
- [Schema Registry Security](https://docs.confluent.io/platform/current/schema-registry/security/index.html)
