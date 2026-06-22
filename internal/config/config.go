package config

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const (
	Name     = "FireScan"
	Version  = "5.0.0"
	Homepage = "https://github.com/mr-coder20/FireScan"
)

var Banner = fmt.Sprintf(`
    ███████╗██╗██████╗ ███████╗███████╗ ██████╗ █████╗ ███╗   ██╗
    ██╔════╝██║██╔══██╗██╔════╝██╔════╝██╔════╝██╔══██╗████╗  ██║
    █████╗  ██║██████╔╝█████╗  ███████╗██║     ███████║██╔██╗ ██║
    ██╔══╝  ██║██╔══██╗██╔══╝  ╚════██║██║     ██╔══██║██║╚██╗██║
    ██║     ██║██║  ██║███████╗███████║╚██████╗██║  ██║██║ ╚████║
    ╚═╝     ╚═╝╚═╝  ╚═╝╚══════╝╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝
                                                                    
   ╔══════════════════════════════════════════════════════════════╗
   ║     Ultimate Hybrid Port Scanner  v%s                       ║
   ║     "Speed of Masscan ❄ Depth of Nmap"                     ║
   ║     %s                                                      ║
   ╚══════════════════════════════════════════════════════════════╝`, Version, Homepage)

type Config struct {
	Target        string
	Ports         string
	Engine        string
	Rate          int
	Retries       int
	Timeout       int
	Threads       int
	AllPorts      bool
	TopPorts      int
	ServiceDetect bool
	OSDetect      bool
	VulnScan      bool
	FastMode      bool
	OutputFile    string
	Format        string
	Quiet         bool
	NoBanner      bool
	PipeMode      bool
	CommandLine   string
}

func Default() *Config {
	return &Config{
		Ports:         "1-1000",
		Engine:        "auto",
		Rate:          10000,
		Retries:       3,
		Timeout:       30,
		ServiceDetect: true,
		Format:        "table",
		Threads:       runtime.NumCPU() * 2,
	}
}

func FromFlags(cmd *cobra.Command, args []string) *Config {
	cfg := Default()

	cfg.Target, _ = cmd.Flags().GetString("target")
	cfg.Ports, _ = cmd.Flags().GetString("ports")
	cfg.AllPorts, _ = cmd.Flags().GetBool("all-ports")
	cfg.TopPorts, _ = cmd.Flags().GetInt("top-ports")
	cfg.Engine, _ = cmd.Flags().GetString("engine")
	cfg.Rate, _ = cmd.Flags().GetInt("rate")
	cfg.Retries, _ = cmd.Flags().GetInt("retries")
	cfg.Timeout, _ = cmd.Flags().GetInt("timeout")
	if t, _ := cmd.Flags().GetInt("threads"); t > 0 {
		cfg.Threads = t
	}
	cfg.ServiceDetect, _ = cmd.Flags().GetBool("service-detect")
	cfg.OSDetect, _ = cmd.Flags().GetBool("os-detect")
	cfg.VulnScan, _ = cmd.Flags().GetBool("vuln")
	cfg.FastMode, _ = cmd.Flags().GetBool("fast")
	cfg.OutputFile, _ = cmd.Flags().GetString("output")
	cfg.Format, _ = cmd.Flags().GetString("format")
	cfg.Quiet, _ = cmd.Flags().GetBool("quiet")
	cfg.NoBanner, _ = cmd.Flags().GetBool("no-banner")

	if p, _ := cmd.Flags().GetBool("pipe"); p {
		cfg.PipeMode = true
	}

	if cfg.Target == "" && len(args) > 0 {
		cfg.Target = args[0]
	}

	if !cfg.PipeMode {
		stat, _ := os.Stdin.Stat()
		cfg.PipeMode = (stat.Mode()&os.ModeCharDevice) == 0
	}

	if cfg.AllPorts {
		cfg.Ports = "1-65535"
	}

	cfg.CommandLine = strings.Join(os.Args[1:], " ")
	return cfg
}

func (c *Config) GetTargets() []string {
	if c.PipeMode {
		var targets []string
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				targets = append(targets, line)
			}
		}
		if len(targets) > 0 {
			return targets
		}
	}
	if c.Target != "" {
		return []string{c.Target}
	}
	return nil
}