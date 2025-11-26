package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/typicalfo/netgaze/internal/collector"
	"github.com/typicalfo/netgaze/internal/model"
	"github.com/typicalfo/netgaze/internal/ui"
)

var (
	enablePorts bool
	output      string
	noStyle     bool
	timeout     time.Duration

	// traceroute subcommand flags
	tracerouteOutFile  string
	tracerouteBaseFile string
	tracerouteNewFile  string
	tracerouteDiffFile string
)

var rootCmd = &cobra.Command{
	Use:   "ng [flags] <ip|domain|url>",
	Short: "Network info gathering tool",
	Long: `netgaze performs common network diagnostics and compiles the data

Mode:
  - Deterministic, offline templated output (no AI)

Examples:
  ng 1.1.1.1
  ng tui google.com --ports
  ng example.com
  `,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE:         runNetgaze,
}

var tuiCmd = &cobra.Command{
	Use:   "tui [flags] <ip|domain|url>",
	Short: "Launch interactive TUI mode",
	Long:  `Launch the terminal user interface for interactive network analysis.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runNetgaze,
}

var tracerouteOutputCmd = &cobra.Command{
	Use:   "to [flags] <ip|domain|url>",
	Short: "Run traceroute and write JSON output",
	Long:  `Perform a traceroute and save the hop list to a JSON file.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTracerouteOutput,
}

var tracerouteCompareCmd = &cobra.Command{
	Use:   "tc [flags] <ip|domain|url>",
	Short: "Run traceroute and compare with a baseline",
	Long:  `Perform a traceroute, compare it with a baseline JSON file, and save the new trace.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTracerouteCompare,
}

func init() {
	rootCmd.Flags().BoolVar(&enablePorts, "ports", false,
		"Enable port scan of common ports (not enabled by default)")
	rootCmd.Flags().StringVar(&output, "output", "text",
		"Output format: text, md, json, raw (for piping or automation)")
	rootCmd.Flags().BoolVar(&noStyle, "no-style", false,
		"Disable all terminal styling and ANSI escape codes")
	var jsonFlag string
	rootCmd.Flags().StringVar(&jsonFlag, "json", "",
		"Legacy alias for --output json")
	rootCmd.Flags().DurationVar(&timeout, "timeout", 15*time.Second,
		"Global timeout for all operations")

	// Hide the legacy --json flag from help but keep for compatibility
	rootCmd.Flags().MarkHidden("json")

	// Add flags to TUI command
	tuiCmd.Flags().BoolVar(&enablePorts, "ports", false,
		"Enable port scan of common ports (not enabled by default)")
	tuiCmd.Flags().DurationVar(&timeout, "timeout", 15*time.Second,
		"Global timeout for all operations")

	// Traceroute output flags
	tracerouteOutputCmd.Flags().StringVarP(&tracerouteOutFile, "out", "o", "", "Output JSON file for traceroute (default: traceroute-<target>-<timestamp>.json)")

	// Traceroute compare flags
	tracerouteCompareCmd.Flags().StringVarP(&tracerouteBaseFile, "base", "b", "", "Baseline traceroute JSON file (optional; auto-detected if omitted)")
	tracerouteCompareCmd.Flags().StringVarP(&tracerouteNewFile, "new", "n", "", "Output JSON file for new traceroute (default: traceroute-new.json)")
	tracerouteCompareCmd.Flags().StringVarP(&tracerouteDiffFile, "diff", "d", "", "Write human-readable comparison output to this file")

	// Add subcommands
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(tracerouteOutputCmd)
	rootCmd.AddCommand(tracerouteCompareCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func runTracerouteOutput(cmd *cobra.Command, args []string) error {
	target := args[0]

	normalizedTarget, err := validateTarget(target)
	if err != nil {
		return fmt.Errorf("invalid target: %w", err)
	}

	// Perform traceroute using collector's helper
	ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
	defer cancel()

	hops, err := collector.Traceroute(ctx, normalizedTarget, timeout)
	if err != nil {
		return fmt.Errorf("traceroute failed: %w", err)
	}

	if tracerouteOutFile == "" {
		timestamp := time.Now().UTC().Format("20060102-150405")
		tracerouteOutFile = fmt.Sprintf("traceroute-%s-%s.json", normalizedTarget, timestamp)
	}

	data, err := json.MarshalIndent(hops, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal traceroute: %w", err)
	}

	if err := os.WriteFile(tracerouteOutFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write traceroute file: %w", err)
	}

	fmt.Printf("Traceroute saved to %s\n", tracerouteOutFile)
	return nil
}

func runTracerouteCompare(cmd *cobra.Command, args []string) error {
	target := args[0]

	normalizedTarget, err := validateTarget(target)
	if err != nil {
		return fmt.Errorf("invalid target: %w", err)
	}

	// Determine baseline file: explicit --base, most recent for target, or fresh run
	basePath := tracerouteBaseFile
	if basePath == "" {
		// Try most recent traceroute JSON for this target in current directory
		pattern := fmt.Sprintf("traceroute-%s-*.json", normalizedTarget)
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			// Use lexicographically last as most recent
			sort.Strings(matches)
			basePath = matches[len(matches)-1]
		}
	}

	if basePath == "" {
		// No baseline available; run traceroute output once to create it
		fmt.Println("No baseline traceroute found; running initial traceroute...")
		if err := runTracerouteOutput(cmd, args); err != nil {
			return err
		}
		// After creating baseline, use default output filename
		basePath = tracerouteOutFile
	}

	// Load baseline hops
	baseData, err := os.ReadFile(basePath)
	if err != nil {
		return fmt.Errorf("failed to read baseline file: %w", err)
	}

	var baseHops []model.TraceHop
	if err := json.Unmarshal(baseData, &baseHops); err != nil {
		return fmt.Errorf("failed to parse baseline file: %w", err)
	}

	// Run new traceroute
	ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
	defer cancel()

	newHops, err := collector.Traceroute(ctx, normalizedTarget, timeout)
	if err != nil {
		return fmt.Errorf("traceroute failed: %w", err)
	}

	if tracerouteNewFile == "" {
		tracerouteNewFile = "traceroute-new.json"
	}

	data, err := json.MarshalIndent(newHops, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal new traceroute: %w", err)
	}

	if err := os.WriteFile(tracerouteNewFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write new traceroute file: %w", err)
	}

	// Compare and print summary
	fmt.Println("Traceroute comparison:")
	for i := range baseHops {
		if i >= len(newHops) {
			fmt.Printf("Hop %d: removed (was %s)\n", baseHops[i].Hop, baseHops[i].IP)
			continue
		}
		if baseHops[i].IP != newHops[i].IP {
			fmt.Printf("Hop %d changed: %s -> %s\n", baseHops[i].Hop, baseHops[i].IP, newHops[i].IP)
		}
	}
	if len(newHops) > len(baseHops) {
		for i := len(baseHops); i < len(newHops); i++ {
			fmt.Printf("Hop %d: new (%s)\n", newHops[i].Hop, newHops[i].IP)
		}
	}

	return nil

	return nil
}

func runNetgaze(cmd *cobra.Command, args []string) error {
	target := args[0]

	// Handle legacy --json flag
	if jsonFlag, _ := cmd.Flags().GetString("json"); jsonFlag != "" {
		output = "json"
	}

	// Validate target
	normalizedTarget, err := validateTarget(target)
	if err != nil {
		return fmt.Errorf("invalid target: %w", err)
	}

	// Validate flags
	if err := validateFlags(); err != nil {
		return err
	}

	// Check if TUI mode is explicitly requested (via subcommand)
	if cmd.HasParent() && cmd.Parent().Name() == "tui" {
		// Run with TUI (no AI in this version)
		return ui.RunTUI(normalizedTarget, true, enablePorts, timeout, nil)
	}

	// Run collection and output to stdout
	report, err := collector.Collect(cmd.Context(), normalizedTarget, collector.Options{
		EnablePorts: enablePorts,
		NoAgent:     true,
		Timeout:     timeout,
	})
	if err != nil {
		return fmt.Errorf("collection failed: %w", err)
	}

	// Output based on format
	return outputReport(report, output)
}

func outputReport(report *model.Report, format string) error {
	switch format {
	case "json":
		return outputJSON(report)
	case "md":
		return outputMarkdown(report)
	case "raw":
		return outputRaw(report)
	default:
		return outputText(report)
	}
}

func outputJSON(report *model.Report) error {
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

func outputMarkdown(report *model.Report) error {
	var md strings.Builder

	md.WriteString(fmt.Sprintf("# %s - Network Intelligence Report\n\n", report.Target))

	// Basic info
	if len(report.IPv4) > 0 {
		md.WriteString(fmt.Sprintf("**IPs:** %s\n\n", strings.Join(report.IPv4, ", ")))
	}

	// Geolocation
	if report.Geo.Country != "" {
		md.WriteString(fmt.Sprintf("**Location:** %s, %s, %s\n\n", report.Geo.City, report.Geo.Region, report.Geo.Country))
		if report.Geo.ISP != "" {
			md.WriteString(fmt.Sprintf("**ISP:** %s\n\n", report.Geo.ISP))
		}
		if report.Geo.ASN != "" {
			md.WriteString(fmt.Sprintf("**ASN:** %s\n\n", report.Geo.ASN))
		}
	}

	// Ping
	if report.Ping.Success {
		md.WriteString(fmt.Sprintf("**Ping:** %d/%d packets, %s avg\n\n",
			report.Ping.PacketsReceived, report.Ping.PacketsSent, report.Ping.AvgRtt))
	}

	// Ports
	if len(report.Ports.Open) > 0 {
		var ports []string
		for _, p := range report.Ports.Open {
			ports = append(ports, fmt.Sprintf("%d", p))
		}
		md.WriteString(fmt.Sprintf("**Open Ports:** %s\n\n", strings.Join(ports, ", ")))
	}

	// TLS
	if report.TLS.Subject != "" {
		md.WriteString(fmt.Sprintf("**TLS:** %s (expires: %s)\n\n", report.TLS.CommonName, report.TLS.NotAfter))
	}

	// Duration
	md.WriteString(fmt.Sprintf("**Duration:** %dms\n\n", report.DurationMs))

	fmt.Print(md.String())
	return nil
}

func outputRaw(report *model.Report) error {
	jsonData, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Print(string(jsonData))
	return nil
}

func outputText(report *model.Report) error {
	// If no-style is requested or output is being piped,
	// fall back to plain text without ANSI codes.
	if noStyle || !isatty.IsTerminal(os.Stdout.Fd()) {
		return outputPlainText(report)
	}

	// Styles
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")).Padding(0, 1)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Padding(0, 1)
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Padding(0, 1)
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Padding(0, 1)
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	newTable := func(rows ...[]string) *table.Table {
		return table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(borderStyle).
			Width(72).
			Rows(rows...)
	}

	// Basic info table
	infoRows := [][]string{
		{labelStyle.Render("Target"), valueStyle.Render(report.Target)},
		{labelStyle.Render("Duration"), valueStyle.Render(fmt.Sprintf("%dms", report.DurationMs))},
	}

	if len(report.IPv4) > 0 {
		infoRows = append(infoRows, []string{labelStyle.Render("IPv4"), valueStyle.Render(fmt.Sprintf("%v", report.IPv4))})
	}
	if len(report.IPv6) > 0 {
		infoRows = append(infoRows, []string{labelStyle.Render("IPv6"), valueStyle.Render(fmt.Sprintf("%v", report.IPv6))})
	}
	if len(report.Errors) > 0 {
		var keys []string
		for k := range report.Errors {
			keys = append(keys, k)
		}
		// Keep order stable for readability
		sort.Strings(keys)

		for i, k := range keys {
			prefix := ""
			if i == 0 {
				prefix = "Errors" // only show label once
			}
			label := ""
			if prefix != "" {
				label = labelStyle.Render(prefix)
			}
			infoRows = append(infoRows, []string{
				label,
				errorStyle.Render(fmt.Sprintf("%s: %s", k, report.Errors[k])),
			})
		}
	}

	infoTable := newTable(infoRows...)

	fmt.Println(infoTable.Render())
	fmt.Println()

	// Ping statistics
	if report.Ping.PacketsSent > 0 {
		var pingValue string
		if report.Ping.Success {
			pingValue = successStyle.Render(fmt.Sprintf("%d/%d packets", report.Ping.PacketsReceived, report.Ping.PacketsSent)) +
				valueStyle.Render(fmt.Sprintf(", %.1f%% loss, avg %s", report.Ping.PacketLossPct, report.Ping.AvgRtt))
		} else {
			pingValue = errorStyle.Render(fmt.Sprintf("%d/%d packets", report.Ping.PacketsReceived, report.Ping.PacketsSent)) +
				valueStyle.Render(fmt.Sprintf(", %.1f%% loss (failed)", report.Ping.PacketLossPct))
		}

		pingTable := newTable([]string{labelStyle.Render("Ping"), pingValue})

		fmt.Println(pingTable.Render())
		fmt.Println()
	}

	// Traceroute summary
	if report.Trace.Success && len(report.Trace.Hops) > 0 {
		traceValue := valueStyle.Render(fmt.Sprintf("%d hops to %s", len(report.Trace.Hops), report.Target))
		traceTable := newTable([]string{labelStyle.Render("Traceroute"), traceValue})

		fmt.Println(traceTable.Render())
		fmt.Println()
	} else if report.Trace.Error != "" {
		traceErr := errorStyle.Render(report.Trace.Error)
		traceTable := newTable([]string{labelStyle.Render("Traceroute"), traceErr})

		fmt.Println(traceTable.Render())
		fmt.Println()
	}

	// WHOIS info
	if report.Whois.Domain != "" || report.Whois.NetName != "" {
		var whoisValue strings.Builder
		if report.Whois.Domain != "" {
			whoisValue.WriteString(report.Whois.Domain)
			if report.Whois.Registrar != "" {
				whoisValue.WriteString(fmt.Sprintf(" (%s)", report.Whois.Registrar))
			}
			if report.Whois.Expires != "" {
				whoisValue.WriteString(fmt.Sprintf(" expires %s", report.Whois.Expires))
			}
		} else {
			whoisValue.WriteString(report.Whois.NetName)
			if report.Whois.OrgName != "" {
				whoisValue.WriteString(fmt.Sprintf(" (%s)", report.Whois.OrgName))
			}
			if report.Whois.Country != "" {
				whoisValue.WriteString(fmt.Sprintf(" [%s]", report.Whois.Country))
			}
		}

		whoisTable := newTable([]string{labelStyle.Render("WHOIS"), valueStyle.Render(whoisValue.String())})

		fmt.Println(whoisTable.Render())
		fmt.Println()
	} else if report.Errors["whois"] != "" {
		whoisErr := errorStyle.Render(report.Errors["whois"])
		whoisTable := newTable([]string{labelStyle.Render("WHOIS"), whoisErr})

		fmt.Println(whoisTable.Render())
		fmt.Println()
	}

	// ASN / Geo info
	if report.Geo.ASN != "" || report.Geo.City != "" {
		var geoRows [][]string
		if report.Geo.ASN != "" {
			asnValue := report.Geo.ASN
			if report.Geo.ASName != "" {
				asnValue += fmt.Sprintf(" (%s)", report.Geo.ASName)
			}
			if report.Geo.CountryCode != "" {
				asnValue += fmt.Sprintf(" [%s]", report.Geo.CountryCode)
			}
			geoRows = append(geoRows, []string{labelStyle.Render("ASN"), valueStyle.Render(asnValue)})
		}

		if report.Geo.City != "" {
			loc := report.Geo.City
			if report.Geo.Region != "" {
				loc += ", " + report.Geo.Region
			}
			if report.Geo.Country != "" {
				loc += ", " + report.Geo.Country
			}
			if report.Geo.ISP != "" {
				loc += fmt.Sprintf(" (%s)", report.Geo.ISP)
			}
			geoRows = append(geoRows, []string{labelStyle.Render("Location"), valueStyle.Render(loc)})
		}

		if len(geoRows) > 0 {
			geoTable := newTable(geoRows...)

			fmt.Println(geoTable.Render())
			fmt.Println()
		}
	} else if report.Errors["asn"] != "" || report.Errors["geo"] != "" {
		var rows [][]string
		if report.Errors["asn"] != "" {
			rows = append(rows, []string{labelStyle.Render("ASN"), errorStyle.Render(report.Errors["asn"])})
		}
		if report.Errors["geo"] != "" {
			rows = append(rows, []string{labelStyle.Render("Location"), errorStyle.Render(report.Errors["geo"])})
		}
		if len(rows) > 0 {
			geoTable := newTable(rows...)
			fmt.Println(geoTable.Render())
			fmt.Println()
		}
	}

	// Port scan results
	if len(report.Ports.Scanned) > 0 {
		openCount := len(report.Ports.Open)
		closedCount := len(report.Ports.Closed)
		var portsValue string
		if openCount > 0 {
			portsValue = successStyle.Render(fmt.Sprintf("%d open", openCount)) +
				valueStyle.Render(fmt.Sprintf(", %d closed (%v)", closedCount, report.Ports.Open))
		} else {
			portsValue = valueStyle.Render(fmt.Sprintf("%d open, %d closed", openCount, closedCount))
		}

		portsTable := newTable([]string{labelStyle.Render("Ports"), portsValue})

		fmt.Println(portsTable.Render())
	}

	return nil
}

func validateTarget(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", fmt.Errorf("target cannot be empty")
	}
	return target, nil
}

func validateFlags() error {
	// Validate output format
	validOutputs := []string{"text", "md", "json", "raw"}
	valid := false
	for _, v := range validOutputs {
		if output == v {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid output format: %s (valid: %s)", output, strings.Join(validOutputs, ", "))
	}

	// Validate timeout
	if timeout < 1*time.Second || timeout > 5*time.Minute {
		return fmt.Errorf("timeout must be between 1s and 5m")
	}

	// JSON output should generally be piped or redirected
	if output == "json" && isatty.IsTerminal(os.Stdout.Fd()) {
		return fmt.Errorf("JSON output requires piping or file redirection")
	}

	return nil
}

// Config file location: ~/.config/netgaze/config.json
type Config struct {
	DefaultTimeout string `json:"default_timeout"`
	EnablePorts    bool   `json:"enable_ports"`
}

func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "netgaze", "config.json")
}

func loadConfig() (*Config, error) {
	configPath := getConfigPath()
	if configPath == "" {
		return &Config{}, nil
	}

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func saveConfig(config *Config) error {
	configPath := getConfigPath()
	if configPath == "" {
		return fmt.Errorf("cannot determine config directory")
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
