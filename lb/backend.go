package lb

import (
	"log"
	"net"
	"time"
)

type Server struct {
	Addr    string
	IsAlive bool
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