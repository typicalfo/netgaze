package model

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

type Report struct {
	// Input & metadata
	Target     string    `json:"target"`      // original user input
	ResolvedAt time.Time `json:"resolved_at"` // UTC timestamp when collection finished
	DurationMs int64     `json:"duration_ms"`

	// DNS resolution
	IPs   []net.IP `json:"ips,omitempty"` // A + AAAA
	IPv4  []string `json:"ipv4,omitempty"`
	IPv6  []string `json:"ipv6,omitempty"`
	PTR   []string `json:"ptr,omitempty"` // reverse DNS
	CNAME []string `json:"cname,omitempty"`
	MX    []string `json:"mx,omitempty"` // preference host
	NS    []string `json:"ns,omitempty"`
	TXT   []string `json:"txt,omitempty"`

	// Geolocation & ASN
	Geo struct {
		IP          string  `json:"ip,omitempty"`
		City        string  `json:"city,omitempty"`
		Region      string  `json:"region,omitempty"`
		RegionCode  string  `json:"region_code,omitempty"`
		Country     string  `json:"country,omitempty"`
		CountryCode string  `json:"country_code,omitempty"`
		Org         string  `json:"org,omitempty"`
		ISP         string  `json:"isp,omitempty"`
		ASN         string  `json:"asn,omitempty"`
		ASName      string  `json:"as_name,omitempty"`
		Latitude    float64 `json:"lat,omitempty"`
		Longitude   float64 `json:"lon,omitempty"`
		Timezone    string  `json:"timezone,omitempty"`
	} `json:"geo"`

	// WHOIS (raw + parsed top fields)
	WhoisRaw string `json:"whois_raw,omitempty"`
	Whois    struct {
		Domain      string   `json:"domain,omitempty"`
		Registrar   string   `json:"registrar,omitempty"`
		Created     string   `json:"created,omitempty"`
		Expires     string   `json:"expires,omitempty"`
		Registrant  string   `json:"registrant,omitempty"`
		NetRange    string   `json:"net_range,omitempty"`
		NetName     string   `json:"net_name,omitempty"`
		OrgName     string   `json:"org_name,omitempty"`
		Country     string   `json:"country,omitempty"`
		AbuseEmails []string `json:"abuse_emails,omitempty"`
	} `json:"whois"`

	// Ping
	Ping struct {
		PacketsSent     int     `json:"sent"`
		PacketsReceived int     `json:"received"`
		PacketLossPct   float64 `json:"loss_percent"`
		MinRtt          string  `json:"min_rtt"` // e.g. "12.4ms"
		AvgRtt          string  `json:"avg_rtt"`
		MaxRtt          string  `json:"max_rtt"`
		StdDevRtt       string  `json:"stddev_rtt,omitempty"`
		Success         bool    `json:"success"`
		Error           string  `json:"error,omitempty"`
	} `json:"ping"`

	// Traceroute
	Trace struct {
		Hops    []TraceHop `json:"hops,omitempty"`
		Success bool       `json:"success"`
		Error   string     `json:"error,omitempty"`
	}

	// Port scan (only when --ports)
	Ports struct {
		Scanned  []int  `json:"scanned_ports,omitempty"`
		Open     []int  `json:"open_ports,omitempty"`
		Closed   []int  `json:"closed_ports,omitempty"`
		Filtered []int  `json:"filtered_ports,omitempty"`
		Error    string `json:"error,omitempty"`
	} `json:"ports,omitempty"`

	// TLS certificate (443 only, opportunistic)
	TLS struct {
		Subject    string   `json:"subject,omitempty"`
		Issuer     string   `json:"issuer,omitempty"`
		CommonName string   `json:"cn,omitempty"`
		AltNames   []string `json:"sans,omitempty"`
		NotBefore  string   `json:"valid_from,omitempty"`
		NotAfter   string   `json:"valid_until,omitempty"`
		Expired    bool     `json:"expired"`
		SelfSigned bool     `json:"self_signed"`
		Error      string   `json:"error,omitempty"`
	} `json:"tls,omitempty"`

	// Errors from individual collectors (for graceful degradation)
	Errors map[string]string `json:"collector_errors,omitempty"` // key = collector name
}

// Helper types
type TraceHop struct {
	Hop     int    `json:"hop"`
	IP      string `json:"ip,omitempty"`
	Host    string `json:"host,omitempty"`
	RTT     string `json:"rtt,omitempty"` // "12.4ms" or "*"
	Timeout bool   `json:"timeout,omitempty"`
}

// ValidateTarget validates and normalizes the input target
func ValidateTarget(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", fmt.Errorf("target cannot be empty")
	}

	// Check if it's a URL
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		u, err := url.Parse(target)
		if err != nil {
			return "", fmt.Errorf("invalid URL: %w", err)
		}
		return u.Hostname(), nil
	}

	// Check if it's an IP address
	if net.ParseIP(target) != nil {
		return target, nil
	}

	// Assume it's a hostname/domain
	// Accept any non-empty string that doesn't contain spaces
	if !strings.Contains(target, " ") {
		return target, nil
	}

	return "", fmt.Errorf("invalid target: must be IP address, domain, or URL")
}

// ToJSON converts Report to JSON bytes
func (r *Report) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// FromJSON creates Report from JSON bytes
func FromJSON(data []byte) (*Report, error) {
	var report Report
	err := json.Unmarshal(data, &report)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return &report, nil
}
