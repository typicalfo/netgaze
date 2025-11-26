package collector

import (
	"context"
	"testing"
	"time"

	"github.com/typicalfo/netgaze/internal/model"
)

func TestCollectTraceroute(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{
			name:    "valid IP",
			target:  "8.8.8.8",
			wantErr: false, // Should not fail even if traceroute fails
		},
		{
			name:    "valid domain",
			target:  "google.com",
			wantErr: false, // Should not fail even if traceroute fails
		},
		{
			name:    "invalid target",
			target:  "nonexistent.invalid.tld",
			wantErr: false, // Should not fail even if traceroute fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &model.Report{
				Target: tt.target,
				Errors: make(map[string]string),
			}

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			err := collectTraceroute(ctx, tt.target, report)

			// Traceroute should not return errors (graceful degradation)
			if err != nil {
				t.Errorf("collectTraceroute() unexpected error = %v", err)
			}

			// Check that traceroute data structure is initialized
			if report.Trace.Success && len(report.Trace.Hops) == 0 {
				t.Error("collectTraceroute() success=true but no hops")
			}
		})
	}
}

func TestResolveTargetIP(t *testing.T) {
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
			ip, err := resolveTargetIP(tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("resolveTargetIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && ip == nil {
				t.Error("resolveTargetIP() expected IP address")
			}
		})
	}
}

func TestParseTracerouteOutput(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		wantErr bool
		wantLen int
	}{
		{
			name: "valid output",
			output: `traceroute to 8.8.8.8 (8.8.8.8), 30 hops max, 60 byte packets
 1  192.168.1.1 (192.168.1.1)  1.234 ms  1.567 ms  1.890 ms
 2  10.0.0.1 (10.0.0.1)  5.123 ms  5.456 ms  5.789 ms
 3  8.8.8.8 (8.8.8.8)  10.123 ms  10.456 ms  10.789 ms`,
			wantErr: false,
			wantLen: 3,
		},
		{
			name:    "empty output",
			output:  "",
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "output with timeouts",
			output: `traceroute to 8.8.8.8 (8.8.8.8), 30 hops max, 60 byte packets
 1  * * *
 2  192.168.1.1 (192.168.1.1)  1.234 ms  1.567 ms  1.890 ms`,
			wantErr: false,
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hops, err := parseTracerouteOutput(tt.output)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseTracerouteOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(hops) != tt.wantLen {
				t.Errorf("parseTracerouteOutput() hops = %d, want %d", len(hops), tt.wantLen)
			}
		})
	}
}

func TestCollectTraceroute_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	report := &model.Report{
		Target: "8.8.8.8",
		Errors: make(map[string]string),
	}

	err := collectTraceroute(ctx, "8.8.8.8", report)
	// Should not error due to graceful degradation
	if err != nil {
		t.Errorf("collectTraceroute() unexpected error = %v", err)
	}
}
