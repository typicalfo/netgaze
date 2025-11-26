package collector

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/typicalfo/netgaze/internal/model"
)

func TestCollectDNS(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{
			name:    "valid domain",
			target:  "example.com",
			wantErr: false,
		},
		{
			name:    "valid IP",
			target:  "8.8.8.8",
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
			report := &model.Report{
				Target: tt.target,
				Errors: make(map[string]string),
			}

			err := collectDNS(context.Background(), tt.target, report)

			if (err != nil) != tt.wantErr {
				t.Errorf("collectDNS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check that some DNS data was collected for valid targets
			if !tt.wantErr {
				if len(report.IPs) == 0 {
					t.Error("collectDNS() expected at least one IP address")
				}
			}
		})
	}
}

func TestResolveIPs(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{
			name:    "example.com",
			target:  "example.com",
			wantErr: false,
		},
		{
			name:    "google.com",
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
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			ips, err := resolveIPs(ctx, tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("resolveIPs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(ips) == 0 {
				t.Error("resolveIPs() expected at least one IP address")
			}
		})
	}
}

func TestResolvePTR(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
	}{
		{
			name:    "Google DNS",
			ip:      "8.8.8.8",
			wantErr: false,
		},
		{
			name:    "Cloudflare DNS",
			ip:      "1.1.1.1",
			wantErr: false,
		},
		{
			name:    "localhost",
			ip:      "127.0.0.1",
			wantErr: false, // localhost may have PTR or may not, both are valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("Invalid IP address: %s", tt.ip)
			}

			names, err := resolvePTR(ctx, ip)

			if (err != nil) != tt.wantErr {
				t.Errorf("resolvePTR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// PTR may return empty results even without error
			_ = names // Use the variable to avoid unused error
		})
	}
}

func TestCollectDNS_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	report := &model.Report{
		Target: "example.com",
		Errors: make(map[string]string),
	}

	err := collectDNS(ctx, "example.com", report)
	if err == nil {
		t.Error("collectDNS() expected timeout error")
	}
}
