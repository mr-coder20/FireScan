package scanner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mr-coder20/FireScan/internal/config"
)

type NaabuEngine struct{}

func (e *NaabuEngine) Name() string { return "naabu" }

func (e *NaabuEngine) Scan(cfg *config.Config) (*ScanResult, error) {
	result := &ScanResult{
		Target:    cfg.Target,
		StartTime: time.Now().Format(time.RFC3339),
		Engine:    "naabu",
	}

	startTime := time.Now()

	if !cfg.Quiet {
		log.Printf("[Naabu] Scanning %s ports=%s", cfg.Target, cfg.Ports)
	}

	args := []string{
		"-host", cfg.Target,
		"-p", cfg.Ports,
		"-rate", strconv.Itoa(cfg.Rate),
		"-silent",
		"-json",
	}

	cmd := exec.Command("naabu", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("naabu: %w\n%s", err, stderr.String())
	}

	openPorts := make(map[int]bool)
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var data struct {
			Port   int    `json:"port"`
			IP     string `json:"ip"`
			Protocol string `json:"protocol"`
		}
		if err := json.Unmarshal([]byte(line), &data); err == nil && data.Port > 0 {
			openPorts[data.Port] = true
		}
	}

	ports := make([]int, 0, len(openPorts))
	for p := range openPorts {
		ports = append(ports, p)
	}
	sort.Ints(ports)

	host := HostResult{IP: cfg.Target, Status: "up"}
	for _, p := range ports {
		host.Ports = append(host.Ports, PortInfo{
			Port: p, Protocol: "tcp", State: "open",
		})
	}
	host.OpenCount = len(host.Ports)
	result.TotalOpen = len(host.Ports)
	result.Hosts = append(result.Hosts, host)

	result.EndTime = time.Now().Format(time.RFC3339)
	result.DurationSec = time.Since(startTime).Seconds()
	result.Duration = fmt.Sprintf("%.1fs", result.DurationSec)

	if !cfg.Quiet {
		log.Printf("[Naabu] Done: %d open ports in %s", result.TotalOpen, result.Duration)
	}

	return result, nil
}