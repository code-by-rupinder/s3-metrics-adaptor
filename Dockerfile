FROM golang:1.23-alpine AS builder

# Build arguments for version information
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown
ARG GIT_BRANCH=unknown

# Set necessary environment variables for static compilation
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GO111MODULE=on \
    GOCACHE=/go-cache \
    GOMODCACHE=/go-mod-cache

# Set working directory
WORKDIR /app

# Install minimal build dependencies
RUN apk add --no-cache ca-certificates

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go-mod-cache \
    go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN --mount=type=cache,target=/go-cache \
    --mount=type=cache,target=/go-mod-cache \
    go build -trimpath \
    -ldflags="-s -w \
    -X main.Version=${VERSION} \
    -X main.BuildTime=${BUILD_TIME} \
    -X main.GitCommit=${GIT_COMMIT} \
    -X main.GitBranch=${GIT_BRANCH}" \
    -o s3_metrics_adapter ./cmd

# Create the final minimal image using distroless for better security
FROM gcr.io/distroless/static:nonroot

# Copy build args to use in final image
ARG VERSION=dev
ARG GIT_COMMIT=unknown

# Add version labels
LABEL version="${VERSION}"
LABEL maintainer="codebyrupinder"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.revision="${GIT_COMMIT}"
LABEL org.opencontainers.image.title="S3 Metrics Adapter"
LABEL org.opencontainers.image.description="S3 Metrics Adapter - converts S3 events to Prometheus metrics"
LABEL org.opencontainers.image.authors="codebyrupinder"
LABEL org.opencontainers.image.vendor="codebyrupinder"

# Set working directory
WORKDIR /app

# Copy binary and config from builder
COPY --from=builder /app/s3_metrics_adapter /usr/local/bin/s3_metrics_adapter
COPY --from=builder /app/config.yaml /app/config.yaml

# Expose metrics port
EXPOSE 8087

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/s3_metrics_adapter"]
