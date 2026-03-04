package service

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

type PortResult struct {
	Port    int
	Open    bool
	Latency time.Duration
}

type Scanner struct {
	timeout time.Duration
}

func NewScanner(timeout time.Duration) *Scanner {
	if timeout == 0 {
		timeout = 2 * time.Second
	}
	return &Scanner{timeout: timeout}
}

// ScanPort checks if a single TCP port is open on the given host.
func (s *Scanner) ScanPort(host string, port int) PortResult {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	start := time.Now()

	conn, err := net.DialTimeout("tcp", addr, s.timeout)
	latency := time.Since(start)

	if err != nil {
		return PortResult{Port: port, Open: false, Latency: latency}
	}
	conn.Close()

	return PortResult{Port: port, Open: true, Latency: latency}
}

// ScanPorts checks multiple TCP ports concurrently and returns results for open ports.
func (s *Scanner) ScanPorts(host string, ports []int) []PortResult {
	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		results []PortResult
	)

	// Limit concurrency to avoid overwhelming the system
	sem := make(chan struct{}, 50)

	for _, port := range ports {
		wg.Add(1)
		sem <- struct{}{}

		go func(p int) {
			defer wg.Done()
			defer func() { <-sem }()

			result := s.ScanPort(host, p)
			if result.Open {
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			}
		}(port)
	}

	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		return results[i].Port < results[j].Port
	})

	return results
}

// ScanCommonPorts scans a predefined list of common development ports.
func (s *Scanner) ScanCommonPorts(host string) []PortResult {
	commonPorts := []int{
		80, 443, 1433, 2181, 3000, 3001, 3306, 4000, 4200, 4443,
		5000, 5173, 5432, 5672, 6379, 6380, 7474, 8000, 8080, 8081,
		8443, 8888, 9000, 9090, 9092, 9200, 9300, 15672, 27017,
	}
	return s.ScanPorts(host, commonPorts)
}

// CheckHealth performs an HTTP health check on a given host:port/path.
func CheckHealth(host string, port int, healthPath string, timeout time.Duration) (bool, time.Duration) {
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	start := time.Now()

	conn, err := net.DialTimeout("tcp", addr, timeout)
	latency := time.Since(start)

	if err != nil {
		return false, latency
	}

	// Send a minimal HTTP GET request
	request := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", healthPath, host)
	_ = conn.SetDeadline(time.Now().Add(timeout))
	_, err = conn.Write([]byte(request))
	if err != nil {
		conn.Close()
		return false, time.Since(start)
	}

	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	conn.Close()

	if err != nil || n == 0 {
		return false, time.Since(start)
	}

	response := string(buf[:n])
	// Check for 2xx status codes
	return len(response) > 12 && (response[9] == '2'), time.Since(start)
}
