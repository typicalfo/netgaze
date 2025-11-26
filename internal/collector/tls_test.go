package collector

import (
	"context"
	"testing"
	"time"

	"github.com/typicalfo/netgaze/internal/model"
)

func TestCollectTLS(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{
			name:    "HTTPS domain with port 443 open",
			target:  "google.com",
			wantErr: false, // Should not fail even if TLS collection fails
		},
		{
			name:    "HTTPS IP with port 443 open",
			target:  "8.8.8.8",
			wantErr: false, // Should not fail even if TLS collection fails
		},
		{
			name:    "invalid target",
			target:  "nonexistent.invalid.tld",
			wantErr: false, // Should not fail even if TLS collection fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &model.Report{
				Target: tt.target,
				Errors: make(map[string]string),
			}
			report.Ports.Open = []int{443} // Simulate port 443 being open

			err := collectTLS(context.Background(), tt.target, report)

			// TLS collection should not return errors (graceful degradation)
			if err != nil {
				t.Errorf("collectTLS() unexpected error = %v", err)
			}

			// Check that TLS data structure is initialized or error is set
			if report.TLS.Subject == "" && report.Errors["tls"] == "" {
				t.Error("collectTLS() expected either TLS data or error")
			}
		})
	}
}

func TestCollectTLS_Port443Closed(t *testing.T) {
	report := &model.Report{
		Target: "google.com",
		Errors: make(map[string]string),
	}
	report.Ports.Open = []int{80, 22} // Port 443 not in open list

	err := collectTLS(context.Background(), "google.com", report)

	// Should not error
	if err != nil {
		t.Errorf("collectTLS() unexpected error = %v", err)
	}

	// Should have error about port 443 not being open
	if report.Errors["tls"] != "Port 443 not open - skipping TLS collection" {
		t.Errorf("collectTLS() expected port 443 error, got: %s", report.Errors["tls"])
	}
}

func TestIsPortOpen(t *testing.T) {
	tests := []struct {
		name      string
		openPorts []int
		port      int
		expected  bool
	}{
		{
			name:      "port is open",
			openPorts: []int{22, 80, 443},
			port:      443,
			expected:  true,
		},
		{
			name:      "port is not open",
			openPorts: []int{22, 80},
			port:      443,
			expected:  false,
		},
		{
			name:      "empty open ports list",
			openPorts: []int{},
			port:      443,
			expected:  false,
		},
		{
			name:      "single port match",
			openPorts: []int{443},
			port:      443,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPortOpen(tt.openPorts, tt.port)
			if result != tt.expected {
				t.Errorf("isPortOpen() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractHostname(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		expected string
	}{
		{
			name:     "plain domain",
			target:   "example.com",
			expected: "example.com",
		},
		{
			name:     "plain IP",
			target:   "8.8.8.8",
			expected: "8.8.8.8",
		},
		{
			name:     "HTTPS URL",
			target:   "https://example.com/path",
			expected: "example.com",
		},
		{
			name:     "HTTP URL",
			target:   "http://example.com/path",
			expected: "example.com",
		},
		{
			name:     "domain with port",
			target:   "example.com:8080",
			expected: "example.com",
		},
		{
			name:     "IP with port",
			target:   "8.8.8.8:53",
			expected: "8.8.8.8",
		},
		{
			name:     "HTTPS URL with port",
			target:   "https://example.com:8443/path",
			expected: "example.com",
		},
		{
			name:     "IPv6 address",
			target:   "2001:db8::1",
			expected: "2001:db8::1",
		},
		{
			name:     "IPv6 with port (should not split)",
			target:   "[2001:db8::1]:443",
			expected: "[2001:db8::1]:443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractHostname(tt.target)
			if result != tt.expected {
				t.Errorf("extractHostname() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetTLSCertificate(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{
			name:    "valid HTTPS domain",
			target:  "google.com",
			wantErr: false,
		},
		{
			name:    "valid HTTPS IP",
			target:  "1.1.1.1",
			wantErr: false,
		},
		{
			name:    "invalid domain",
			target:  "nonexistent.invalid.tld",
			wantErr: true,
		},
		{
			name:    "invalid IP",
			target:  "999.999.999.999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			result, err := getTLSCertificate(ctx, tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("getTLSCertificate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("getTLSCertificate() expected result")
			}

			if !tt.wantErr && result != nil {
				// Validate result structure
				if result.Subject == "" {
					t.Error("getTLSCertificate() expected subject")
				}
				if result.Issuer == "" {
					t.Error("getTLSCertificate() expected issuer")
				}
				if result.NotBefore == "" {
					t.Error("getTLSCertificate() expected notBefore")
				}
				if result.NotAfter == "" {
					t.Error("getTLSCertificate() expected notAfter")
				}
			}
		})
	}
}

func TestPopulateTLSData(t *testing.T) {
	tests := []struct {
		name     string
		result   *TLSResult
		expected struct {
			Subject    string
			Issuer     string
			CommonName string
			AltNames   int
			Expired    bool
			SelfSigned bool
		}
	}{
		{
			name: "full result",
			result: &TLSResult{
				Subject:    "CN=example.com",
				Issuer:     "CN=Let's Encrypt Authority X3",
				CommonName: "example.com",
				AltNames:   []string{"example.com", "www.example.com"},
				NotBefore:  "2023-01-01T00:00:00Z",
				NotAfter:   "2024-01-01T00:00:00Z",
				Expired:    false,
				SelfSigned: false,
			},
			expected: struct {
				Subject    string
				Issuer     string
				CommonName string
				AltNames   int
				Expired    bool
				SelfSigned bool
			}{
				Subject:    "CN=example.com",
				Issuer:     "CN=Let's Encrypt Authority X3",
				CommonName: "example.com",
				AltNames:   2,
				Expired:    false,
				SelfSigned: false,
			},
		},
		{
			name:   "nil result",
			result: nil,
			expected: struct {
				Subject    string
				Issuer     string
				CommonName string
				AltNames   int
				Expired    bool
				SelfSigned bool
			}{
				Subject:    "",
				Issuer:     "",
				CommonName: "",
				AltNames:   0,
				Expired:    false,
				SelfSigned: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &model.Report{
				Errors: make(map[string]string),
			}

			populateTLSData(tt.result, report)

			if report.TLS.Subject != tt.expected.Subject {
				t.Errorf("populateTLSData() subject = %v, want %v", report.TLS.Subject, tt.expected.Subject)
			}

			if report.TLS.Issuer != tt.expected.Issuer {
				t.Errorf("populateTLSData() issuer = %v, want %v", report.TLS.Issuer, tt.expected.Issuer)
			}

			if report.TLS.CommonName != tt.expected.CommonName {
				t.Errorf("populateTLSData() commonName = %v, want %v", report.TLS.CommonName, tt.expected.CommonName)
			}

			if len(report.TLS.AltNames) != tt.expected.AltNames {
				t.Errorf("populateTLSData() altNames = %v, want %v", len(report.TLS.AltNames), tt.expected.AltNames)
			}

			if report.TLS.Expired != tt.expected.Expired {
				t.Errorf("populateTLSData() expired = %v, want %v", report.TLS.Expired, tt.expected.Expired)
			}

			if report.TLS.SelfSigned != tt.expected.SelfSigned {
				t.Errorf("populateTLSData() selfSigned = %v, want %v", report.TLS.SelfSigned, tt.expected.SelfSigned)
			}
		})
	}
}

func TestCollectTLS_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	report := &model.Report{
		Target: "google.com",
		Errors: make(map[string]string),
	}
	report.Ports.Open = []int{443} // Simulate port 443 being open

	err := collectTLS(ctx, "google.com", report)
	// Should not error due to graceful degradation
	if err != nil {
		t.Errorf("collectTLS() unexpected error = %v", err)
	}
}
