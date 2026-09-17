package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"layer4/config"
	"layer4/lb"
)

func main() {
	
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("❌ Failed to load config file: %v", err)
	}



	interval, err := time.ParseDuration(cfg.Server.HealthCheckInterval)
	if err != nil {
		log.Fatalf("❌ Health check interval parsing error: %v", err)
	}

	loadBalancer := lb.NewLoadBalancer(cfg.Backends)
	loadBalancer.StartHealthCheck(interval)

	
	
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
		go loadBalancer.HandleConnection(conn)
	}
}