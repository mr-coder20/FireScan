package scanner

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mr-coder20/FireScan/internal/config"
)

type NmapEngine struct{}

func (e *NmapEngine) Name() string { return "nmap" }

func (e *NmapEngine) Scan(cfg *config.Config) (*ScanResult, error) {
	result := &ScanResult{
		Target:    cfg.Target,
		StartTime: time.Now().Format(time.RFC3339),
		Engine:    "nmap",
	}

	startTime := time.Now()
	ports := cfg.Ports
	if cfg.AllPorts {
		ports = "1-65535"
	}

	args := []string{"-sS", "-T4", "--max-retries", strconv.Itoa(cfg.Retries)}
	if cfg.ServiceDetect {
		args = append(args, "-sV", "--version-intensity", "7")
	}
	if cfg.OSDetect {
		args = append(args, "-O")
	}
	if cfg.VulnScan {
		args = append(args, "--script", "vuln")
	}
	args = append(args, "-oX", "-", "-p", ports, cfg.Target)

	if !cfg.Quiet {
		log.Printf("[Nmap] nmap %s", strings.Join(args, " "))
	}

	cmd := exec.Command("nmap", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("nmap failed: %w\n%s", err, stderr.String())
	}

	result.Hosts = ParseNmapXML(stdout.String())
	for _, h := range result.Hosts {
		result.TotalOpen += h.OpenCount
	}
	result.EndTime = time.Now().Format(time.RFC3339)
	result.DurationSec = time.Since(startTime).Seconds()
	result.Duration = fmt.Sprintf("%.1fs", result.DurationSec)

	if !cfg.Quiet {
		log.Printf("[Nmap] Done: %d hosts, %d open ports in %s",
			len(result.Hosts), result.TotalOpen, result.Duration)
	}

	return result, nil
}

func ParseNmapXML(xml string) []HostResult {
	var hosts []HostResult
	hostRegex := regexp.MustCompile(`(?s)<host[^>]*>(.*?)</host>`)

	for _, m := range hostRegex.FindAllStringSubmatch(xml, -1) {
		hostData := m[1]
		host := HostResult{}

		if ipMatch := regexp.MustCompile(`<address addr="([^"]+)" addrtype="ipv[46]"`).FindStringSubmatch(hostData); len(ipMatch) > 1 {
			host.IP = ipMatch[1]
		}
		if macMatch := regexp.MustCompile(`<address addr="([^"]+)" addrtype="mac"`).FindStringSubmatch(hostData); len(macMatch) > 1 {
			host.MAC = macMatch[1]
		}
		if hnMatch := regexp.MustCompile(`<hostname name="([^"]+)"`).FindStringSubmatch(hostData); len(hnMatch) > 1 {
			host.Hostname = hnMatch[1]
		}
		if statusMatch := regexp.MustCompile(`<status state="([^"]+)"`).FindStringSubmatch(hostData); len(statusMatch) > 1 {
			host.Status = statusMatch[1]
		}
		if osMatch := regexp.MustCompile(`<osmatch name="([^"]+)"`).FindStringSubmatch(hostData); len(osMatch) > 1 {
			host.OS = osMatch[1]
		}

		portRegex := regexp.MustCompile(`(?s)<port protocol="([^"]+)" portid="(\d+)">(.*?)</port>`)
		for _, pm := range portRegex.FindAllStringSubmatch(hostData, -1) {
			pi := PortInfo{Protocol: pm[1]}
			pi.Port, _ = strconv.Atoi(pm[2])
			portBody := pm[3]
			if stateMatch := regexp.MustCompile(`<state state="([^"]+)"`).FindStringSubmatch(portBody); len(stateMatch) > 1 {
				pi.State = stateMatch[1]
			}
			if svcMatch := regexp.MustCompile(`<service name="([^"]+)"`).FindStringSubmatch(portBody); len(svcMatch) > 1 {
				pi.Service = svcMatch[1]
				if prodMatch := regexp.MustCompile(`product="([^"]+)"`).FindStringSubmatch(portBody); len(prodMatch) > 1 {
					pi.Product = prodMatch[1]
				}
				if verMatch := regexp.MustCompile(`version="([^"]+)"`).FindStringSubmatch(portBody); len(verMatch) > 1 {
					pi.Version = verMatch[1]
				}
			}
			host.Ports = append(host.Ports, pi)
		}

		for _, p := range host.Ports {
			if p.State == "open" {
				host.OpenCount++
			}
		}

		hosts = append(hosts, host)
	}
	return hosts
}
