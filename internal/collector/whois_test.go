package collector

import (
	"context"
	"testing"
	"time"

	"github.com/typicalfo/netgaze/internal/model"
)

func TestCollectWhois(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{
			name:    "valid domain",
			target:  "example.com",
			wantErr: false, // Should not fail even if WHOIS fails
		},
		{
			name:    "valid IP",
			target:  "8.8.8.8",
			wantErr: false, // Should not fail even if WHOIS fails
		},
		{
			name:    "invalid target",
			target:  "nonexistent.invalid.tld",
			wantErr: false, // Should not fail even if WHOIS fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &model.Report{
				Target: tt.target,
				Errors: make(map[string]string),
			}

			err := collectWhois(context.Background(), tt.target, report)

			// WHOIS should not return errors (graceful degradation)
			if err != nil {
				t.Errorf("collectWhois() unexpected error = %v", err)
			}

			// Check that WHOIS data structure is initialized
			if report.WhoisRaw == "" && report.Errors["whois"] == "" {
				t.Error("collectWhois() expected either raw data or error")
			}
		})
	}
}

func TestParseWhoisData(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		expected struct {
			Domain    string
			Registrar string
			Created   string
			Expires   string
		}
	}{
		{
			name: "domain WHOIS",
			data: `Domain Name: EXAMPLE.COM
Registry Domain ID: 2336799_DOMAIN_COM-VRSN
Registrar WHOIS Server: whois.registrar.com
Registrar URL: http://www.registrar.com
Updated Date: 2023-08-14T07:15:28Z
Creation Date: 1995-08-14T04:00:00Z
Registry Expiry Date: 2024-08-13T04:00:00Z
Registrar: Example Registrar
Registrar Abuse Contact Email: abuse@example.com`,
			expected: struct {
				Domain    string
				Registrar string
				Created   string
				Expires   string
			}{
				Domain:    "EXAMPLE.COM",
				Registrar: "Example Registrar",
				Created:   "1995-08-14T04:00:00Z",
				Expires:   "2024-08-13T04:00:00Z",
			},
		},
		{
			name: "IP WHOIS",
			data: `inetnum:        8.8.8.0 - 8.8.8.255
netname:        GOOGLE-CLOUD
descr:          Google LLC
country:        US
admin-c:        GCN46-RIPE
tech-c:         GCN46-RIPE
status:         ASSIGNED PA
mnt-by:         RIPE-NCC-HM-MNT
mnt-lower:      GOOGLE-MNT
created:        2022-09-01T09:53:11Z
last-modified:  2022-09-01T09:53:11Z
source:         RIPE`,
			expected: struct {
				Domain    string
				Registrar string
				Created   string
				Expires   string
			}{
				Domain:    "",
				Registrar: "",
				Created:   "2022-09-01T09:53:11Z",
				Expires:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &model.Report{
				Errors: make(map[string]string),
			}

			parseWhoisData(tt.data, report)

			if report.Whois.Domain != tt.expected.Domain {
				t.Errorf("parseWhoisData() domain = %v, want %v", report.Whois.Domain, tt.expected.Domain)
			}

			if report.Whois.Registrar != tt.expected.Registrar {
				t.Errorf("parseWhoisData() registrar = %v, want %v", report.Whois.Registrar, tt.expected.Registrar)
			}

			if report.Whois.Created != tt.expected.Created {
				t.Errorf("parseWhoisData() created = %v, want %v", report.Whois.Created, tt.expected.Created)
			}

			if report.Whois.Expires != tt.expected.Expires {
				t.Errorf("parseWhoisData() expires = %v, want %v", report.Whois.Expires, tt.expected.Expires)
			}
		})
	}
}

func TestExtractField(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		patterns []string
		want     string
	}{
		{
			name:     "domain extraction",
			data:     "Domain Name: EXAMPLE.COM",
			patterns: []string{`domain name:\s*(.+)`},
			want:     "EXAMPLE.COM",
		},
		{
			name:     "registrar extraction",
			data:     "Registrar: Example Registrar Inc.",
			patterns: []string{`registrar:\s*(.+)`},
			want:     "Example Registrar Inc.",
		},
		{
			name:     "multiple patterns",
			data:     "Creation Date: 2023-01-01",
			patterns: []string{`created:\s*(.+)`, `creation date:\s*(.+)`},
			want:     "2023-01-01",
		},
		{
			name:     "no match",
			data:     "Some other text",
			patterns: []string{`domain:\s*(.+)`},
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractField(tt.data, tt.data, tt.patterns)
			if got != tt.want {
				t.Errorf("extractField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractEmails(t *testing.T) {
	tests := []struct {
		name string
		data string
		want []string
	}{
		{
			name: "abuse emails",
			data: "Registrar Abuse Contact Email: abuse@example.com\nAdmin Email: admin@example.org",
			want: []string{"abuse@example.com", "admin@example.org"},
		},
		{
			name: "technical emails",
			data: "Technical Contact: tech@company.net\nOrg Email: info@company.com",
			want: []string{"tech@company.net", "info@company.com"},
		},
		{
			name: "no emails",
			data: "No email addresses here",
			want: []string{},
		},
		{
			name: "duplicate emails",
			data: "Email: test@example.com\nEmail: TEST@EXAMPLE.COM",
			want: []string{"test@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractEmails(tt.data)
			if len(got) != len(tt.want) {
				t.Errorf("extractEmails() length = %d, want %d", len(got), len(tt.want))
				return
			}

			for i, email := range got {
				if email != tt.want[i] {
					t.Errorf("extractEmails()[%d] = %v, want %v", i, email, tt.want[i])
				}
			}
		})
	}
}

func TestCollectWhois_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	report := &model.Report{
		Target: "example.com",
		Errors: make(map[string]string),
	}

	err := collectWhois(ctx, "example.com", report)
	// Should not error due to graceful degradation
	if err != nil {
		t.Errorf("collectWhois() unexpected error = %v", err)
	}
}
