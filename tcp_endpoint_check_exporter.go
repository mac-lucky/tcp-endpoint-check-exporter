package main

import (
    "fmt"
    "os"
    "net"
    "net/http"
    "time"
    "strings"
    "strconv"
    "log"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "gopkg.in/yaml.v2"
)

type Config struct {
    Targets []struct {
        Host string `yaml:"host"`
        Port int    `yaml:"port"`
        Env  string `yaml:"env,omitempty"`
    } `yaml:"targets"`
}

var (
    connectionStatus = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "host_port_connection_status",
            Help: "Connection status to host:port (1 = success, 0 = failure)",
        },
        []string{"host", "port", "env"},
    )
)

func init() {
    prometheus.MustRegister(connectionStatus)
}

func checkConnection(host string, port int) bool {
    timeout := 5 * time.Second
    target := fmt.Sprintf("%s:%d", host, port)
    conn, err := net.DialTimeout("tcp", target, timeout)
    if err != nil {
        log.Printf("Failed to connect to %s: %v", target, err)
        return false
    }
    defer conn.Close()
    log.Printf("Successfully connected to %s", target)
    return true
}

func main() {
    log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
    log.Println("Starting host-port checker...")

    var config Config

    // Read config file from mounted location
    data, err := os.ReadFile("/config/config.yml")
    if (err != nil) {
        log.Printf("Warning: Cannot find config file: %v\n", err)
        log.Println("Falling back to default check: google.com:443")
        config = Config{
            Targets: []struct {
                Host string `yaml:"host"`
                Port int    `yaml:"port"`
                Env  string `yaml:"env,omitempty"`
            }{
                {
                    Host: "google.com",
                    Port: 443,
                },
            },
        }
    } else {
        log.Println("Successfully loaded config file")
        err = yaml.Unmarshal(data, &config)
        if err != nil {
            log.Fatalf("Failed to parse config: %v", err)
        }
    }

    // Get check interval from environment variable
    checkInterval := 30 // default 30 seconds
    if interval := os.Getenv("CHECK_INTERVAL_SECONDS"); interval != "" {
        if i, err := strconv.Atoi(interval); err == nil && i > 0 {
            checkInterval = i
        }
    }
    log.Printf("Check interval set to %d seconds", checkInterval)

    var config Config

    // Read config file from mounted location
    data, err := os.ReadFile("/config/config.yml")
    if (err != nil) {
        log.Printf("Warning: Cannot find config file: %v\n", err)
        log.Println("Falling back to default check: google.com:443")
        config = Config{
            Targets: []struct {
                Host string `yaml:"host"`
                Port int    `yaml:"port"`
                Env  string `yaml:"env,omitempty"`
            }{
                {
                    Host: "google.com",
                    Port: 443,
                },
            },
        }
    } else {
        log.Println("Successfully loaded config file")
        err = yaml.Unmarshal(data, &config)
        if err != nil {
            log.Fatalf("Failed to parse config: %v", err)
        }
    }

    // Get check interval from environment variable
    checkInterval := 30 // default 30 seconds
    if interval := os.Getenv("CHECK_INTERVAL_SECONDS"); interval != "" {
        if i, err := strconv.Atoi(interval); err == nil && i > 0 {
            checkInterval = i
        }    }
    log.Printf("Check interval set to %d seconds", checkInterval)

    // Start periodic checks
    go func() {
        log.Println("Starting periodic checks...")
        for {
            checkAllTargets(config.Targets)
            time.Sleep(time.Duration(checkInterval) * time.Second)
        }
    }()

    // Get metrics port from environment variable
    metricsPort := os.Getenv("METRICS_PORT")
    if metricsPort == "" {
        metricsPort = "2112" // default port
    }
    if !strings.HasPrefix(metricsPort, ":") {
        metricsPort = ":" + metricsPort
    }

    log.Printf("Starting metrics server on port%s", metricsPort)
    // Expose metrics endpoint
    http.Handle("/metrics", promhttp.Handler())
    if err := http.ListenAndServe(metricsPort, nil); err != nil {
        log.Fatalf("Failed to start metrics server: %v", err)
    }
}
