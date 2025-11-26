package collector

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/typicalfo/netgaze/internal/model"
)

func collectPorts(ctx context.Context, target string, report *model.Report) error {
	// Create independent context for port scan to avoid cancellation by other collectors
	// Use 30 second timeout for port scan specifically
	portCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get IP address from target
	ip, err := getPortsTargetIP(target)
	if err != nil {
		report.Errors["ports"] = fmt.Sprintf("Failed to resolve target for port scan: %v", err)
		// Don't return error for port scan - it's optional
		return nil
	}

	// Run port scan in goroutine to respect context
	resultChan := make(chan *PortScanResult, 1)
	errorChan := make(chan error, 1)

	go func() {
		result, err := scanPorts(portCtx, ip, report)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	// Wait for completion or timeout
	select {
	case result := <-resultChan:
		populatePortData(result, report)
	case err := <-errorChan:
		report.Errors["ports"] = fmt.Sprintf("Port scan failed: %v", err)
		// Don't return error for port scan - it's optional
		return nil
	case <-portCtx.Done():
		report.Errors["ports"] = "Port scan timeout"
		// Don't return error for port scan - it's optional
		return nil
	}

	return nil
}

func getPortsTargetIP(target string) (string, error) {
	// If target is already an IP, return it
	if ip := net.ParseIP(target); ip != nil {
		return ip.String(), nil
	}

	// Otherwise resolve using DNS
	ips, err := net.LookupIP(target)
	if err != nil {
		return "", err
	}

	if len(ips) == 0 {
		return "", fmt.Errorf("no IP addresses found for %s", target)
	}

	// Prefer IPv4 for port scanning
	for _, ip := range ips {
		if ip.To4() != nil {
			return ip.String(), nil
		}
	}

	// Fall back to IPv6 if no IPv4 available
	return ips[0].String(), nil
}

func scanPorts(ctx context.Context, target string, report *model.Report) (*PortScanResult, error) {
	result := &PortScanResult{
		Scanned: getCommonPorts(),
	}

	// Simple TCP connect scan for each port
	for _, port := range getCommonPorts() {
		select {
		case <-ctx.Done():
			return result, fmt.Errorf("scan timeout")
		default:
			// Try to connect to port
			address := fmt.Sprintf("%s:%d", target, port)
			conn, err := net.DialTimeout("tcp", address, 1*time.Second)

			if err == nil {
				// Port is open
				result.Open = append(result.Open, port)
				conn.Close()
			} else {
				// Port is closed or filtered
				result.Closed = append(result.Closed, port)
			}
		}
	}

	return result, nil
}

func getCommonPorts() []int {
	// Exact list from dev-docs/13-exact-common-ports-list.md
	return []int{
		22,    // SSH
		53,    // DNS
		80,    // HTTP
		110,   // POP3
		135,   // RPC
		139,   // NetBIOS
		143,   // IMAP
		443,   // HTTPS
		993,   // IMAPS
		995,   // POP3S
		1723,  // PPTP
		3306,  // MySQL
		3389,  // RDP
		445,   // SMB
		5900,  // VNC
		8080,  // HTTP-Alt
		8443,  // HTTPS-Alt
		992,   // TelnetS
		10000, // Webmin
		1433,  // MSSQL
	}
}

func populatePortData(result *PortScanResult, report *model.Report) {
	if result == nil {
		return
	}

	report.Ports.Scanned = result.Scanned
	report.Ports.Open = result.Open
	report.Ports.Closed = result.Closed
	// Note: Simple scan doesn't distinguish filtered from closed
	report.Ports.Filtered = []int{} // Empty for now
}

// PortScanResult represents the result of a port scan
type PortScanResult struct {
	Scanned  []int
	Open     []int
	Closed   []int
	Filtered []int
}
