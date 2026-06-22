package api

import (
	"encoding/json"
	"fmt"
	"github.com/mr-coder20/FireScan/internal/config"
	"github.com/mr-coder20/FireScan/internal/output"
	"github.com/mr-coder20/FireScan/internal/scanner"
)

// FireScanAPI provides a public API for embedding FireScan in other Go programs
type FireScanAPI struct {
	cfg *config.Config
}

// ScanConfig mirrors the CLI config for programmatic use
type ScanConfig struct {
	Target        string `json:"target"`
	Ports         string `json:"ports"`
	Engine        string `json:"engine"`
	Rate          int    `json:"rate"`
	Retries       int    `json:"retries"`
	Timeout       int    `json:"timeout"`
	Threads       int    `json:"threads"`
	ServiceDetect bool   `json:"serviceDetect"`
	OSDetect      bool   `json:"osDetect"`
	VulnScan      bool   `json:"vulnScan"`
	FastMode      bool   `json:"fastMode"`
	Format        string `json:"format"`
}

// ScanResult is the public scan result type
type ScanResult struct {
	Target      string             `json:"target"`
	Duration    string             `json:"duration"`
	DurationSec float64            `json:"durationSeconds"`
	Hosts       []scanner.HostResult `json:"hosts"`
	TotalOpen   int                `json:"totalOpenPorts"`
	Engine      string             `json:"engine"`
	Errors      []string           `json:"errors,omitempty"`
}

// NewScanner creates a new FireScan scanner
func NewScanner(cfg *ScanConfig) *FireScanAPI {
	c := config.Default()
	c.Target = cfg.Target
	c.Ports = cfg.Ports
	if c.Ports == "" {
		c.Ports = "1-1000"
	}
	c.Engine = cfg.Engine
	if c.Engine == "" {
		c.Engine = "auto"
	}
	c.Rate = cfg.Rate
	if c.Rate == 0 {
		c.Rate = 10000
	}
	c.Retries = cfg.Retries
	if c.Retries == 0 {
		c.Retries = 3
	}
	c.Timeout = cfg.Timeout
	if c.Timeout == 0 {
		c.Timeout = 30
	}
	c.Threads = cfg.Threads
	if c.Threads == 0 {
		c.Threads = 100
	}
	c.ServiceDetect = cfg.ServiceDetect
	c.OSDetect = cfg.OSDetect
	c.VulnScan = cfg.VulnScan
	c.FastMode = cfg.FastMode
	c.Format = cfg.Format
	if c.Format == "" {
		c.Format = "json"
	}
	c.Quiet = true
	c.NoBanner = true

	return &FireScanAPI{cfg: c}
}

// Run executes the scan and returns results
func (api *FireScanAPI) Run() (*ScanResult, error) {
	tools := scanner.DetectTools()
	engine := scanner.NewEngine(api.cfg, tools)
	api.cfg.Engine = engine.Name()

	result, err := engine.Scan(api.cfg)
	if err != nil {
		return nil, err
	}

	return &ScanResult{
		Target:      result.Target,
		Duration:    result.Duration,
		DurationSec: result.DurationSec,
		Hosts:       result.Hosts,
		TotalOpen:   result.TotalOpen,
		Engine:      result.Engine,
		Errors:      result.Errors,
	}, nil
}

// RunJSON executes scan and returns JSON string
func (api *FireScanAPI) RunJSON() string {
	result, err := api.Run()
	if err != nil {
		return fmt.Sprintf(`{"error":"%s"}`, err.Error())
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data)
}

// RunHTML executes scan and returns HTML report
func (api *FireScanAPI) RunHTML() string {
	result, err := api.Run()
	if err != nil {
		return fmt.Sprintf("<html><body><h1>Error</h1><p>%s</p></body></html>", err.Error())
	}
	scanResult := &scanner.ScanResult{
		Target:      result.Target,
		Duration:    result.Duration,
		DurationSec: result.DurationSec,
		Hosts:       result.Hosts,
		TotalOpen:   result.TotalOpen,
		Engine:      result.Engine,
		Errors:      result.Errors,
	}
	return output.FormatHTML(scanResult)
}