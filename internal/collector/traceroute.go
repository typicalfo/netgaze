package collector

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/typicalfo/netgaze/internal/model"
)

func collectTraceroute(ctx context.Context, target string, report *model.Report) error {
	hops, err := Traceroute(ctx, target, 20*time.Second)
	if err != nil {
		report.Errors["traceroute"] = fmt.Sprintf("Traceroute failed: %v", err)
		report.Trace.Error = err.Error()
		// Don't return error for traceroute - it's optional
		return nil
	}

	report.Trace.Hops = hops
	report.Trace.Success = len(hops) > 0

	return nil
}

// Traceroute runs a traceroute for the given target and returns the hop list.
// It is used by both the collector and CLI subcommands.
func Traceroute(ctx context.Context, target string, timeout time.Duration) ([]model.TraceHop, error) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ip, err := resolveTargetIP(target)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target: %w", err)
	}

	return runSystemTraceroute(ctx, ip.String())
}

func runSystemTraceroute(ctx context.Context, target string) ([]model.TraceHop, error) {
	// Try different traceroute commands based on OS
	var cmd *exec.Cmd

	// On macOS/Linux, use traceroute with -n flag (no DNS resolution) for speed
	cmd = exec.CommandContext(ctx, "traceroute", "-n", "-m", "15", "-w", "3", target)

	// Run command
	output, err := cmd.Output()
	if err != nil {
		// Try without flags as fallback
		cmd = exec.CommandContext(ctx, "traceroute", target)
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("traceroute command failed: %w", err)
		}
	}

	return parseTracerouteOutput(string(output))
}

func resolveTargetIP(target string) (net.IP, error) {
	// If target is already an IP, return it
	if ip := net.ParseIP(target); ip != nil {
		return ip, nil
	}

	// Otherwise resolve using DNS
	ips, err := net.LookupIP(target)
	if err != nil {
		return nil, err
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no IP addresses found for %s", target)
	}

	// Prefer IPv4 for traceroute
	for _, ip := range ips {
		if ip.To4() != nil {
			return ip, nil
		}
	}

	// Fall back to IPv6 if no IPv4 available
	return ips[0], nil
}

func parseTracerouteOutput(output string) ([]model.TraceHop, error) {
	var hops []model.TraceHop
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "traceroute to") {
			continue
		}

		// Parse hop line
		// Format: "1  gateway (192.168.1.1)  1.234 ms  1.567 ms  1.890 ms"
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		hopNum, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}

		traceHop := model.TraceHop{
			Hop: hopNum,
		}

		// Extract IP address and RTT
		for j := 1; j < len(fields); j++ {
			field := fields[j]

			// Check if field looks like an IP address
			if net.ParseIP(field) != nil {
				traceHop.IP = field
				continue
			}

			// Check if field looks like RTT (ends with "ms")
			if strings.HasSuffix(field, "ms") {
				rttStr := strings.TrimSuffix(field, "ms")
				if rtt, err := strconv.ParseFloat(rttStr, 64); err == nil {
					traceHop.RTT = fmt.Sprintf("%.1fms", rtt)
					break // Use first RTT value
				}
			}
		}

		// Try to resolve hostname if we have IP
		if traceHop.IP != "" && traceHop.Host == "" {
			if names, err := net.LookupAddr(traceHop.IP); err == nil && len(names) > 0 {
				traceHop.Host = names[0]
			}
		}

		hops = append(hops, traceHop)
	}

	return hops, nil
}
