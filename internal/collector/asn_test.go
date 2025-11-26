package collector

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/typicalfo/netgaze/internal/model"
)

func TestCollectASN(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{
			name:    "valid IP",
			target:  "8.8.8.8",
			wantErr: false, // Should not fail even if ASN fails
		},
		{
			name:    "valid domain",
			target:  "google.com",
			wantErr: false, // Should not fail even if ASN fails
		},
		{
			name:    "invalid target",
			target:  "nonexistent.invalid.tld",
			wantErr: false, // Should not fail even if ASN fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &model.Report{
				Target: tt.target,
				Errors: make(map[string]string),
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := collectASN(ctx, tt.target, report)

			// ASN should not return errors (graceful degradation)
			if err != nil {
				t.Errorf("collectASN() unexpected error = %v", err)
			}

			// Check that ASN data structure is initialized
			if report.Geo.ASN != "" && report.Geo.IP == "" {
				t.Error("collectASN() ASN set but IP not set")
			}
		})
	}
}

func TestGetTargetIP(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{
			name:    "valid IPv4",
			target:  "8.8.8.8",
			wantErr: false,
		},
		{
			name:    "valid IPv6",
			target:  "2001:db8::1",
			wantErr: false,
		},
		{
			name:    "valid domain",
			target:  "google.com",
			wantErr: false,
		},
		{
			name:    "invalid domain",
			target:  "nonexistent.invalid.tld",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip, err := getTargetIP(tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("getTargetIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && ip == nil {
				t.Error("getTargetIP() expected IP address")
			}
		})
	}
}

func TestReverseIP(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		want    string
		wantErr bool
	}{
		{
			name:    "IPv4 address",
			ip:      "8.8.8.8",
			want:    "8.8.8.8",
			wantErr: false,
		},
		{
			name:    "IPv4 different octets",
			ip:      "192.168.1.1",
			want:    "1.1.168.192",
			wantErr: false,
		},
		{
			name:    "invalid IP",
			ip:      "invalid",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil && !tt.wantErr {
				t.Fatalf("Invalid test IP: %s", tt.ip)
			}

			got, err := reverseIP(ip)

			if (err != nil) != tt.wantErr {
				t.Errorf("reverseIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("reverseIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTeamCymruResult(t *testing.T) {
	tests := []struct {
		name     string
		result   string
		expected struct {
			ASN     string
			IP      string
			Country string
			ASName  string
		}
	}{
		{
			name:   "Google ASN",
			result: "15169 | 8.8.8.8 | 8.8.8.0/24 | US | arin | 2012-03-30 | GOOGLE-CLOUD-PLATFORM",
			expected: struct {
				ASN     string
				IP      string
				Country string
				ASName  string
			}{
				ASN:     "15169",
				IP:      "8.8.8.8",
				Country: "US",
				ASName:  "GOOGLE-CLOUD-PLATFORM",
			},
		},
		{
			name:   "Cloudflare ASN",
			result: "13335 | 1.1.1.1 | 1.1.1.0/24 | US | arin | 2010-07-14 | CLOUDFLARENET",
			expected: struct {
				ASN     string
				IP      string
				Country string
				ASName  string
			}{
				ASN:     "13335",
				IP:      "1.1.1.1",
				Country: "US",
				ASName:  "CLOUDFLARENET",
			},
		},
		{
			name:   "invalid format",
			result: "invalid data",
			expected: struct {
				ASN     string
				IP      string
				Country string
				ASName  string
			}{
				ASN:     "",
				IP:      "",
				Country: "",
				ASName:  "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &model.Report{
				Errors: make(map[string]string),
			}

			parseTeamCymruResult(tt.result, report)

			if report.Geo.ASN != tt.expected.ASN {
				t.Errorf("parseTeamCymruResult() ASN = %v, want %v", report.Geo.ASN, tt.expected.ASN)
			}

			if report.Geo.IP != tt.expected.IP {
				t.Errorf("parseTeamCymruResult() IP = %v, want %v", report.Geo.IP, tt.expected.IP)
			}

			if report.Geo.CountryCode != tt.expected.Country {
				t.Errorf("parseTeamCymruResult() Country = %v, want %v", report.Geo.CountryCode, tt.expected.Country)
			}

			if report.Geo.ASName != tt.expected.ASName {
				t.Errorf("parseTeamCymruResult() ASName = %v, want %v", report.Geo.ASName, tt.expected.ASName)
			}

			if report.Geo.Org != tt.expected.ASName {
				t.Errorf("parseTeamCymruResult() Org = %v, want %v", report.Geo.Org, tt.expected.ASName)
			}
		})
	}
}

func TestCollectASN_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	report := &model.Report{
		Target: "8.8.8.8",
		Errors: make(map[string]string),
	}

	err := collectASN(ctx, "8.8.8.8", report)
	// Should not error due to graceful degradation
	if err != nil {
		t.Errorf("collectASN() unexpected error = %v", err)
	}
}
