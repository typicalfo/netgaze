package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/typicalfo/netgaze/internal/model"
)

func collectGeo(ctx context.Context, target string, report *model.Report) error {
	// Create context with 8-second timeout
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	// Get IP address from target
	ip, err := getGeoTargetIP(target)
	if err != nil {
		report.Errors["geo"] = fmt.Sprintf("Failed to resolve target for geolocation: %v", err)
		// Don't return error for geolocation - it's optional
		return nil
	}

	// Run geolocation lookup in goroutine to respect context
	resultChan := make(chan *GeoResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		resp, err := lookupGeolocation(ctx, ip.String())
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- resp
	}()

	// Wait for completion or timeout
	select {
	case resp := <-resultChan:
		populateGeoData(resp, report)
	case err := <-errorChan:
		report.Errors["geo"] = fmt.Sprintf("Geolocation lookup failed: %v", err)
		// Don't return error for geolocation - it's optional
		return nil
	case <-ctx.Done():
		report.Errors["geo"] = "Geolocation lookup timeout"
		// Don't return error for geolocation - it's optional
		return nil
	}

	return nil
}

func getGeoTargetIP(target string) (net.IP, error) {
	// If target is already an IP, return it
	if ip := net.ParseIP(target); ip != nil {
		return ip, nil
	}

	// Otherwise resolve using DNS
	ips, err := net.LookupIP(target)
	if err != nil {
		return nil, err
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no IP addresses found for %s", target)
	}

	// Prefer IPv4 for geolocation
	for _, ip := range ips {
		if ip.To4() != nil {
			return ip, nil
		}
	}

	// Fall back to IPv6 if no IPv4 available
	return ips[0], nil
}

func lookupGeolocation(ctx context.Context, ip string) (*GeoResponse, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	// Make request to ip-api.com
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set HTTP headers
	req.Header.Set("User-Agent", "netgaze/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status: %d", resp.StatusCode)
	}

	// Parse JSON response
	var geoResp GeoResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoResp); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Check API response status
	if strings.ToLower(geoResp.Status) != "success" {
		return nil, fmt.Errorf("API error: %s", geoResp.Message)
	}

	return &geoResp, nil
}

func populateGeoData(resp *GeoResponse, report *model.Report) {
	if resp == nil {
		return
	}

	// Populate geolocation information
	report.Geo.IP = resp.Query
	report.Geo.City = resp.City
	report.Geo.Region = resp.Region
	report.Geo.RegionCode = resp.RegionCode
	report.Geo.Country = resp.Country
	report.Geo.CountryCode = resp.CountryCode
	report.Geo.Org = resp.Org
	report.Geo.ISP = resp.ISP
	report.Geo.Latitude = resp.Lat
	report.Geo.Longitude = resp.Lon
	report.Geo.Timezone = resp.Timezone

	// If we have ASN info from other collector, preserve it
	if report.Geo.ASN == "" && resp.AS != "" {
		report.Geo.ASN = resp.AS
	}
}

// GeoResponse represents the response from ip-api.com
type GeoResponse struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionCode  string  `json:"regionCode"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Message     string  `json:"message"`
	Query       string  `json:"query"`
}
