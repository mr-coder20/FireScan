package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mr-coder20/FireScan/internal/config"
	"github.com/mr-coder20/FireScan/internal/output"
	"github.com/mr-coder20/FireScan/internal/scanner"
	"github.com/spf13/cobra"
)

var cfg *config.Config

func main() {
	rootCmd := &cobra.Command{
		Use:   "firescan",
		Short: "🔥 FireScan — Ultimate Hybrid Port Scanner",
		Long: `FireScan is an intelligent hybrid port scanner that combines
the SPEED of Masscan/RustScan with the DEPTH of Nmap.

  "Speed of Masscan ❄ Depth of Nmap"

It automatically selects the best scanning engine based on your target
and requirements. Supports all major output formats including HTML reports.`,
		Version: config.Version,
		RunE:    runScan,
	}

	// Target flags
	rootCmd.Flags().StringP("target", "t", "", "Target IP, hostname, CIDR, or range")
	rootCmd.Flags().StringP("ports", "p", "1-1000", "Port range (e.g. 22,80,443 or 1-1000)")
	rootCmd.Flags().Bool("all-ports", false, "Scan all 65535 ports")
	rootCmd.Flags().Int("top-ports", 0, "Scan top N ports")

	// Engine flags
	rootCmd.Flags().StringP("engine", "e", "auto", "Scanner engine: auto, masscan, rustscan, nmap, naive")
	rootCmd.Flags().Int("rate", 10000, "Packets per second")
	rootCmd.Flags().Int("retries", 3, "Max retries")
	rootCmd.Flags().IntP("timeout", "T", 30, "Timeout in seconds")
	rootCmd.Flags().IntP("threads", "n", runtime.NumCPU()*2, "Thread count")

	// Feature flags - FIXED: "sV" changed to "V" (single char)
	rootCmd.Flags().BoolP("service-detect", "V", true, "Service/version detection")
	rootCmd.Flags().BoolP("os-detect", "O", false, "OS detection")
	rootCmd.Flags().Bool("vuln", false, "Vulnerability scan (Nmap NSE vuln scripts)")
	rootCmd.Flags().Bool("fast", false, "Fast mode")

	// Output flags
	rootCmd.Flags().StringP("output", "o", "", "Output file")
	rootCmd.Flags().StringP("format", "f", "table", "Output format: table, json, csv, html, greppable")
	rootCmd.Flags().BoolP("quiet", "q", false, "Quiet mode")
	rootCmd.Flags().Bool("no-banner", false, "Hide banner")
	rootCmd.Flags().Bool("pipe", false, "Read targets from stdin")

	rootCmd.SetVersionTemplate(`{{printf "FireScan version %s\n" .Version}}`)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runScan(cmd *cobra.Command, args []string) error {
	cfg = config.FromFlags(cmd, args)

	if !cfg.Quiet && !cfg.NoBanner {
		output.PrintBanner()
	}

	tools := scanner.DetectTools()
	if !cfg.Quiet {
		output.PrintToolStatus(tools)
	}

	engine := scanner.NewEngine(cfg, tools)
	if !cfg.Quiet {
		output.PrintEngineSelection(engine.Name(), cfg)
	}

	targets := cfg.GetTargets()
	if len(targets) == 0 {
		return fmt.Errorf("no target specified. Use: firescan <target> -p <ports>")
	}

	var allResults []*scanner.ScanResult
	for i, target := range targets {
		if len(targets) > 1 && !cfg.Quiet {
			output.PrintProgress(i+1, len(targets), target)
		}
		cfg.Target = target
		result, err := engine.Scan(cfg)
		if err != nil {
			output.PrintError(target, err)
			result = &scanner.ScanResult{
				Target: target,
				Engine: engine.Name(),
				Errors: []string{err.Error()},
			}
		}
		allResults = append(allResults, result)
	}

	finalResult := scanner.MergeResults(allResults)
	output.Render(finalResult, cfg)

	return nil
}
