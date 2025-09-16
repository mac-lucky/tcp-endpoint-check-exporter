# TCP Endpoint Check Exporter

[![Docker Pulls](https://img.shields.io/docker/pulls/maclucky/tcp-endpoint-check-exporter)](https://hub.docker.com/r/maclucky/tcp-endpoint-check-exporter)
[![Docker Image Version](https://img.shields.io/docker/v/maclucky/tcp-endpoint-check-exporter/latest)](https://hub.docker.com/r/maclucky/tcp-endpoint-check-exporter/tags)
[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/mac-lucky/tcp-endpoint-check-exporter/ci-cd.yml)](https://github.com/mac-lucky/tcp-endpoint-check-exporter/actions)

A Prometheus exporter that monitors TCP endpoint connectivity and exposes metrics about their availability.

## Features

- TCP connection checks for multiple endpoints
- Prometheus metrics exposition
- Configurable check intervals
- YAML-based configuration
- Concurrent checks for better performance
- Docker support

## Configuration

### Environment Variables

- `CHECK_INTERVAL_SECONDS`: Time between checks (default: 30)
- `METRICS_PORT`: Port to expose Prometheus metrics (default: 2112)

### Config File

```yaml
targets:
  - host: "google.com"
    port: 443
    env: "prod1"
  - host: "example.com"
    port: 80
    env: "staging"
```

If no configuration file is provided, the exporter defaults to checking `google.com:443`.

## Metrics

The exporter provides the following metric:

- `tcp_endpoint_up`: TCP endpoint connectivity status (1 for up, 0 for down)
  - Labels:
    - `host`: Target hostname or IP
    - `port`: Target port
    - `env`: Environment label (optional)
    - `alias`: alias for host (optional)

## Docker Usage

### Local Build

1. Create a `config.yml` in the current directory. Based on Config File section

2. Build the container:
```bash
# Quick build (uses default Go version)
docker build -t tcp-endpoint-check-exporter .

# Exact build (uses Go version from go.mod - recommended)
docker build --build-arg GO_VERSION=$(grep "^go " go.mod | cut -d' ' -f2) -t tcp-endpoint-check-exporter .

# Development build (includes debugging tools)
docker build --target development -t tcp-endpoint-check-exporter:dev .
```

3. Run the container:
```bash
docker run -d \
  --name tcp-endpoint-check-exporter \
  -p 2112:2112 \
  -v $(pwd)/config.yml:/config/config.yml \
  -e CHECK_INTERVAL_SECONDS=30 \
  -e METRICS_PORT=2112 \
  tcp-endpoint-check-exporter
```

### Pre-built Images

There are pre-built images available on Docker Hub and GitHub Container Registry:

```bash
# Docker Hub
docker run -d \
  --name tcp-endpoint-check-exporter \
  -p 2112:2112 \
  -v $(pwd)/config.yml:/config/config.yml \
  -e CHECK_INTERVAL_SECONDS=30 \
  -e METRICS_PORT=2112 \
  maclucky/tcp-endpoint-check-exporter:latest

# GitHub Container Registry
docker run -d \
  --name tcp-endpoint-check-exporter \
  -p 2112:2112 \
  -v $(pwd)/config.yml:/config/config.yml \
  -e CHECK_INTERVAL_SECONDS=30 \
  -e METRICS_PORT=2112 \
  ghcr.io/mac-lucky/tcp-endpoint-check-exporter:latest
```

## Prometheus Configuration

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'tcp-endpoint-check'
    static_configs:
      - targets: ['localhost:2112']
```

## Building from Source

### Go Binary
```bash
go mod download
go build -o tcp_endpoint_check_exporter
./tcp_endpoint_check_exporter
```

### Multi-platform Docker Build
```bash
# Build for multiple architectures
docker buildx build --platform linux/amd64,linux/arm64 -t tcp-endpoint-check-exporter .
```

## Development

### Quick Development Workflow
```bash
# 1. Build and test locally
docker build --build-arg GO_VERSION=$(grep "^go " go.mod | cut -d' ' -f2) -t tcp-endpoint-check-exporter .

# 2. Run with test config
docker run -d -p 2112:2112 -v $(pwd)/config.yml:/config/config.yml tcp-endpoint-check-exporter

# 3. Check metrics
curl http://localhost:2112/metrics

# 4. View logs
docker logs tcp-endpoint-check-exporter

# 5. Clean up
docker stop tcp-endpoint-check-exporter && docker rm tcp-endpoint-check-exporter
```

### Development Container
```bash
# Build development image with debugging tools
docker build --target development -t tcp-endpoint-check-exporter:dev .

# Run with shell access
docker run -it --rm -p 2112:2112 -v $(pwd):/workspace tcp-endpoint-check-exporter:dev /bin/sh
```

