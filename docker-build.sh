#!/bin/bash

# Docker Build and Push Script for S3 Event Exporter
# This script builds the Docker image with proper versioning and pushes to Docker Hub

set -e

# Default values
DOCKER_REGISTRY="docker.io"
DOCKER_REPOSITORY="codebyrupinder/s3-event-exporter"
DEFAULT_VERSION="latest"
PLATFORMS="linux/amd64,linux/arm64"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Build and optionally push Docker images for S3 Event Exporter

OPTIONS:
    -v, --version VERSION     Set the version tag (default: latest)
    -p, --push               Push to Docker Hub after building
    -m, --multi-platform     Build for multiple platforms (linux/amd64,linux/arm64)
    -r, --repository REPO    Docker repository (default: codebyrupinder/s3-event-exporter)
    -h, --help               Show this help message

EXAMPLES:
    $0 -v 1.0.0              Build version 1.0.0 locally
    $0 -v 1.0.0 -p           Build and push version 1.0.0
    $0 -v 1.0.0 -p -m        Build and push multi-platform version 1.0.0
    $0 -p                    Build and push latest

ENVIRONMENT VARIABLES:
    DOCKER_USERNAME          Docker Hub username (required for push)
    DOCKER_PASSWORD          Docker Hub password/token (required for push)

EOF
}

# Parse command line arguments
VERSION="$DEFAULT_VERSION"
PUSH=false
MULTI_PLATFORM=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -p|--push)
            PUSH=true
            shift
            ;;
        -m|--multi-platform)
            MULTI_PLATFORM=true
            shift
            ;;
        -r|--repository)
            DOCKER_REPOSITORY="$2"
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Validate prerequisites
print_info "Checking prerequisites..."

# Check if Docker is installed and running
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed or not in PATH"
    exit 1
fi

if ! docker info &> /dev/null; then
    print_error "Docker daemon is not running"
    exit 1
fi

# Check if we're in the right directory
if [[ ! -f "Dockerfile" ]]; then
    print_error "Dockerfile not found. Please run this script from the project root directory."
    exit 1
fi

if [[ ! -f "go.mod" ]]; then
    print_error "go.mod not found. Please run this script from the project root directory."
    exit 1
fi

# Generate build metadata
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
if git rev-parse --git-dir > /dev/null 2>&1; then
    GIT_COMMIT=$(git rev-parse --short HEAD)
    GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
else
    GIT_COMMIT="unknown"
    GIT_BRANCH="main"
fi

print_info "Build metadata:"
echo "  Version: $VERSION"
echo "  Build time: $BUILD_TIME"
echo "  Git commit: $GIT_COMMIT"
echo "  Git branch: $GIT_BRANCH"
echo "  Repository: $DOCKER_REPOSITORY"

# Build arguments
BUILD_ARGS="--build-arg VERSION=${VERSION} --build-arg BUILD_TIME=${BUILD_TIME} --build-arg GIT_COMMIT=${GIT_COMMIT} --build-arg GIT_BRANCH=${GIT_BRANCH}"

# Tags to build
TAGS="-t ${DOCKER_REPOSITORY}:${VERSION}"
if [[ "$VERSION" != "latest" ]]; then
    TAGS="$TAGS -t ${DOCKER_REPOSITORY}:latest"
fi

# Build the image
print_info "Building Docker image..."

if [[ "$MULTI_PLATFORM" == true ]]; then
    print_info "Building for multiple platforms: $PLATFORMS"
    
    # Create buildx builder if not exists
    if ! docker buildx inspect multiplatform >/dev/null 2>&1; then
        print_info "Creating buildx builder..."
        docker buildx create --name multiplatform --driver docker-container --use
        docker buildx inspect --bootstrap
    else
        docker buildx use multiplatform
    fi
    
    if [[ "$PUSH" == true ]]; then
        docker buildx build --platform $PLATFORMS $BUILD_ARGS $TAGS --push .
    else
        print_warning "Multi-platform builds require --push flag to work properly"
        docker buildx build --platform $PLATFORMS $BUILD_ARGS $TAGS --load .
    fi
else
    # Single platform build
    docker build $BUILD_ARGS $TAGS .
fi

print_success "Docker image built successfully!"

# Show built images
print_info "Built images:"
docker images $DOCKER_REPOSITORY --format "table {{.Repository}}\t{{.Tag}}\t{{.ID}}\t{{.Size}}\t{{.CreatedSince}}"

# Test the image
print_info "Testing the built image..."
if docker run --rm "${DOCKER_REPOSITORY}:${VERSION}" -version; then
    print_success "Image test passed!"
else
    print_error "Image test failed!"
    exit 1
fi

# Push to Docker Hub if requested
if [[ "$PUSH" == true ]]; then
    print_info "Pushing to Docker Hub..."
    
    # Check credentials
    if [[ -z "$DOCKER_USERNAME" ]] || [[ -z "$DOCKER_PASSWORD" ]]; then
        print_warning "Docker credentials not found in environment variables."
        print_info "Please log in to Docker Hub:"
        docker login
    else
        print_info "Logging in to Docker Hub..."
        echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
    fi
    
    if [[ "$MULTI_PLATFORM" != true ]]; then
        # Push each tag
        docker push "${DOCKER_REPOSITORY}:${VERSION}"
        if [[ "$VERSION" != "latest" ]]; then
            docker push "${DOCKER_REPOSITORY}:latest"
        fi
    fi
    
    print_success "Images pushed successfully!"
    print_info "Published images:"
    echo "  ${DOCKER_REPOSITORY}:${VERSION}"
    if [[ "$VERSION" != "latest" ]]; then
        echo "  ${DOCKER_REPOSITORY}:latest"
    fi
fi

# Summary
print_success "Build completed successfully!"
echo ""
echo "To run the image:"
echo "  docker run -d -p 8087:8087 ${DOCKER_REPOSITORY}:${VERSION}"
echo ""
echo "To pull from Docker Hub:"
echo "  docker pull ${DOCKER_REPOSITORY}:${VERSION}"
echo ""
echo "Documentation: ./DOCKER_DEPLOYMENT_GUIDE.md"
