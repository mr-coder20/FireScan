package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mr-coder20/FireScan/internal/config"
	"github.com/mr-coder20/FireScan/internal/scanner"
)

func PrintBanner() {
	fmt.Println(config.Banner)
}

func PrintToolStatus(tools scanner.ToolStatus) {
	var parts []string
	if tools.Masscan {
		parts = append(parts, "✅ masscan")
	} else {
		parts = append(parts, "❌ masscan")
	}
	if tools.RustScan {
		parts = append(parts, "✅ rustscan")
	} else {
		parts = append(parts, "❌ rustscan")
	}
	if tools.Nmap {
		parts = append(parts, "✅ nmap")
	} else {
		parts = append(parts, "❌ nmap")
	}
	if tools.Naabu {
		parts = append(parts, "✅ naabu")
	} else {
		parts = append(parts, "❌ naabu")
	}
	fmt.Printf("  Tools: %s\n\n", strings.Join(parts, " | "))
}

func PrintEngineSelection(engine string, cfg *config.Config) {
	fmt.Printf("  Engine: %s\n", engine)
	fmt.Printf("  Target: %s\n", cfg.Target)
	fmt.Printf("  Ports:  %s\n\n", cfg.Ports)
}

func PrintProgress(current, total int, target string) {
	fmt.Printf("\n  [%d/%d] Scanning: %s\n", current, total, target)
}

func PrintError(target string, err error) {
	fmt.Printf("  ✗ Error: %s — %v\n", target, err)
}

func Render(result *scanner.ScanResult, cfg *config.Config) {
	var output string

	switch cfg.Format {
	case "json":
		output = formatJSON(result)
	case "csv":
		output = "IP,Port,Protocol,State,Service,Version\n" + formatCSV(result)
	case "html":
		output = formatHTML(result)
	case "greppable":
		output = formatGreppable(result)
	default:
		output = formatTable(result)
	}

	fmt.Print(output)

	if cfg.OutputFile != "" {
		ext := strings.ToLower(filepath.Ext(cfg.OutputFile))
		switch ext {
		case ".json":
			output = formatJSON(result)
		case ".html", ".htm":
			output = formatHTML(result)
		case ".csv":
			output = "IP,Port,Protocol,State,Service,Version\n" + formatCSV(result)
		}

		dir := filepath.Dir(cfg.OutputFile)
		if dir != "." {
			os.MkdirAll(dir, 0755)
		}
		if err := os.WriteFile(cfg.OutputFile, []byte(output), 0644); err != nil {
			fmt.Printf("  ✗ Error writing file: %v\n", err)
		} else {
			fmt.Printf("\n  ✓ Report saved to: %s\n", cfg.OutputFile)
		}
	}
}

func formatTable(result *scanner.ScanResult) string {
	var b strings.Builder

	b.WriteString(strings.Repeat("─", 60) + "\n")
	b.WriteString(fmt.Sprintf("  FireScan Summary\n"))
	b.WriteString(strings.Repeat("─", 60) + "\n")
	b.WriteString(fmt.Sprintf("  Target:  %s\n", result.Target))
	b.WriteString(fmt.Sprintf("  Engine:  %s\n", result.Engine))
	b.WriteString(fmt.Sprintf("  Time:    %s\n", result.Duration))
	b.WriteString(fmt.Sprintf("  Hosts:   %d\n", len(result.Hosts)))
	b.WriteString(fmt.Sprintf("  Open:    %d\n", result.TotalOpen))
	b.WriteString(strings.Repeat("─", 60) + "\n\n")

	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			b.WriteString(fmt.Sprintf("  ⚠ %s\n", err))
		}
		b.WriteString("\n")
	}

	for _, host := range result.Hosts {
		b.WriteString(fmt.Sprintf("  %s", host.IP))
		if host.Hostname != "" {
			b.WriteString(fmt.Sprintf(" (%s)", host.Hostname))
		}
		if host.OS != "" {
			b.WriteString(fmt.Sprintf(" [OS: %s]", host.OS))
		}
		b.WriteString("\n")

		if host.OpenCount == 0 {
			b.WriteString("  └─ No open ports found\n\n")
			continue
		}

		b.WriteString(fmt.Sprintf("  ┌─ Open ports: %d\n", host.OpenCount))
		b.WriteString(fmt.Sprintf("  │  %-6s %-6s %-8s %-15s %s\n",
			"PORT", "PROTO", "STATE", "SERVICE", "VERSION"))
		b.WriteString("  │  " + strings.Repeat("─", 55) + "\n")

		for _, p := range host.Ports {
			svc := p.Service
			if p.Product != "" {
				svc = p.Product
			}
			ver := p.Version
			if ver == "" {
				ver = p.Banner
			}
			if len(ver) > 25 {
				ver = ver[:25] + "…"
			}
			b.WriteString(fmt.Sprintf("  │  %-6d %-6s %-8s %-15s %s\n",
				p.Port, p.Protocol, p.State, svc, ver))
		}
		b.WriteString("  └─ " + strings.Repeat("─", 55) + "\n\n")
	}

	return b.String()
}

func formatJSON(result *scanner.ScanResult) string {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error":"%s"}`, err.Error())
	}
	return string(data)
}

func formatCSV(result *scanner.ScanResult) string {
	var rows []string
	for _, host := range result.Hosts {
		for _, p := range host.Ports {
			row := fmt.Sprintf("%s,%d,%s,%s,%s,%s",
				host.IP, p.Port, p.Protocol, p.State, p.Service, p.Version)
			rows = append(rows, row)
		}
	}
	return strings.Join(rows, "\n")
}

func formatGreppable(result *scanner.ScanResult) string {
	var b strings.Builder
	for _, host := range result.Hosts {
		for _, p := range host.Ports {
			if p.State == "open" {
				fmt.Fprintf(&b, "open %s %s %s:%d\n",
					p.Protocol, strconv.Itoa(p.Port), host.IP, p.Port)
			}
		}
	}
	return b.String()
}

func formatHTML(result *scanner.ScanResult) string {
	var portRows strings.Builder
	for _, host := range result.Hosts {
		for _, p := range host.Ports {
			bgColor := "#e8f5e9"
			if p.State == "filtered" {
				bgColor = "#fff3e0"
			} else if p.State == "closed" {
				bgColor = "#ffebee"
			}
			portRows.WriteString(fmt.Sprintf(`
			<tr style="background-color: %s;">
				<td>%s</td>
				<td>%d</td>
				<td>%s</td>
				<td>%s</td>
				<td>%s</td>
				<td>%s</td>
			</tr>`, bgColor, host.IP, p.Port, p.Protocol, p.State, p.Service, p.Version))
		}
	}

	hostCount := len(result.Hosts)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>FireScan Report — %s</title>
	<style>
		* { margin: 0; padding: 0; box-sizing: border-box; }
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			background: #0f172a;
			color: #e2e8f0;
			padding: 20px;
		}
		.container { max-width: 1200px; margin: 0 auto; }
		.header {
			text-align: center;
			padding: 40px 0;
			border-bottom: 1px solid #334155;
			margin-bottom: 30px;
		}
		.header h1 {
			font-size: 2.5em;
			background: linear-gradient(135deg, #38bdf8, #818cf8);
			-webkit-background-clip: text;
			-webkit-text-fill-color: transparent;
		}
		.summary-grid {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
			gap: 15px;
			margin-bottom: 30px;
		}
		.summary-card {
			background: #1e293b;
			border: 1px solid #334155;
			border-radius: 12px;
			padding: 20px;
			text-align: center;
		}
		.summary-card .label { color: #94a3b8; font-size: 0.9em; text-transform: uppercase; }
		.summary-card .value { font-size: 2em; font-weight: bold; margin-top: 5px; }
		table {
			width: 100%%;
			border-collapse: collapse;
			background: #1e293b;
			border-radius: 12px;
			overflow: hidden;
			margin-top: 20px;
		}
		th { background: #334155; padding: 12px 10px; text-align: left; font-weight: 600; color: #94a3b8; }
		td { padding: 10px; border-bottom: 1px solid #1e293b; }
		tr:hover { opacity: 0.9; }
		.footer { text-align: center; padding: 30px; color: #64748b; margin-top: 40px; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>🔥 FireScan Report</h1>
			<p>Generated by FireScan v%s</p>
		</div>
		<div class="summary-grid">
			<div class="summary-card">
				<div class="label">Target</div>
				<div class="value">%s</div>
			</div>
			<div class="summary-card">
				<div class="label">Engine</div>
				<div class="value">%s</div>
			</div>
			<div class="summary-card">
				<div class="label">Duration</div>
				<div class="value">%s</div>
			</div>
			<div class="summary-card">
				<div class="label">Hosts</div>
				<div class="value">%d</div>
			</div>
			<div class="summary-card">
				<div class="label">Open Ports</div>
				<div class="value" style="color: #4ade80;">%d</div>
			</div>
		</div>
		<table>
			<thead>
				<tr>
					<th>IP</th><th>Port</th><th>Protocol</th><th>State</th><th>Service</th><th>Version</th>
				</tr>
			</thead>
			<tbody>%s</tbody>
		</table>
		<div class="footer">FireScan v%s — MIT License</div>
	</div>
</body>
</html>`,
		result.Target, config.Version,
		result.Target, result.Engine, result.Duration,
		hostCount, result.TotalOpen,
		portRows.String(), config.Version)
}
