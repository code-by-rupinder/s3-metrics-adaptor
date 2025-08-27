#!/bin/bash

# Exit on any error
set -e

# Configuration
APP_NAME="s3-event-exporter"
BUILD_DIR="bin"
MAIN_PATH="cmd/main.go"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building ${APP_NAME}...${NC}"
echo "Version: ${VERSION}"
echo "Build Time: ${BUILD_TIME}"
echo "Commit: ${COMMIT_HASH}"

# Create build directory if it doesn't exist
mkdir -p "${BUILD_DIR}"

# Run tests
echo -e "\n${GREEN}Running tests...${NC}"
go test -v ./... || {
    echo -e "${RED}Tests failed${NC}"
    exit 1
}

# Build the application
echo -e "\n${GREEN}Building binary...${NC}"
CGO_ENABLED=0 go build \
    -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${COMMIT_HASH} -X main.GitBranch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo unknown)" \
    -o "${BUILD_DIR}/${APP_NAME}" \
    "${MAIN_PATH}"

# Check if build was successful
if [ $? -eq 0 ]; then
    echo -e "\n${GREEN}Build successful!${NC}"
    echo "Binary location: ${BUILD_DIR}/${APP_NAME}"
    # Show binary size
    ls -lh "${BUILD_DIR}/${APP_NAME}"
else
    echo -e "\n${RED}Build failed${NC}"
    exit 1
fi

# Optional: copy config file to bin directory
if [ -f "cmd/config.yaml" ]; then
    echo -e "\n${GREEN}Copying config file to bin directory...${NC}"
    cp cmd/config.yaml "${BUILD_DIR}/"
fi

