package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"
)

// Config struct to map the YAML file
type Config struct {
	Server struct {
		Port                string `yaml:"port"`
		HealthCheckInterval string `yaml:"health_check_interval"`
	} `yaml:"server"`
	Backends []string `yaml:"backends"`
}

// LoadConfig reads and parses the YAML configuration file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

type Server struct {
	Addr    string
	IsAlive bool
}

type LoadBalancer struct {
	servers []*Server
	counter uint64
	mu      sync.RWMutex
}

func NewLoadBalancer(addrs []string) *LoadBalancer {
	var servers []*Server
	for _, addr := range addrs {
		servers = append(servers, &Server{Addr: addr, IsAlive: true})
	}
	return &LoadBalancer{servers: servers}
}

func (lb *LoadBalancer) StartHealthCheck(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		lb.checkServers()
		for range ticker.C {
			lb.checkServers()
		}
	}()
}

func (lb *LoadBalancer) checkServers() {
	for _, s := range lb.servers {
		conn, err := net.DialTimeout("tcp", s.Addr, 2*time.Second)

		lb.mu.Lock()
		if err != nil {
			if s.IsAlive {
				log.Printf("[HealthCheck] ❌ Backend down: %s\n", s.Addr)
				s.IsAlive = false
			}
		} else {
			conn.Close()
			if !s.IsAlive {
				log.Printf("[HealthCheck] ✅ Backend recovered: %s\n", s.Addr)
				s.IsAlive = true
			}
		}
		lb.mu.Unlock()
	}
}

func (lb *LoadBalancer) getNextBackend() *Server {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	var healthyServers []*Server
	for _, s := range lb.servers {
		if s.IsAlive {
			healthyServers = append(healthyServers, s)
		}
	}

	if len(healthyServers) == 0 {
		return nil
	}

	idx := atomic.AddUint64(&lb.counter, 1)
	return healthyServers[idx%uint64(len(healthyServers))]
}

func (lb *LoadBalancer) handleConnection(src net.Conn) {
	defer src.Close()

	targetServer := lb.getNextBackend()
	if targetServer == nil {
		log.Println("❌ No healthy backend available! Connection closed.")
		return
	}

	dst, err := net.DialTimeout("tcp", targetServer.Addr, 3*time.Second)
	if err != nil {
		log.Printf("Failed to connect to backend %s: %v\n", targetServer.Addr, err)
		return
	}
	defer dst.Close()

	done := make(chan struct{}, 2)

	go func() {
		io.Copy(dst, src)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(src, dst)
		done <- struct{}{}
	}()

	<-done
}

func main() {
	// 1. Read configuration from YAML file
	cfg, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("❌ Failed to load config file: %v", err)
	}

	// Parse health check interval duration
	interval, err := time.ParseDuration(cfg.Server.HealthCheckInterval)
	if err != nil {
		log.Fatalf("❌ Health check interval parsing error: %v", err)
	}

	// 2. Initialize LoadBalancer with dynamic backends
	lb := NewLoadBalancer(cfg.Backends)
	lb.StartHealthCheck(interval)

	// 3. Listen on the port defined in YAML
	listener, err := net.Listen("tcp", cfg.Server.Port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.Server.Port, err)
	}
	defer listener.Close()

	fmt.Printf("🚀 L4 Load Balancer running on %s with Health Checking...\n", cfg.Server.Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go lb.handleConnection(conn)
	}
}