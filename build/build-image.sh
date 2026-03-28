#!/bin/bash

# Build script for k8s-http-fake-operator Docker image
# This script automates the Docker image build process

set -e

# Configuration
IMAGE_NAME="${IMAGE_NAME:-k8s-http-fake-operator}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
DOCKERFILE="${DOCKERFILE:-build/Dockerfile}"
BUILD_CONTEXT="${BUILD_CONTEXT:-..}"
NO_CACHE="${NO_CACHE:-false}"
SKIP_BUILD="${SKIP_BUILD:-false}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Print usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Build Docker image for k8s-http-fake-operator

OPTIONS:
    -n, --name IMAGE_NAME       Docker image name (default: k8s-http-fake-operator)
    -t, --tag IMAGE_TAG         Docker image tag (default: latest)
    -f, --dockerfile DOCKERFILE Path to Dockerfile (default: Dockerfile)
    -c, --context BUILD_CONTEXT Build context directory (default: .)
    --no-cache                  Build without using cache
    --skip-build                Skip binary build step (use existing binary)
    -h, --help                  Show this help message

ENVIRONMENT VARIABLES:
    IMAGE_NAME                  Docker image name
    IMAGE_TAG                   Docker image tag
    DOCKERFILE                  Path to Dockerfile
    BUILD_CONTEXT               Build context directory
    NO_CACHE                    Build without using cache (true/false)
    SKIP_BUILD                  Skip binary build step (true/false)

EXAMPLES:
    # Build with default settings
    $0

    # Build with custom name and tag
    $0 --name my-operator --tag v1.0.0

    # Build without cache
    $0 --no-cache

    # Build with custom Dockerfile
    $0 --dockerfile Dockerfile.custom

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--name)
            IMAGE_NAME="$2"
            shift 2
            ;;
        -t|--tag)
            IMAGE_TAG="$2"
            shift 2
            ;;
        -f|--dockerfile)
            DOCKERFILE="$2"
            shift 2
            ;;
        -c|--context)
            BUILD_CONTEXT="$2"
            shift 2
            ;;
        --no-cache)
            NO_CACHE="true"
            shift
            ;;
        --skip-build)
            SKIP_BUILD="true"
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Main build process
main() {
    log_info "Starting Docker image build process..."
    log_info "Image name: $IMAGE_NAME"
    log_info "Image tag: $IMAGE_TAG"
    log_info "Dockerfile: $DOCKERFILE"
    log_info "Build context: $BUILD_CONTEXT"
    log_info "No cache: $NO_CACHE"

    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi

    # Check if Dockerfile exists
    if [ ! -f "$BUILD_CONTEXT/$DOCKERFILE" ]; then
        log_error "Dockerfile not found: $BUILD_CONTEXT/$DOCKERFILE"
        exit 1
    fi

    # Build binary (unless skipped)
    if [ "$SKIP_BUILD" = "false" ]; then
        log_info "Building binary..."
        if ! go build -o "manager" "../cmd/main.go"; then
            log_error "Binary build failed!"
            exit 1
        fi
        log_info "Binary built successfully!"
    else
        log_info "Skipping binary build step..."
        if [ ! -f "manager" ]; then
            log_error "Binary not found: manager"
            log_error "Please build the binary first or remove --skip-build flag"
            exit 1
        fi
        log_info "Using existing binary: manager"
    fi

    # Build Docker image
    log_info "Building Docker image..."

    BUILD_CMD="docker build --pull=false"

    if [ "$NO_CACHE" = "true" ]; then
        BUILD_CMD="$BUILD_CMD --no-cache"
    fi

    BUILD_CMD="$BUILD_CMD -t $IMAGE_NAME:$IMAGE_TAG"
    BUILD_CMD="$BUILD_CMD -f $BUILD_CONTEXT/$DOCKERFILE"
    BUILD_CMD="$BUILD_CMD $BUILD_CONTEXT"
    
    log_info "Executing: $BUILD_CMD"
    
    if $BUILD_CMD; then
        log_info "Docker image built successfully!"
        log_info "Image: $IMAGE_NAME:$IMAGE_TAG"
        
        # Show image size
        IMAGE_SIZE=$(docker images $IMAGE_NAME:$IMAGE_TAG --format "{{.Size}}")
        log_info "Image size: $IMAGE_SIZE"
        
        # List the image
        docker images $IMAGE_NAME:$IMAGE_TAG
        
        log_info "You can now push the image to your registry:"
        log_info "  docker tag $IMAGE_NAME:$IMAGE_TAG <registry>/$IMAGE_NAME:$IMAGE_TAG"
        log_info "  docker push <registry>/$IMAGE_NAME:$IMAGE_TAG"
        
    else
        log_error "Docker image build failed!"
        exit 1
    fi
}

# Run main function
main