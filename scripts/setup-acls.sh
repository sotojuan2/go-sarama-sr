#!/bin/bash

# =============================================================================
# Confluent Cloud ACLs Setup Script
# =============================================================================
# Este script configura automáticamente las ACLs necesarias para el proyecto
# go-sarama-sr, incluyendo permisos para Kafka y Schema Registry.
#
# Uso:
#   ./scripts/setup-acls.sh [OPTIONS]
#
# Opciones:
#   -k, --kafka-sa      Service account ID para Kafka (ej: sa-pg9nnk5)
#   -s, --sr-sa         Service account ID para Schema Registry (ej: sa-xq7ookz)
#   -t, --topic-prefix  Prefijo de tópicos (default: js)
#   -g, --group-prefix  Prefijo de consumer groups (default: go-sarama-consumer)
#   -c, --cluster-id    ID del cluster Kafka (opcional)
#   -r, --sr-cluster    ID del cluster Schema Registry (opcional)
#   -e, --environment   ID del environment (para RBAC en clusters dedicados)
#   --rbac              Usar RBAC en lugar de ACLs (para clusters dedicados)
#   --dry-run           Mostrar comandos sin ejecutar
#   -h, --help          Mostrar esta ayuda
#
# Ejemplos:
#   # Configurar ACLs básicas
#   ./scripts/setup-acls.sh -k sa-pg9nnk5 -s sa-xq7ookz
#
#   # Configurar RBAC para cluster dedicado
#   ./scripts/setup-acls.sh -k sa-pg9nnk5 -s sa-xq7ookz --rbac -e env-9zj7y5
#
#   # Solo mostrar comandos (dry run)
#   ./scripts/setup-acls.sh -k sa-pg9nnk5 -s sa-xq7ookz --dry-run
# =============================================================================

set -e

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Variables por defecto
KAFKA_SA=""
SR_SA=""
TOPIC_PREFIX="js"
CONSUMER_GROUP_PREFIX="go-sarama-consumer"
CLUSTER_ID=""
SR_CLUSTER_ID=""
ENVIRONMENT_ID=""
USE_RBAC=false
DRY_RUN=false

# Función para mostrar ayuda
show_help() {
    head -n 40 "$0" | tail -n +3 | sed 's/^# //' | sed 's/^#//'
}

# Función para logging
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Función para ejecutar comandos (con soporte para dry-run)
execute_command() {
    local cmd="$1"
    local description="$2"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        echo -e "${YELLOW}[DRY-RUN]${NC} $description"
        echo "  Command: $cmd"
    else
        log "$description"
        if eval "$cmd"; then
            success "✅ $description completado"
        else
            error "❌ Error ejecutando: $description"
            return 1
        fi
    fi
}

# Parsear argumentos
while [[ $# -gt 0 ]]; do
    case $1 in
        -k|--kafka-sa)
            KAFKA_SA="$2"
            shift 2
            ;;
        -s|--sr-sa)
            SR_SA="$2"
            shift 2
            ;;
        -t|--topic-prefix)
            TOPIC_PREFIX="$2"
            shift 2
            ;;
        -g|--group-prefix)
            CONSUMER_GROUP_PREFIX="$2"
            shift 2
            ;;
        -c|--cluster-id)
            CLUSTER_ID="$2"
            shift 2
            ;;
        -r|--sr-cluster)
            SR_CLUSTER_ID="$2"
            shift 2
            ;;
        -e|--environment)
            ENVIRONMENT_ID="$2"
            shift 2
            ;;
        --rbac)
            USE_RBAC=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            error "Opción desconocida: $1"
            show_help
            exit 1
            ;;
    esac
done

# Validar argumentos requeridos
if [[ -z "$KAFKA_SA" ]]; then
    error "Service account de Kafka es requerido (-k, --kafka-sa)"
    exit 1
fi

if [[ -z "$SR_SA" ]]; then
    error "Service account de Schema Registry es requerido (-s, --sr-sa)"
    exit 1
fi

if [[ "$USE_RBAC" == "true" && -z "$ENVIRONMENT_ID" ]]; then
    error "Environment ID es requerido para RBAC (-e, --environment)"
    exit 1
fi

# Mostrar configuración
echo "============================================="
echo "🔧 Confluent Cloud ACLs Setup"
echo "============================================="
echo "Kafka Service Account:     $KAFKA_SA"
echo "Schema Registry SA:        $SR_SA"
echo "Topic Prefix:              $TOPIC_PREFIX"
echo "Consumer Group Prefix:     $CONSUMER_GROUP_PREFIX"
echo "Use RBAC:                  $USE_RBAC"
echo "Dry Run:                   $DRY_RUN"
if [[ -n "$ENVIRONMENT_ID" ]]; then
    echo "Environment ID:            $ENVIRONMENT_ID"
fi
if [[ -n "$CLUSTER_ID" ]]; then
    echo "Kafka Cluster ID:          $CLUSTER_ID"
fi
if [[ -n "$SR_CLUSTER_ID" ]]; then
    echo "Schema Registry Cluster:   $SR_CLUSTER_ID"
fi
echo "============================================="
echo

# Verificar que confluent CLI esté disponible
if ! command -v confluent &> /dev/null; then
    error "confluent CLI no encontrado. Instala desde: https://docs.confluent.io/confluent-cli/current/install.html"
    exit 1
fi

# Función para configurar ACLs básicas
setup_basic_acls() {
    log "🔧 Configurando ACLs para cluster básico..."
    
    # ACLs para Kafka
    execute_command "confluent kafka acl create --allow --service-account $KAFKA_SA --operation IDEMPOTENT_WRITE --cluster-scope" \
        "Crear ACL IDEMPOTENT_WRITE para cluster"
    
    execute_command "confluent kafka acl create --allow --service-account $KAFKA_SA --operation READ --topic $TOPIC_PREFIX --prefix" \
        "Crear ACL READ para tópicos con prefijo '$TOPIC_PREFIX'"
    
    execute_command "confluent kafka acl create --allow --service-account $KAFKA_SA --operation WRITE --topic $TOPIC_PREFIX --prefix" \
        "Crear ACL WRITE para tópicos con prefijo '$TOPIC_PREFIX'"
    
    execute_command "confluent kafka acl create --allow --service-account $KAFKA_SA --operation READ --consumer-group $CONSUMER_GROUP_PREFIX --prefix" \
        "Crear ACL READ para consumer groups con prefijo '$CONSUMER_GROUP_PREFIX'"
    
    success "✅ ACLs de Kafka configuradas correctamente"
}

# Función para configurar RBAC
setup_rbac() {
    log "🔧 Configurando RBAC para cluster dedicado/standard..."
    
    if [[ -z "$CLUSTER_ID" ]]; then
        warning "Cluster ID no especificado, intentando detectar automáticamente..."
        CLUSTER_ID=$(confluent kafka cluster list -o json | jq -r '.[0].id' 2>/dev/null || echo "")
        if [[ -n "$CLUSTER_ID" ]]; then
            log "Cluster detectado: $CLUSTER_ID"
        else
            error "No se pudo detectar el cluster ID automáticamente. Especifica con --cluster-id"
            return 1
        fi
    fi
    
    if [[ -z "$SR_CLUSTER_ID" ]]; then
        warning "Schema Registry Cluster ID no especificado, intentando detectar automáticamente..."
        SR_CLUSTER_ID=$(confluent schema-registry cluster list -o json | jq -r '.[0].id' 2>/dev/null || echo "")
        if [[ -n "$SR_CLUSTER_ID" ]]; then
            log "Schema Registry Cluster detectado: $SR_CLUSTER_ID"
        else
            warning "No se pudo detectar el Schema Registry Cluster ID automáticamente"
        fi
    fi
    
    # RBAC para Kafka
    execute_command "confluent iam rbac role-binding create --principal User:$KAFKA_SA --role DeveloperWrite --environment $ENVIRONMENT_ID --kafka-cluster $CLUSTER_ID --resource Topic:$TOPIC_PREFIX --prefix" \
        "Crear role binding DeveloperWrite para tópicos"
    
    execute_command "confluent iam rbac role-binding create --principal User:$KAFKA_SA --role DeveloperRead --environment $ENVIRONMENT_ID --kafka-cluster $CLUSTER_ID --resource Group:$CONSUMER_GROUP_PREFIX --prefix" \
        "Crear role binding DeveloperRead para consumer groups"
    
    # RBAC para Schema Registry (si está disponible)
    if [[ -n "$SR_CLUSTER_ID" ]]; then
        execute_command "confluent iam rbac role-binding create --principal User:$SR_SA --role ResourceOwner --environment $ENVIRONMENT_ID --schema-registry-cluster $SR_CLUSTER_ID --resource Subject:*" \
            "Crear role binding ResourceOwner para Schema Registry (todos los subjects)"
    else
        warning "Schema Registry Cluster ID no disponible, saltando configuración RBAC para SR"
    fi
    
    success "✅ RBAC configurado correctamente"
}

# Función para verificar configuración
verify_setup() {
    log "🔍 Verificando configuración..."
    
    if [[ "$USE_RBAC" == "true" ]]; then
        execute_command "confluent iam rbac role-binding list --principal User:$KAFKA_SA" \
            "Listar role bindings para Kafka SA"
        
        if [[ -n "$SR_CLUSTER_ID" ]]; then
            execute_command "confluent iam rbac role-binding list --principal User:$SR_SA" \
                "Listar role bindings para Schema Registry SA"
        fi
    else
        execute_command "confluent kafka acl list --service-account $KAFKA_SA" \
            "Listar ACLs para Kafka SA"
    fi
}

# Función principal
main() {
    if [[ "$USE_RBAC" == "true" ]]; then
        setup_rbac
    else
        setup_basic_acls
    fi
    
    echo
    verify_setup
    
    echo
    echo "============================================="
    success "🎉 Configuración completada!"
    echo "============================================="
    echo
    
    if [[ "$DRY_RUN" == "false" ]]; then
        log "Próximos pasos:"
        echo "1. ✅ Crear API keys para los service accounts:"
        echo "   confluent api-key create --resource <KAFKA_CLUSTER_ID> --service-account $KAFKA_SA"
        echo "   confluent api-key create --resource <SR_CLUSTER_ID> --service-account $SR_SA"
        echo
        echo "2. ✅ Actualizar tu archivo .env con las nuevas credenciales"
        echo
        echo "3. ✅ Probar la conectividad:"
        echo "   ./confluent_cli/test_consumer.sh"
        echo "   ./confluent_cli/schema_registry_consumer.sh"
    else
        log "Ejecuta el script sin --dry-run para aplicar los cambios"
    fi
}

# Ejecutar función principal
main
