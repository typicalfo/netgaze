package collector

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/typicalfo/netgaze/internal/model"
)

func collectASN(ctx context.Context, target string, report *model.Report) error {
	// Create context with 8-second timeout
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	// Get IP address from target
	ip, err := getTargetIP(target)
	if err != nil {
		report.Errors["asn"] = fmt.Sprintf("Failed to resolve target for ASN lookup: %v", err)
		// Don't return error for ASN - it's optional
		return nil
	}

	// Perform Team Cymru DNS lookup
	resultChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	go func() {
		result, err := lookupTeamCymru(ip)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	// Wait for completion or timeout
	select {
	case result := <-resultChan:
		parseTeamCymruResult(result, report)
	case err := <-errorChan:
		report.Errors["asn"] = fmt.Sprintf("ASN DNS lookup failed: %v", err)
		// Don't return error for ASN - it's optional
		return nil
	case <-ctx.Done():
		report.Errors["asn"] = "ASN DNS lookup timeout"
		// Don't return error for ASN - it's optional
		return nil
	}

	return nil
}

func getTargetIP(target string) (net.IP, error) {
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

	// Prefer IPv4 for ASN lookup
	for _, ip := range ips {
		if ip.To4() != nil {
			return ip, nil
		}
	}

	// Fall back to IPv6 if no IPv4 available
	return ips[0], nil
}

func lookupTeamCymru(ip net.IP) (string, error) {
	// Reverse IP for DNS lookup
	reversedIP, err := reverseIP(ip)
	if err != nil {
		return "", err
	}

	// Query Team Cymru DNS
	query := fmt.Sprintf("%s.origin.asn.cymru.com", reversedIP)
	txtRecords, err := net.LookupTXT(query)
	if err != nil {
		return "", err
	}

	if len(txtRecords) == 0 {
		return "", fmt.Errorf("no TXT records found")
	}

	return txtRecords[0], nil
}

func reverseIP(ip net.IP) (string, error) {
	if ip.To4() != nil {
		// IPv4
		ipv4 := ip.To4()
		return fmt.Sprintf("%d.%d.%d.%d", ipv4[3], ipv4[2], ipv4[1], ipv4[0]), nil
	}

	// IPv6 (simplified - full implementation would be more complex)
	ipv6 := ip.To16()
	if ipv6 == nil {
		return "", fmt.Errorf("invalid IP address")
	}

	// For IPv6, we'll use a simplified approach
	// In practice, you'd want to implement full nibble-wise reversal
	return fmt.Sprintf("%s", ip.String()), nil
}

func parseTeamCymruResult(result string, report *model.Report) {
	// Parse Team Cymru format: "ASN | IP | BGP Prefix | Country | Registry | Allocated | AS Name"
	// Example: "15169 | 8.8.8.8 | 8.8.8.0/24 | US | arin | 2012-03-30 | GOOGLE-CLOUD-PLATFORM"

	parts := strings.Split(result, " | ")
	if len(parts) < 7 {
		return
	}

	// Extract ASN
	asn := strings.TrimSpace(parts[0])
	if asn != "" {
		report.Geo.ASN = asn
	}

	// Extract country
	country := strings.TrimSpace(parts[3])
	if country != "" {
		report.Geo.CountryCode = country
	}

	// Extract AS name
	asName := strings.TrimSpace(parts[6])
	if asName != "" {
		report.Geo.ASName = asName
		report.Geo.Org = asName
	}

	// Store the IP being queried
	if len(parts) > 1 {
		report.Geo.IP = strings.TrimSpace(parts[1])
	}
}
