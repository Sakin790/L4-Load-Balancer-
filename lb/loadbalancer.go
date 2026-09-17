package lb

import (
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

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

func (lb *LoadBalancer) HandleConnection(src net.Conn) {
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