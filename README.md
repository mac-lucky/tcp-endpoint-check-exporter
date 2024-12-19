# TCP Endpoint Check Exporter

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

Create a `config.yml` file in the `/config` directory:

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

## Docker Usage

1. Build the container:
```bash
docker build -t tcp-endpoint-check-exporter .
```

2. Run the container:
```bash
docker run -d \
  -p 2112:2112 \
  -v $(pwd)/config.yml:/config/config.yml \
  -e CHECK_INTERVAL_SECONDS=30 \
  -e METRICS_PORT=2112 \
  tcp-endpoint-check-exporter
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

```bash
go mod init tcp-endpoint-check-exporter
go mod tidy
go build -o tcp_endpoint_check_exporter
```

