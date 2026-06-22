package scanner

import (
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mr-coder20/FireScan/internal/config"
)

type NaiveEngine struct{}

func (e *NaiveEngine) Name() string { return "naive" }

func (e *NaiveEngine) Scan(cfg *config.Config) (*ScanResult, error) {
	result := &ScanResult{
		Target:    cfg.Target,
		StartTime: time.Now().Format(time.RFC3339),
		Engine:    "naive",
	}

	ports := parsePorts(cfg.Ports)
	total := len(ports)

	if !cfg.Quiet {
		log.Printf("[Naive] Scanning %s: %d ports (threads=%d)", cfg.Target, total, cfg.Threads)
	}

	startTime := time.Now()
	host := HostResult{IP: cfg.Target, Status: "up"}

	var scanned int32
	var mu sync.Mutex
	var wg sync.WaitGroup
	portChan := make(chan int, total)

	// Progress reporter
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s := atomic.LoadInt32(&scanned)
				pct := float64(s) * 100 / float64(total)
				if !cfg.Quiet {
					log.Printf("[Naive] %d/%d (%.1f%%)", s, total, pct)
				}
			case <-done:
				return
			}
		}
	}()

	// Start workers
	for i := 0; i < cfg.Threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range portChan {
				pi := scanTCPPort(cfg.Target, port, cfg.Timeout)
				atomic.AddInt32(&scanned, 1)
				if pi.State == "open" {
					mu.Lock()
					host.Ports = append(host.Ports, pi)
					mu.Unlock()
				}
			}
		}()
	}

	// Feed ports
	for _, p := range ports {
		portChan <- p
	}
	close(portChan)
	wg.Wait()
	close(done)

	sort.Slice(host.Ports, func(i, j int) bool {
		return host.Ports[i].Port < host.Ports[j].Port
	})
	host.OpenCount = len(host.Ports)
	result.Hosts = append(result.Hosts, host)
	result.TotalOpen = host.OpenCount

	result.EndTime = time.Now().Format(time.RFC3339)
	result.DurationSec = time.Since(startTime).Seconds()
	result.Duration = fmt.Sprintf("%.1fs", result.DurationSec)

	if !cfg.Quiet {
		log.Printf("[Naive] Done: %d ports scanned, %d open in %s",
			total, result.TotalOpen, result.Duration)
	}

	return result, nil
}

func scanTCPPort(target string, port, timeout int) PortInfo {
	pi := PortInfo{Port: port, Protocol: "tcp", State: "filtered"}
	addr := net.JoinHostPort(target, strconv.Itoa(port))

	conn, err := net.DialTimeout("tcp", addr, time.Duration(timeout)*time.Second)
	if err != nil {
		if strings.Contains(err.Error(), "refused") {
			pi.State = "closed"
		}
		return pi
	}
	defer conn.Close()

	pi.State = "open"
	conn.SetDeadline(time.Now().Add(2 * time.Second))

	buf := make([]byte, 256)
	n, _ := conn.Read(buf)
	if n > 0 {
		pi.Banner = strings.TrimSpace(string(buf[:n]))
	}

	// Simple service detection
	switch port {
	case 22:
		pi.Service = "ssh"
	case 21:
		pi.Service = "ftp"
	case 25:
		pi.Service = "smtp"
	case 53:
		pi.Service = "dns"
	case 80, 8080, 8000:
		pi.Service = "http"
	case 443, 8443:
		pi.Service = "https"
	case 110:
		pi.Service = "pop3"
	case 143:
		pi.Service = "imap"
	case 3306:
		pi.Service = "mysql"
	case 5432:
		pi.Service = "postgresql"
	case 6379:
		pi.Service = "redis"
	case 27017:
		pi.Service = "mongodb"
	case 3389:
		pi.Service = "rdp"
	case 5900:
		pi.Service = "vnc"
	default:
		if n > 0 {
			pi.Service = "unknown"
		}
	}

	return pi
}

func parsePorts(portSpec string) []int {
	var ports []int

	if portSpec == "1-65535" {
		ports = make([]int, 65535)
		for i := range ports {
			ports[i] = i + 1
		}
		return ports
	}

	for _, part := range strings.Split(portSpec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") {
			parts := strings.SplitN(part, "-", 2)
			start, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
			for p := start; p <= end; p++ {
				ports = append(ports, p)
			}
		} else {
			p, _ := strconv.Atoi(part)
			ports = append(ports, p)
		}
	}

	return ports
}
