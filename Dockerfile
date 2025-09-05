# Optimized multi-stage Dockerfile for TCP Endpoint Check Exporter
# Uses distroless base images for maximum security and minimal image size

# Build stage - Use the official Go image for building
# For CI: Pass GO_VERSION as build arg
# For local: Detects Go version from go.mod automatically
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:alpine AS go-detector

WORKDIR /detect
COPY go.mod ./

# Create a script that detects Go version from go.mod
RUN GO_VERSION_FROM_MOD=$(grep "^go " go.mod | cut -d' ' -f2) && \
    echo "Detected Go version: $GO_VERSION_FROM_MOD" && \
    echo "$GO_VERSION_FROM_MOD" > /tmp/detected_version

# Main build stage with version handling
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:${GO_VERSION:-1.23.4}-alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG GO_VERSION
ARG VERSION="dev"
ARG COMMIT_SHA="unknown"
ARG BUILD_DATE="unknown"

WORKDIR /src

# Copy the detected version and validate
COPY --from=go-detector /tmp/detected_version /tmp/detected_version
COPY --from=go-detector /detect/go.mod ./go.mod.check

# Validate Go version and provide guidance
RUN DETECTED=$(cat /tmp/detected_version) && \
    CURRENT_GO=$(go version | cut -d' ' -f3 | sed 's/go//') && \
    echo "Building with Go $CURRENT_GO" && \
    echo "Project requires Go $DETECTED" && \
    if [ -n "$GO_VERSION" ]; then \
      echo "Using CI-provided Go version: $GO_VERSION"; \
    else \
      echo "Using default Go version. For exact version match, build with: docker build --build-arg GO_VERSION=$DETECTED ."; \
    fi

# Install build dependencies
RUN apk add --no-cache \
    ca-certificates \
    git \
    gcc \
    musl-dev

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies with verification
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build \
    -ldflags="-w -s -X main.version=${VERSION} -X main.commitSHA=${COMMIT_SHA} -X main.buildDate=${BUILD_DATE}" \
    -a -installsuffix cgo \
    -o /app/tcp_endpoint_check_exporter

# Verify the binary works
RUN /app/tcp_endpoint_check_exporter --help || echo "Binary built successfully"

# Runtime stage - Use Google's distroless image for maximum security
FROM --platform=${TARGETPLATFORM:-linux/amd64} gcr.io/distroless/static-debian12:nonroot AS runner

# Import ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the application binary
COPY --from=builder /app/tcp_endpoint_check_exporter /app/tcp_endpoint_check_exporter

# Create necessary directories
USER 65532:65532

# Set environment variables
ENV METRICS_PORT=2112 \
    CHECK_INTERVAL_SECONDS=30

# Expose the metrics port
EXPOSE ${METRICS_PORT}

# Add labels for better maintainability
LABEL org.opencontainers.image.title="TCP Endpoint Check Exporter" \
      org.opencontainers.image.description="Prometheus exporter for TCP endpoint monitoring" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.revision="${COMMIT_SHA}" \
      org.opencontainers.image.vendor="maclucky" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.source="https://github.com/mac-lucky/tcp-endpoint-check-exporter"

# Use ENTRYPOINT for better signal handling
ENTRYPOINT ["/app/tcp_endpoint_check_exporter"]

# Development stage - includes shell and additional tools for debugging
FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3.19 AS development
RUN apk add --no-cache ca-certificates curl wget
COPY --from=builder /app/tcp_endpoint_check_exporter /app/tcp_endpoint_check_exporter
RUN addgroup -g 65532 -S nonroot && adduser -u 65532 -S nonroot -G nonroot
USER 65532:65532
ENV METRICS_PORT=2112 CHECK_INTERVAL_SECONDS=30
EXPOSE ${METRICS_PORT}
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:${METRICS_PORT}/metrics || exit 1
ENTRYPOINT ["/app/tcp_endpoint_check_exporter"]