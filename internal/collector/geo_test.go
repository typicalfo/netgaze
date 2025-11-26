package collector

import (
	"context"
	"testing"
	"time"

	"github.com/typicalfo/netgaze/internal/model"
)

func TestCollectGeo(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{
			name:    "valid IP",
			target:  "8.8.8.8",
			wantErr: false, // Should not fail even if geo fails
		},
		{
			name:    "valid domain",
			target:  "google.com",
			wantErr: false, // Should not fail even if geo fails
		},
		{
			name:    "invalid target",
			target:  "nonexistent.invalid.tld",
			wantErr: false, // Should not fail even if geo fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &model.Report{
				Target: tt.target,
				Errors: make(map[string]string),
			}

			err := collectGeo(context.Background(), tt.target, report)

			// Geolocation should not return errors (graceful degradation)
			if err != nil {
				t.Errorf("collectGeo() unexpected error = %v", err)
			}

			// Check that geolocation data structure is initialized
			if report.Geo.IP != "" && report.Errors["geo"] == "" {
				// If we have an IP, we should have some geo data or an error
				if report.Geo.City == "" && report.Geo.Country == "" {
					t.Error("collectGeo() expected some geolocation data")
				}
			}
		})
	}
}

func TestGetGeoTargetIP(t *testing.T) {
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
			ip, err := getGeoTargetIP(tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("getGeoTargetIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && ip == nil {
				t.Error("getGeoTargetIP() expected IP address")
			}
		})
	}
}

func TestLookupGeolocation(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
	}{
		{
			name:    "Google DNS IP",
			ip:      "8.8.8.8",
			wantErr: false,
		},
		{
			name:    "Cloudflare DNS IP",
			ip:      "1.1.1.1",
			wantErr: false,
		},
		{
			name:    "invalid IP",
			ip:      "999.999.999.999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resp, err := lookupGeolocation(ctx, tt.ip)

			if (err != nil) != tt.wantErr {
				t.Errorf("lookupGeolocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && resp == nil {
				t.Error("lookupGeolocation() expected response")
			}

			if !tt.wantErr && resp != nil {
				// Check that response has expected fields
				if resp.Query != tt.ip {
					t.Errorf("lookupGeolocation() query = %v, want %v", resp.Query, tt.ip)
				}
			}
		})
	}
}

func TestPopulateGeoData(t *testing.T) {
	tests := []struct {
		name     string
		response *GeoResponse
		expected struct {
			City        string
			Country     string
			CountryCode string
			ISP         string
			Org         string
			Latitude    float64
			Longitude   float64
		}
	}{
		{
			name: "full response",
			response: &GeoResponse{
				Status:      "success",
				Query:       "8.8.8.8",
				City:        "Mountain View",
				Region:      "California",
				RegionCode:  "CA",
				Country:     "United States",
				CountryCode: "US",
				Lat:         37.4056,
				Lon:         -122.0775,
				Timezone:    "America/Los_Angeles",
				ISP:         "Google LLC",
				Org:         "Google LLC",
				AS:          "AS15169",
			},
			expected: struct {
				City        string
				Country     string
				CountryCode string
				ISP         string
				Org         string
				Latitude    float64
				Longitude   float64
			}{
				City:        "Mountain View",
				Country:     "United States",
				CountryCode: "US",
				ISP:         "Google LLC",
				Org:         "Google LLC",
				Latitude:    37.4056,
				Longitude:   -122.0775,
			},
		},
		{
			name:     "nil response",
			response: nil,
			expected: struct {
				City        string
				Country     string
				CountryCode string
				ISP         string
				Org         string
				Latitude    float64
				Longitude   float64
			}{
				City:        "",
				Country:     "",
				CountryCode: "",
				ISP:         "",
				Org:         "",
				Latitude:    0,
				Longitude:   0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &model.Report{
				Errors: make(map[string]string),
			}

			populateGeoData(tt.response, report)

			if report.Geo.City != tt.expected.City {
				t.Errorf("populateGeoData() city = %v, want %v", report.Geo.City, tt.expected.City)
			}

			if report.Geo.Country != tt.expected.Country {
				t.Errorf("populateGeoData() country = %v, want %v", report.Geo.Country, tt.expected.Country)
			}

			if report.Geo.CountryCode != tt.expected.CountryCode {
				t.Errorf("populateGeoData() countryCode = %v, want %v", report.Geo.CountryCode, tt.expected.CountryCode)
			}

			if report.Geo.ISP != tt.expected.ISP {
				t.Errorf("populateGeoData() ISP = %v, want %v", report.Geo.ISP, tt.expected.ISP)
			}

			if report.Geo.Org != tt.expected.Org {
				t.Errorf("populateGeoData() org = %v, want %v", report.Geo.Org, tt.expected.Org)
			}

			if report.Geo.Latitude != tt.expected.Latitude {
				t.Errorf("populateGeoData() latitude = %v, want %v", report.Geo.Latitude, tt.expected.Latitude)
			}

			if report.Geo.Longitude != tt.expected.Longitude {
				t.Errorf("populateGeoData() longitude = %v, want %v", report.Geo.Longitude, tt.expected.Longitude)
			}
		})
	}
}

func TestCollectGeo_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	report := &model.Report{
		Target: "8.8.8.8",
		Errors: make(map[string]string),
	}

	err := collectGeo(ctx, "8.8.8.8", report)
	// Should not error due to graceful degradation
	if err != nil {
		t.Errorf("collectGeo() unexpected error = %v", err)
	}
}
