# Confluent Cloud ACLs y Roles

Esta documentación explica los roles y ACLs (Access Control Lists) necesarios para trabajar con Confluent Cloud, tanto para Kafka como para Schema Registry.

## Tabla de Contenidos

- [Conceptos Básicos](#conceptos-básicos)
- [Service Accounts Requeridas](#service-accounts-requeridas)
- [ACLs para Cluster Básico](#acls-para-cluster-básico)
- [Roles RBAC para Cluster Dedicado](#roles-rbac-para-cluster-dedicado)
- [Configuración en .env](#configuración-en-env)
- [Scripts de Configuración](#scripts-de-configuración)
- [Verificación y Troubleshooting](#verificación-y-troubleshooting)

## Conceptos Básicos

### Diferencias entre ACLs y RBAC

- **ACLs (Access Control Lists)**: Usadas en clusters **básicos** de Confluent Cloud
- **RBAC (Role-Based Access Control)**: Usadas en clusters **dedicados** y **standard**

### Service Accounts vs User Accounts

- **Service Accounts**: Para aplicaciones automatizadas (recomendado para producción)
- **User Accounts**: Para desarrolladores individuales

## Service Accounts Requeridas

Para este proyecto necesitas **dos service accounts** con sus respectivas API keys:

### 1. Service Account para Kafka
```bash
# Crear service account para Kafka
confluent iam service-account create "go-sarama-kafka" --description "Service account para productores/consumidores Kafka"

# Ejemplo de respuesta:
# +-------------+------------------+
# | ID          | sa-pg9nnk5       |
# | Name        | go-sarama-kafka  |
# | Description | Service account para productores/consumidores Kafka |
# +-------------+------------------+
```

### 2. Service Account para Schema Registry
```bash
# Crear service account para Schema Registry
confluent iam service-account create "go-sarama-sr" --description "Service account para Schema Registry"

# Ejemplo de respuesta:
# +-------------+------------------+
# | ID          | sa-xq7ookz       |
# | Name        | go-sarama-sr     |
# | Description | Service account para Schema Registry |
# +-------------+------------------+
```

## ACLs para Cluster Básico

### Visualización de ACLs Existentes

Basándose en la imagen proporcionada, las ACLs configuradas son:

```
Principal         | Permission | Operation        | Resource Type | Resource Name | Pattern Type
------------------|------------|------------------|---------------|---------------|-------------
User:sa-pg9nnk5   | ALLOW      | IDEMPOTENT_WRITE | CLUSTER       | kafka-cluster | LITERAL
User:sa-pg9nnk5   | ALLOW      | READ             | TOPIC         | js            | PREFIXED
User:sa-pg9nnk5   | ALLOW      | WRITE            | TOPIC         | js            | PREFIXED
```

### Comandos para Configurar ACLs en Cluster Básico

#### 1. ACLs para Kafka (Service Account: sa-pg9nnk5)

```bash
# Listar ACLs existentes
confluent kafka acl list --service-account sa-pg9nnk5

# IDEMPOTENT_WRITE en cluster (para productores idempotentes)
confluent kafka acl create \
  --allow \
  --service-account sa-pg9nnk5 \
  --operation IDEMPOTENT_WRITE \
  --cluster-scope

# READ en tópicos con prefijo "js" (para consumidores)
confluent kafka acl create \
  --allow \
  --service-account sa-pg9nnk5 \
  --operation READ \
  --topic js \
  --prefix

# WRITE en tópicos con prefijo "js" (para productores)
confluent kafka acl create \
  --allow \
  --service-account sa-pg9nnk5 \
  --operation WRITE \
  --topic js \
  --prefix

# READ en consumer groups (necesario para consumidores)
confluent kafka acl create \
  --allow \
  --service-account sa-pg9nnk5 \
  --operation READ \
  --consumer-group go-sarama-consumer \
  --prefix
```

#### 2. API Keys para Kafka

```bash
# Crear API key para el service account de Kafka
confluent api-key create --resource <KAFKA_CLUSTER_ID> --service-account sa-pg9nnk5

# Ejemplo de salida:
# +------------+------------------------------------------------------------------+
# | API Key    | EW6SNF25TI73HOU7                                                 |
# | API Secret | cflt+dflBDGCBnaVcLe7DcGa9bYtdQtWWLhji/lH5YCYF1MKjnehgvm2sVmNLmGQ |
# +------------+------------------------------------------------------------------+
```

#### 3. ACLs para Schema Registry

Para Schema Registry en cluster básico, también se usan ACLs:

```bash
# Crear API key para Schema Registry
confluent api-key create --resource <SCHEMA_REGISTRY_CLUSTER_ID> --service-account sa-xq7ookz

# Ejemplo de salida:
# +------------+------------------------------------------------------------------+
# | API Key    | CCCB5CHQBSL7FPUK                                                 |
# | API Secret | cflt/NDN30Aa/Eqq4shUBhf2O7GLU5RPsvEpxpqWzSg5yRAvBieeLKa99dt/16PA |
# +------------+------------------------------------------------------------------+
```

## Roles RBAC para Cluster Dedicado

### Schema Registry RBAC (Clusters Dedicados/Standard)

Si usas un cluster dedicado o standard, necesitas configurar roles RBAC:

```bash
# ResourceOwner para todos los subjects en Schema Registry
confluent iam rbac role-binding create \
  --principal User:sa-xq7ookz \
  --role ResourceOwner \
  --environment env-9zj7y5 \
  --schema-registry-cluster lsrc-9380d5 \
  --resource "Subject:*"

# DeveloperRead para leer schemas (alternativa más restrictiva)
confluent iam rbac role-binding create \
  --principal User:sa-xq7ookz \
  --role DeveloperRead \
  --environment env-9zj7y5 \
  --schema-registry-cluster lsrc-9380d5 \
  --resource "Subject:js_shoe-value"

# DeveloperWrite para escribir schemas
confluent iam rbac role-binding create \
  --principal User:sa-xq7ookz \
  --role DeveloperWrite \
  --environment env-9zj7y5 \
  --schema-registry-cluster lsrc-9380d5 \
  --resource "Subject:js_shoe-value"
```

### Kafka RBAC (Clusters Dedicados/Standard)

```bash
# DeveloperWrite para topics
confluent iam rbac role-binding create \
  --principal User:sa-pg9nnk5 \
  --role DeveloperWrite \
  --environment env-9zj7y5 \
  --kafka-cluster lkc-xxxxx \
  --resource "Topic:js_shoe"

# DeveloperRead para consumer groups
confluent iam rbac role-binding create \
  --principal User:sa-pg9nnk5 \
  --role DeveloperRead \
  --environment env-9zj7y5 \
  --kafka-cluster lkc-xxxxx \
  --resource "Group:go-sarama-consumer"
```

## Configuración en .env

El archivo `.env` debe contener las credenciales de ambas service accounts:

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

## Scripts de Configuración

### Script para Generar ACLs Automáticamente

```bash
#!/bin/bash
# scripts/setup-acls.sh

set -e

SERVICE_ACCOUNT_KAFKA="sa-pg9nnk5"
SERVICE_ACCOUNT_SR="sa-xq7ookz"
TOPIC_PREFIX="js"
CONSUMER_GROUP_PREFIX="go-sarama-consumer"

echo "🔧 Configurando ACLs para Kafka..."

# ACLs para Kafka
confluent kafka acl create --allow --service-account $SERVICE_ACCOUNT_KAFKA --operation IDEMPOTENT_WRITE --cluster-scope
confluent kafka acl create --allow --service-account $SERVICE_ACCOUNT_KAFKA --operation READ --topic $TOPIC_PREFIX --prefix
confluent kafka acl create --allow --service-account $SERVICE_ACCOUNT_KAFKA --operation WRITE --topic $TOPIC_PREFIX --prefix
confluent kafka acl create --allow --service-account $SERVICE_ACCOUNT_KAFKA --operation READ --consumer-group $CONSUMER_GROUP_PREFIX --prefix

echo "✅ ACLs configuradas correctamente"
```

## Uso de ACLs con Scripts de Consumo

### Script para Verificar Magic Byte y Schema ID

El script `confluent_cli/schema_registry_consumer.sh` usa las ACLs de **READ** configuradas:

```bash
# Las ACLs de READ permiten:
# 1. Leer mensajes del tópico js_shoe
# 2. Decodificar el magic byte (0x00)
# 3. Extraer el schema ID
# 4. Consultar el schema en Schema Registry

./confluent_cli/schema_registry_consumer.sh
```

### Permisos Requeridos para el Script

```
┌─────────────────────────────────────────────────────────────────┐
│ Flujo de Permisos para schema_registry_consumer.sh              │
├─────────────────────────────────────────────────────────────────┤
│ 1. READ en topic js_shoe (ACL Kafka)                           │
│    └─> Leer mensajes binarios del tópico                       │
│                                                                 │
│ 2. Decodificar mensaje localmente                              │
│    ├─> Magic byte (0x00)                                       │
│    ├─> Schema ID (4 bytes)                                     │
│    └─> Payload Avro                                            │
│                                                                 │
│ 3. READ en Schema Registry (API Key SR)                        │
│    └─> Consultar schema por ID                                 │
│                                                                 │
│ 4. Mostrar información formateada                              │
│    ├─> Offset, partition, timestamp                            │
│    ├─> Schema ID y detalles                                    │
│    └─> Payload decodificado                                    │
└─────────────────────────────────────────────────────────────────┘
```

## Verificación y Troubleshooting

### Verificar ACLs Configuradas

```bash
# Listar todas las ACLs para el service account
confluent kafka acl list --service-account sa-pg9nnk5

# Verificar conectividad a Kafka
./confluent_cli/test_consumer.sh

# Verificar conectividad a Schema Registry
curl -u "${SCHEMA_REGISTRY_API_KEY}:${SCHEMA_REGISTRY_API_SECRET}" \
  "${SCHEMA_REGISTRY_URL}/subjects"
```

### Problemas Comunes

#### 1. Error de Autorización en Kafka
```
Error: Topic authorization failed
```
**Solución**: Verificar que el service account tiene permisos READ/WRITE en el tópico

#### 2. Error de Schema Registry
```
Error: 401 Unauthorized
```
**Solución**: Verificar que las credenciales de Schema Registry son correctas

#### 3. Consumer Group Permissions
```
Error: Group authorization failed
```
**Solución**: Agregar permisos READ para consumer groups:
```bash
confluent kafka acl create --allow --service-account sa-pg9nnk5 --operation READ --consumer-group go-sarama-consumer --prefix
```

### Comandos de Diagnóstico

```bash
# Verificar service accounts
confluent iam service-account list

# Verificar API keys
confluent api-key list

# Verificar clusters disponibles
confluent kafka cluster list

# Verificar Schema Registry
confluent schema-registry cluster list
```

## Seguridad y Buenas Prácticas

### 1. Principio de Menor Privilegio
- Usa permisos específicos en lugar de wildcards cuando sea posible
- Separa service accounts por función (Kafka vs Schema Registry)

### 2. Rotación de Credenciales
```bash
# Crear nueva API key
confluent api-key create --resource <CLUSTER_ID> --service-account <SA_ID>

# Eliminar API key antigua
confluent api-key delete <OLD_API_KEY>
```

### 3. Monitoreo de Accesos
- Revisa regularmente los logs de auditoría
- Usa herramientas de monitoreo para detectar accesos anómalos

### 4. Variables de Entorno
- Nunca hardcodees credenciales en el código
- Usa archivos `.env` locales (no versionados)
- En producción, usa sistemas de gestión de secretos

## Referencias

- [Confluent Cloud ACL Documentation](https://docs.confluent.io/cloud/current/access-management/acl.html)
- [Confluent Cloud RBAC Documentation](https://docs.confluent.io/cloud/current/access-management/rbac.html)
- [Schema Registry Security](https://docs.confluent.io/platform/current/schema-registry/security/index.html)
