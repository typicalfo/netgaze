package model

import (
	"strings"
	"testing"
	"time"
)

func TestValidateTarget(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		want    string
		wantErr bool
	}{
		{
			name:    "valid IPv4",
			target:  "1.1.1.1",
			want:    "1.1.1.1",
			wantErr: false,
		},
		{
			name:    "valid IPv6",
			target:  "2001:db8::1",
			want:    "2001:db8::1",
			wantErr: false,
		},
		{
			name:    "valid domain",
			target:  "example.com",
			want:    "example.com",
			wantErr: false,
		},
		{
			name:    "valid URL HTTP",
			target:  "http://example.com/path",
			want:    "example.com",
			wantErr: false,
		},
		{
			name:    "valid URL HTTPS",
			target:  "https://sub.example.com/path?query=1",
			want:    "sub.example.com",
			wantErr: false,
		},
		{
			name:    "empty target",
			target:  "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			target:  "http://",
			want:    "",
			wantErr: false, // Will be treated as hostname
		},
		{
			name:    "invalid IP",
			target:  "999.999.999.999",
			want:    "999.999.999.999",
			wantErr: false, // Will be treated as hostname
		},
		{
			name:    "whitespace",
			target:  "  example.com  ",
			want:    "example.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateTarget(tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateTarget() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReport_ToJSON(t *testing.T) {
	report := &Report{
		Target:     "example.com",
		ResolvedAt: time.Now().UTC(),
		DurationMs: 5000,
		IPv4:       []string{"93.184.216.34"},
	}

	data, err := report.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Verify it's valid JSON
	if !strings.Contains(string(data), `"target": "example.com"`) {
		t.Error("JSON output missing target field")
	}
}

func TestFromJSON(t *testing.T) {
	jsonData := `{
		"target": "example.com",
		"resolved_at": "2023-01-01T00:00:00Z",
		"duration_ms": 5000,
		"ipv4": ["93.184.216.34"]
	}`

	report, err := FromJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	if report.Target != "example.com" {
		t.Errorf("FromJSON() target = %v, want %v", report.Target, "example.com")
	}
}

func TestFromJSON_Invalid(t *testing.T) {
	invalidJSON := `{"target": "example.com", "invalid": }`

	_, err := FromJSON([]byte(invalidJSON))
	if err == nil {
		t.Error("FromJSON() expected error for invalid JSON")
	}
}
