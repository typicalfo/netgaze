package collector

import (
	"context"
	"testing"
	"time"

	"github.com/typicalfo/netgaze/internal/model"
)

func TestCollectPorts(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{
			name:    "valid IP",
			target:  "8.8.8.8",
			wantErr: false, // Should not fail even if port scan fails
		},
		{
			name:    "valid domain",
			target:  "google.com",
			wantErr: false, // Should not fail even if port scan fails
		},
		{
			name:    "invalid target",
			target:  "nonexistent.invalid.tld",
			wantErr: false, // Should not fail even if port scan fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &model.Report{
				Target: tt.target,
				Errors: make(map[string]string),
			}

			err := collectPorts(context.Background(), tt.target, report)

			// Port scan should not return errors (graceful degradation)
			if err != nil {
				t.Errorf("collectPorts() unexpected error = %v", err)
			}

			// Check that port scan data structure is initialized
			if len(report.Ports.Scanned) == 0 && report.Errors["ports"] == "" {
				t.Error("collectPorts() expected either scanned ports or error")
			}
		})
	}
}

func TestGetPortsTargetIP(t *testing.T) {
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
			ip, err := getPortsTargetIP(tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("getPortsTargetIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && ip == "" {
				t.Error("getPortsTargetIP() expected IP address")
			}
		})
	}
}

func TestScanPorts(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{
			name:    "localhost - should have some open ports",
			target:  "127.0.0.1",
			wantErr: false,
		},
		{
			name:    "invalid IP",
			target:  "999.999.999.999",
			wantErr: false, // TCP connect fails gracefully, no error returned
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			result, err := scanPorts(ctx, tt.target, &model.Report{})

			if (err != nil) != tt.wantErr {
				t.Errorf("scanPorts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("scanPorts() expected result")
			}

			if !tt.wantErr && result != nil {
				if len(result.Scanned) == 0 {
					t.Error("scanPorts() expected scanned ports")
				}

				if len(result.Open) > len(result.Scanned) {
					t.Error("scanPorts() open ports cannot exceed scanned ports")
				}

				if len(result.Closed) > len(result.Scanned) {
					t.Error("scanPorts() closed ports cannot exceed scanned ports")
				}
			}
		})
	}
}

func TestGetCommonPorts(t *testing.T) {
	ports := getCommonPorts()

	// Should return exactly 20 ports
	if len(ports) != 20 {
		t.Errorf("getCommonPorts() returned %d ports, want 20", len(ports))
	}

	// Check for specific expected ports
	expectedPorts := []int{22, 53, 80, 443, 8080}
	for _, expected := range expectedPorts {
		found := false
		for _, port := range ports {
			if port == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("getCommonPorts() missing expected port %d", expected)
		}
	}

	// Check that ports are in valid range
	for _, port := range ports {
		if port < 1 || port > 65535 {
			t.Errorf("getCommonPorts() invalid port number: %d", port)
		}
	}
}

func TestPopulatePortData(t *testing.T) {
	tests := []struct {
		name     string
		result   *PortScanResult
		expected struct {
			Scanned  int
			Open     int
			Closed   int
			Filtered int
		}
	}{
		{
			name: "full result",
			result: &PortScanResult{
				Scanned:  []int{22, 80, 443},
				Open:     []int{22, 80},
				Closed:   []int{443},
				Filtered: []int{},
			},
			expected: struct {
				Scanned  int
				Open     int
				Closed   int
				Filtered int
			}{
				Scanned:  3,
				Open:     2,
				Closed:   1,
				Filtered: 0,
			},
		},
		{
			name:   "nil result",
			result: nil,
			expected: struct {
				Scanned  int
				Open     int
				Closed   int
				Filtered int
			}{
				Scanned:  0,
				Open:     0,
				Closed:   0,
				Filtered: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &model.Report{
				Errors: make(map[string]string),
			}

			populatePortData(tt.result, report)

			if len(report.Ports.Scanned) != tt.expected.Scanned {
				t.Errorf("populatePortData() scanned = %d, want %d", len(report.Ports.Scanned), tt.expected.Scanned)
			}

			if len(report.Ports.Open) != tt.expected.Open {
				t.Errorf("populatePortData() open = %d, want %d", len(report.Ports.Open), tt.expected.Open)
			}

			if len(report.Ports.Closed) != tt.expected.Closed {
				t.Errorf("populatePortData() closed = %d, want %d", len(report.Ports.Closed), tt.expected.Closed)
			}

			if len(report.Ports.Filtered) != tt.expected.Filtered {
				t.Errorf("populatePortData() filtered = %d, want %d", len(report.Ports.Filtered), tt.expected.Filtered)
			}
		})
	}
}

func TestCollectPorts_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	report := &model.Report{
		Target: "8.8.8.8",
		Errors: make(map[string]string),
	}

	err := collectPorts(ctx, "8.8.8.8", report)
	// Should not error due to graceful degradation
	if err != nil {
		t.Errorf("collectPorts() unexpected error = %v", err)
	}
}
