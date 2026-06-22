package scanner

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mr-coder20/FireScan/internal/config"
)

type MasscanEngine struct{}

func (e *MasscanEngine) Name() string { return "masscan" }

func (e *MasscanEngine) Scan(cfg *config.Config) (*ScanResult, error) {
	result := &ScanResult{
		Target:    cfg.Target,
		StartTime: time.Now().Format(time.RFC3339),
		Engine:    "masscan",
	}

	if !cfg.Quiet {
		log.Printf("[Masscan] Scanning %s ports=%s rate=%d", cfg.Target, cfg.Ports, cfg.Rate)
	}

	tmpFile, err := os.CreateTemp("", "firescan-*.txt")
	if err != nil {
		return nil, fmt.Errorf("temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	args := []string{
		cfg.Target, "-p", cfg.Ports,
		"--rate", strconv.Itoa(cfg.Rate),
		"--retries", strconv.Itoa(cfg.Retries),
		"-oL", tmpPath, "--wait", "3",
	}

	startTime := time.Now()
	cmd := exec.Command("masscan", args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("masscan: %w", err)
	}

	hostPorts := make(map[string][]PortInfo)
	file, err := os.Open(tmpPath)
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if !strings.HasPrefix(line, "open") {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				port, _ := strconv.Atoi(parts[2])
				hostPorts[parts[3]] = append(hostPorts[parts[3]], PortInfo{
					Port: port, Protocol: parts[1], State: "open",
				})
			}
		}
	}

	for ip, ports := range hostPorts {
		sort.Slice(ports, func(i, j int) bool { return ports[i].Port < ports[j].Port })
		result.Hosts = append(result.Hosts, HostResult{
			IP: ip, Status: "up", Ports: ports, OpenCount: len(ports),
		})
		result.TotalOpen += len(ports)
	}

	result.EndTime = time.Now().Format(time.RFC3339)
	result.DurationSec = time.Since(startTime).Seconds()
	result.Duration = fmt.Sprintf("%.1fs", result.DurationSec)

	if !cfg.Quiet {
		log.Printf("[Masscan] Done: %d hosts, %d open ports in %s",
			len(result.Hosts), result.TotalOpen, result.Duration)
	}

	return result, nil
}
