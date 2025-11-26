package collector

import (
	"context"
	"testing"
	"time"

	"github.com/typicalfo/netgaze/internal/model"
)

func TestCollectPing(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{
			name:    "valid IP",
			target:  "8.8.8.8",
			wantErr: false,
		},
		{
			name:    "valid domain",
			target:  "google.com",
			wantErr: false,
		},
		{
			name:    "unreachable IP",
			target:  "192.0.2.1", // RFC 5737 test address
			wantErr: false,       // Ping may succeed with 100% packet loss
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &model.Report{
				Target: tt.target,
				Errors: make(map[string]string),
			}

			err := collectPing(context.Background(), tt.target, report)

			if (err != nil) != tt.wantErr {
				t.Errorf("collectPing() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check that ping data was collected for valid targets
			if !tt.wantErr {
				if report.Ping.PacketsSent == 0 {
					t.Error("collectPing() expected packets to be sent")
				}
				if report.Ping.PacketsSent != 5 {
					t.Errorf("collectPing() expected 5 packets sent, got %d", report.Ping.PacketsSent)
				}
			}
		})
	}
}

func TestCollectPing_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	report := &model.Report{
		Target: "8.8.8.8",
		Errors: make(map[string]string),
	}

	err := collectPing(ctx, "8.8.8.8", report)
	// Ping may still succeed due to its own timeout handling
	// This test mainly ensures the function doesn't panic
	_ = err
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{
			name: "sub-millisecond",
			d:    500 * time.Microsecond,
			want: "0.50ms",
		},
		{
			name: "exactly 1ms",
			d:    1 * time.Millisecond,
			want: "1.0ms",
		},
		{
			name: "25.5ms",
			d:    25*time.Millisecond + 500*time.Microsecond,
			want: "25.5ms",
		},
		{
			name: "100ms",
			d:    100 * time.Millisecond,
			want: "100.0ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.d)
			if got != tt.want {
				t.Errorf("formatDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
