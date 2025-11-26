package collector

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/typicalfo/netgaze/internal/model"
)

func collectTLS(ctx context.Context, target string, report *model.Report) error {
	// Check if port 443 is open from port scan results
	if !isPortOpen(report.Ports.Open, 443) {
		report.Errors["tls"] = "Port 443 not open - skipping TLS collection"
		return nil
	}

	// Create context with 4-second timeout
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	// Run TLS collection in goroutine to respect context
	resultChan := make(chan *TLSResult, 1)
	errorChan := make(chan error, 1)

	go func() {
		result, err := getTLSCertificate(ctx, target)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	// Wait for completion or timeout
	select {
	case result := <-resultChan:
		populateTLSData(result, report)
	case err := <-errorChan:
		report.Errors["tls"] = fmt.Sprintf("TLS collection failed: %v", err)
		// Don't return error for TLS collection - it's optional
		return nil
	case <-ctx.Done():
		report.Errors["tls"] = "TLS collection timeout"
		// Don't return error for TLS collection - it's optional
		return nil
	}

	return nil
}

func isPortOpen(openPorts []int, port int) bool {
	for _, p := range openPorts {
		if p == port {
			return true
		}
	}
	return false
}

func getTLSCertificate(ctx context.Context, target string) (*TLSResult, error) {
	hostname := extractHostname(target)

	// Create dialer with timeout
	dialer := &net.Dialer{Timeout: 4 * time.Second}

	// Configure TLS connection
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // we want cert even if invalid
		ServerName:         hostname,
	}

	// Connect with TLS
	address := fmt.Sprintf("%s:443", hostname)
	conn, err := tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("TLS connection failed: %w", err)
	}
	defer conn.Close()

	// Get certificate chain
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil, fmt.Errorf("no TLS certificates presented")
	}

	// Extract leaf certificate information
	cert := certs[0]
	result := &TLSResult{
		Subject:    cert.Subject.String(),
		Issuer:     cert.Issuer.String(),
		CommonName: cert.Subject.CommonName,
		AltNames:   cert.DNSNames,
		NotBefore:  cert.NotBefore.Format(time.RFC3339),
		NotAfter:   cert.NotAfter.Format(time.RFC3339),
		Expired:    time.Now().After(cert.NotAfter),
		SelfSigned: cert.Issuer.CommonName == cert.Subject.CommonName,
	}

	return result, nil
}

func extractHostname(target string) string {
	// If target is a URL, extract hostname
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		u, err := url.Parse(target)
		if err == nil {
			return u.Hostname()
		}
	}

	// If target contains port, extract just the host part
	if strings.Contains(target, ":") && !strings.Contains(target, "]") {
		host, _, err := net.SplitHostPort(target)
		if err == nil {
			return host
		}
	}

	// Return target as-is (domain or IP)
	return target
}

func populateTLSData(result *TLSResult, report *model.Report) {
	if result == nil {
		return
	}

	report.TLS.Subject = result.Subject
	report.TLS.Issuer = result.Issuer
	report.TLS.CommonName = result.CommonName
	report.TLS.AltNames = result.AltNames
	report.TLS.NotBefore = result.NotBefore
	report.TLS.NotAfter = result.NotAfter
	report.TLS.Expired = result.Expired
	report.TLS.SelfSigned = result.SelfSigned
}

// TLSResult represents the result of TLS certificate collection
type TLSResult struct {
	Subject    string
	Issuer     string
	CommonName string
	AltNames   []string
	NotBefore  string
	NotAfter   string
	Expired    bool
	SelfSigned bool
}
