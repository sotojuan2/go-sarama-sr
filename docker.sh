#!/bin/bash

# Docker Helper Script for Go Sarama Schema Registry Producer
# This script provides convenient commands for building and running the containerized producers
# Note: This application requires librdkafka (C library) for Schema Registry integration
# Architecture: Uses Debian-based images for glibc compatibility with librdkafka

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
IMAGE_NAME="go-sarama-producer"
IMAGE_TAG="latest"
FULL_IMAGE_NAME="${IMAGE_NAME}:${IMAGE_TAG}"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

show_help() {
    echo "Docker Helper Script for Go Sarama Producer"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  build                 Build the Docker image"
    echo "  run [PRODUCER]        Run a specific producer (continuous|robust|enhanced)"
    echo "  run-compose [PROFILE] Run using docker-compose with profile"
    echo "  logs [CONTAINER]      Show logs for a container"
    echo "  stop                  Stop all running containers"
    echo "  clean                 Clean up containers and images"
    echo "  test                  Build and test all producers"
    echo "  help                  Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 build"
    echo "  $0 run continuous"
    echo "  $0 run-compose robust"
    echo "  $0 logs continuous-producer"
    echo "  $0 test"
    echo ""
    echo "Producer Types:"
    echo "  continuous  - MVP continuous producer (recommended for production)"
    echo "  basic       - Simple producer implementation (minimal features)"
    echo ""
    echo "Note: robust and enhanced producers are temporarily unavailable in Docker"
    echo "      due to missing errorhandling, logging, and metrics dependencies"
}

check_env_file() {
    if [[ ! -f ".env" ]]; then
        log_warning ".env file not found!"
        log_info "Creating .env from .env.example..."
        if [[ -f ".env.example" ]]; then
            cp .env.example .env
            log_warning "Please edit .env with your actual Confluent Cloud credentials before running!"
            return 1
        else
            log_error ".env.example file not found. Cannot create .env file."
            return 1
        fi
    fi
    return 0
}

build_image() {
    log_info "Building Docker image: ${FULL_IMAGE_NAME}"
    
    if docker build -t "${FULL_IMAGE_NAME}" .; then
        log_success "Docker image built successfully!"
        docker images | grep "${IMAGE_NAME}" | head -5
    else
        log_error "Failed to build Docker image"
        return 1
    fi
}

run_producer() {
    local producer=$1
    if [[ -z "$producer" ]]; then
        log_error "Producer type not specified. Use: continuous or basic"
        return 1
    fi
    
    check_env_file || return 1
    
    local binary_name
    case $producer in
        continuous)
            binary_name="continuous_producer"
            ;;
        basic)
            binary_name="producer"
            ;;
        robust|enhanced)
            log_error "Producer type '$producer' is temporarily unavailable in Docker"
            log_info "Missing dependencies: errorhandling, logging, and metrics packages"
            log_info "Run directly outside Docker until dependencies are completed"
            return 1
            ;;
        *)
            log_error "Invalid producer type: $producer"
            log_info "Valid types: continuous, basic"
            return 1
            ;;
    esac
    
    log_info "Running ${producer} producer..."
    docker run --rm -it \
        --name "${producer}-producer" \
        --env-file .env \
        "${FULL_IMAGE_NAME}" \
        "./bin/${binary_name}"
}

run_compose() {
    local profile=${1:-continuous}
    
    check_env_file || return 1
    
    log_info "Running docker-compose with profile: ${profile}"
    docker-compose --profile "${profile}" up
}

show_logs() {
    local container=${1:-continuous-producer}
    log_info "Showing logs for container: ${container}"
    
    if docker ps --format "table {{.Names}}" | grep -q "${container}"; then
        docker logs -f "${container}"
    elif docker-compose ps --services | grep -q "${container}"; then
        docker-compose logs -f "${container}"
    else
        log_error "Container ${container} not found"
        log_info "Available containers:"
        docker ps --format "table {{.Names}}\t{{.Status}}"
    fi
}

stop_containers() {
    log_info "Stopping all producer containers..."
    
    # Stop docker-compose services
    if [[ -f "docker-compose.yml" ]]; then
        docker-compose down
    fi
    
    # Stop any standalone containers
    for container in continuous-producer robust-producer enhanced-producer; do
        if docker ps -q -f name="${container}" | grep -q .; then
            log_info "Stopping ${container}..."
            docker stop "${container}" 2>/dev/null || true
        fi
    done
    
    log_success "All containers stopped"
}

clean_up() {
    log_info "Cleaning up containers and images..."
    
    stop_containers
    
    # Remove containers
    docker container prune -f
    
    # Remove the built image
    if docker images -q "${FULL_IMAGE_NAME}" | grep -q .; then
        log_info "Removing image: ${FULL_IMAGE_NAME}"
        docker rmi "${FULL_IMAGE_NAME}"
    fi
    
    # Clean up docker-compose
    if [[ -f "docker-compose.yml" ]]; then
        docker-compose down --rmi all --volumes --remove-orphans
    fi
    
    log_success "Cleanup completed"
}

test_all() {
    log_info "Building and testing available producers..."
    
    # Build the image
    build_image || return 1
    
    check_env_file || return 1
    
    # Test each available producer (run for 10 seconds each)
    for producer in continuous basic; do
        log_info "Testing ${producer} producer..."
        local binary_name
        case $producer in
            continuous) binary_name="continuous_producer" ;;
            basic) binary_name="producer" ;;
        esac
        
        timeout 10s docker run --rm \
            --env-file .env \
            "${FULL_IMAGE_NAME}" \
            "./bin/${binary_name}" || {
                if [[ $? -eq 124 ]]; then
                    log_success "${producer} producer test completed (timed out as expected)"
                else
                    log_error "${producer} producer test failed"
                    return 1
                fi
            }
    done
    
    log_success "All available producer tests completed successfully!"
    log_info "Note: robust and enhanced producers are not available in Docker yet"
}

# Main script logic
case "${1:-help}" in
    build)
        build_image
        ;;
    run)
        run_producer "$2"
        ;;
    run-compose)
        run_compose "$2"
        ;;
    logs)
        show_logs "$2"
        ;;
    stop)
        stop_containers
        ;;
    clean)
        clean_up
        ;;
    test)
        test_all
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        log_error "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
