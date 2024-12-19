FROM golang:alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.* ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o tcp_endpoint_check_exporter

# Final stage
FROM alpine

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/tcp_endpoint_check_exporter .

# Create a directory for the mounted config
RUN mkdir /config

# Set default metrics port and check interval
ENV METRICS_PORT=2112
ENV CHECK_INTERVAL_SECONDS=30

# Expose prometheus metrics port
EXPOSE ${METRICS_PORT}

# Run the application with correct config path
CMD ["/app/tcp_endpoint_check_exporter"]
