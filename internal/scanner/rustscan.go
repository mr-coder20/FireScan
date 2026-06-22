package scanner

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mr-coder20/FireScan/internal/config"
)

type RustScanEngine struct{}

func (e *RustScanEngine) Name() string { return "rustscan" }

func (e *RustScanEngine) Scan(cfg *config.Config) (*ScanResult, error) {
	result := &ScanResult{
		Target:    cfg.Target,
		StartTime: time.Now().Format(time.RFC3339),
		Engine:    "rustscan",
	}

	startTime := time.Now()

	if !cfg.Quiet {
		log.Printf("[RustScan] Fast scanning %s ports=%s", cfg.Target, cfg.Ports)
	}

	args := []string{
		"-a", cfg.Target,
		"-p", cfg.Ports,
		"-t", strconv.Itoa(cfg.Threads),
		"--timeout", strconv.Itoa(cfg.Timeout * 1000),
		"--no-config",
		"--greppable",
	}

	cmd := exec.Command("rustscan", args...)
	cmd.Stderr = nil

	output, err := cmd.Output()
	if err != nil && len(output) == 0 {
		return nil, fmt.Errorf("rustscan: %w", err)
	}

	openPorts := make(map[int]bool)
	re := regexp.MustCompile(`(\d+)\s*->\s*([\d.]+)`)

	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "Open ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				addr := parts[1]
				if idx := strings.LastIndex(addr, ":"); idx >= 0 {
					port, _ := strconv.Atoi(addr[idx+1:])
					if port > 0 {
						openPorts[port] = true
					}
				}
			}
		}
		if matches := re.FindStringSubmatch(line); len(matches) >= 3 {
			port, _ := strconv.Atoi(matches[1])
			openPorts[port] = true
		}
	}

	if !cfg.Quiet {
		log.Printf("[RustScan] Discovered %d open ports", len(openPorts))
	}

	// Try Nmap for service detection if available
	if cfg.ServiceDetect && len(openPorts) > 0 {
		portList := make([]string, 0, len(openPorts))
		for p := range openPorts {
			portList = append(portList, strconv.Itoa(p))
		}
		sort.Strings(portList)
		portsStr := strings.Join(portList, ",")

		nmapArgs := []string{
			"-sS", "-sV", "--version-intensity", "5",
			"-T4", "-p", portsStr, "-oX", "-",
			"--max-retries", "2", cfg.Target,
		}

		nmapCmd := exec.Command("nmap", nmapArgs...)
		var nmapOut bytes.Buffer
		nmapCmd.Stdout = &nmapOut
		if err := nmapCmd.Run(); err == nil {
			parsedHosts := ParseNmapXML(nmapOut.String())
			if len(parsedHosts) > 0 {
				result.Hosts = parsedHosts
				for _, h := range result.Hosts {
					result.TotalOpen += h.OpenCount
				}
				result.EndTime = time.Now().Format(time.RFC3339)
				result.DurationSec = time.Since(startTime).Seconds()
				result.Duration = fmt.Sprintf("%.1fs", result.DurationSec)
				if !cfg.Quiet {
					log.Printf("[RustScan] Done: %d hosts, %d open ports in %s",
						len(result.Hosts), result.TotalOpen, result.Duration)
				}
				return result, nil
			}
		}
	}

	// Fallback: just list open ports
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
		log.Printf("[RustScan] Done: %d hosts, %d open ports in %s",
			len(result.Hosts), result.TotalOpen, result.Duration)
	}

	return result, nil
}
