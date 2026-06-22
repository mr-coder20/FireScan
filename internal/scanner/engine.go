package scanner

import (
	"encoding/json"
	"os/exec"
	"time"

	"github.com/mr-coder20/FireScan/internal/config"
)

type PortInfo struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	State    string `json:"state"`
	Service  string `json:"service,omitempty"`
	Version  string `json:"version,omitempty"`
	Product  string `json:"product,omitempty"`
	Banner   string `json:"banner,omitempty"`
	Extra    string `json:"extra,omitempty"`
	OS       string `json:"os,omitempty"`
	CPE      string `json:"cpe,omitempty"`
}

type HostResult struct {
	IP        string     `json:"ip"`
	Hostname  string     `json:"hostname,omitempty"`
	MAC       string     `json:"mac,omitempty"`
	Status    string     `json:"status"`
	Ports     []PortInfo `json:"ports"`
	OpenCount int        `json:"open_count"`
	OS        string     `json:"os,omitempty"`
}

type ScanResult struct {
	Target      string       `json:"target"`
	StartTime   string       `json:"start_time"`
	EndTime     string       `json:"end_time"`
	Duration    string       `json:"duration"`
	DurationSec float64      `json:"duration_seconds"`
	Hosts       []HostResult `json:"hosts"`
	TotalOpen   int          `json:"total_open_ports"`
	Engine      string       `json:"engine_used"`
	Errors      []string     `json:"errors,omitempty"`
}

type Engine interface {
	Name() string
	Scan(cfg *config.Config) (*ScanResult, error)
}

type ToolStatus struct {
	Masscan  bool
	RustScan bool
	Nmap     bool
	Naabu    bool
}

func DetectTools() ToolStatus {
	return ToolStatus{
		Masscan:  commandExists("masscan"),
		RustScan: commandExists("rustscan"),
		Nmap:     commandExists("nmap"),
		Naabu:    commandExists("naabu"),
	}
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func NewEngine(cfg *config.Config, tools ToolStatus) Engine {
	engineName := cfg.Engine
	if engineName == "auto" {
		engineName = selectEngine(cfg, tools)
	}

	switch engineName {
	case "masscan":
		return &MasscanEngine{}
	case "rustscan":
		return &RustScanEngine{}
	case "nmap":
		return &NmapEngine{}
	case "naabu":
		return &NaabuEngine{}
	default:
		return &NaiveEngine{}
	}
}

func selectEngine(cfg *config.Config, tools ToolStatus) string {
	isFullScan := cfg.Ports == "1-65535" || cfg.AllPorts
	isWideScan := stringsContains(cfg.Target, "/") || stringsContains(cfg.Target, "-")

	switch {
	case isWideScan && tools.Masscan:
		return "masscan"
	case isFullScan && tools.RustScan:
		return "rustscan"
	case tools.Nmap:
		return "nmap"
	case tools.Naabu:
		return "naabu"
	default:
		return "naive"
	}
}

func stringsContains(s, sub string) bool {
	return len(s) >= len(sub) && containsStr(s, sub)
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func MergeResults(results []*ScanResult) *ScanResult {
	if len(results) == 0 {
		return &ScanResult{}
	}
	if len(results) == 1 {
		return results[0]
	}
	merged := results[0]
	for _, r := range results[1:] {
		merged.Hosts = append(merged.Hosts, r.Hosts...)
		merged.TotalOpen += r.TotalOpen
		merged.Errors = append(merged.Errors, r.Errors...)
	}
	merged.EndTime = time.Now().Format(time.RFC3339)
	return merged
}

func (r *ScanResult) ToJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}
