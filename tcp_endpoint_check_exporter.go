package main

import (
    "fmt"
    "log"
    "net"
    "net/http"
    "os"
    "strconv"
    "time"
    "sync"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "gopkg.in/yaml.v2"
)

type Target struct {
    Host  string `yaml:"host"`
    Port  int    `yaml:"port"`
    Env   string `yaml:"env,omitempty"`
    Alias string `yaml:"alias,omitempty"`
}

type Config struct {
    Targets []Target `yaml:"targets"`
}

var (
    tcpEndpointUp = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "tcp_endpoint_up",
            Help: "TCP endpoint connectivity status (1 for up, 0 for down)",
        },
        []string{"host", "port", "env", "alias"},
    )
)

func init() {
    prometheus.MustRegister(tcpEndpointUp)
}

func loadConfig() []Target {
    log.Printf("Loading configuration from /config/config.yml")
    data, err := os.ReadFile("/config/config.yml")
    if err != nil {
        log.Printf("Warning: Could not read config file: %v. Using default target", err)
        return []Target{{Host: "google.com", Port: 443}}
    }

    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        log.Printf("Warning: Could not parse config file: %v. Using default target", err)
        return []Target{{Host: "google.com", Port: 443}}
    }

    if len(config.Targets) == 0 {
        log.Printf("Warning: No targets found in config. Using default target")
        return []Target{{Host: "google.com", Port: 443}}
    }

    log.Printf("Loaded %d targets from configuration", len(config.Targets))
    return config.Targets
}

func checkEndpoint(target Target) {
    start := time.Now()
    address := fmt.Sprintf("%s:%d", target.Host, target.Port)
    alias := target.Alias
    if alias == "" {
        alias = target.Host
    }
    env := target.Env
    if env == "" {
        env = "default"
    }
    
    log.Printf("Checking endpoint %s (env: %s, alias: %s)", address, env, alias)
    
    conn, err := net.DialTimeout("tcp", address, 5*time.Second)
    duration := time.Since(start)
    
    if err != nil {
        log.Printf("❌ Failed to connect to %s: %v (took %v)", address, err, duration)
        tcpEndpointUp.WithLabelValues(target.Host, strconv.Itoa(target.Port), env, alias).Set(0)
        return
    }
    
    conn.Close()
    successMsg := fmt.Sprintf("✅ Successfully connected to %s (took %v)", address, duration)
    if env != "default" || alias != target.Host {
        successMsg = fmt.Sprintf("✅ Successfully connected to %s [env: %s, alias: %s] (took %v)", address, env, alias, duration)
    }
    log.Printf("%s", successMsg)
    tcpEndpointUp.WithLabelValues(target.Host, strconv.Itoa(target.Port), env, alias).Set(1)
}

func main() {
    // TCP Endpoint Check Exporter - monitors endpoint connectivity and reports detailed metrics
    log.Printf("Starting TCP Endpoint Check Exporter")
    
    // Get check interval from environment
    intervalStr := os.Getenv("CHECK_INTERVAL_SECONDS")
    interval := 30 // default interval
    if i, err := strconv.Atoi(intervalStr); err == nil && i > 0 {
        interval = i
    }
    log.Printf("Check interval set to %d seconds", interval)

    // Get metrics port from environment
    metricsPort := os.Getenv("METRICS_PORT")
    if metricsPort == "" {
        metricsPort = "2112"
    }
    log.Printf("Metrics port set to %s", metricsPort)

    targets := loadConfig()

    // Start periodic checks
    go func() {
        for {
            log.Printf("Starting concurrent checks for %d targets", len(targets))
            start := time.Now()
            var wg sync.WaitGroup
            for _, target := range targets {
                wg.Add(1)
                go func(t Target) {
                    defer wg.Done()
                    checkEndpoint(t)
                }(target)
            }
            wg.Wait()
            duration := time.Since(start)
            log.Printf("Completed all checks in %v", duration)
            time.Sleep(time.Duration(interval) * time.Second)
        }
    }()

    // Expose metrics endpoint
    http.Handle("/metrics", promhttp.Handler())
    log.Printf("Starting metrics server on :%s", metricsPort)
    log.Printf("Metrics available at: \033[34mhttp://localhost:%s/metrics\033[0m", metricsPort)
    if err := http.ListenAndServe(":"+metricsPort, nil); err != nil {
        log.Fatal(err)
    }
}
