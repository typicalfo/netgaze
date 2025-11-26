# Addendum 12 – Template Contents

**Responsibility**: Define exact templates for non-AI output formatting.

**Files**: `internal/templates/summary.txt`, `internal/templates/summary.md`, `internal/templates/raw.txt`

**Template system**
- Go `text/template` with `//go:embed` for compile-time inclusion
- All templates receive the same `model.Report` struct
- Lipgloss styling applied at runtime for terminal output
- Consistent data across all formats

**Template 1: summary.txt (colored terminal)**
```go
//go:embed summary.txt
var summaryText string

{{- $geo := .Geo -}}
{{- $ping := .Ping -}}
{{- $whois := .Whois -}}
{{- $ports := .Ports -}}
{{- $tls := .TLS -}}

{{.Target}} - Network Intelligence Report
{{if .IPs}}IPs: {{range $i, $ip := .IPv4}}{{if $i}}, {{end}}{{$ip}}{{end}}{{if .IPv6}}{{if .IPv4}}, {{end}}{{range $i, $ip := .IPv6}}{{if $i}}, {{end}}{{$ip}}{{end}}{{end}}{{end}}
{{if $geo.Country}}Location: {{$geo.City}}, {{$geo.Region}}, {{$geo.Country}} ({{$geo.ISP}}){{end}}
{{if $geo.ASN}}ASN: {{$geo.ASN}} - {{$geo.ASName}}{{end}}

{{if $ping.Success}}Ping: {{$ping.PacketsReceived}}/{{$ping.PacketsSent}} packets, {{$ping.AvgRtt}} avg{{end}}
{{if $whois.Domain}}Domain: {{$whois.Domain}} ({{$whois.Registrar}}){{end}}
{{if $whois.NetRange}}NetRange: {{$whois.NetRange}} ({{$whois.OrgName}}){{end}}

{{if $ports.Open}}Open Ports: {{range $i, $port := $ports.Open}}{{if $i}}, {{end}}{{$port}}{{end}}{{end}}
{{if $tls.Subject}}TLS: {{$tls.CommonName}} (expires: {{$tls.NotAfter}}){{if $tls.Expired}} [EXPIRED]{{end}}{{end}}

Duration: {{.DurationMs}}ms
{{if .Errors}}Warnings: {{range $collector, $error := .Errors}}{{$collector}}: {{$error}} {{end}}{{end}}
```

**Template 2: summary.md (GitHub-flavored markdown)**
```go
//go:embed summary.md
var summaryMarkdown string

# Network Intelligence Report: {{.Target}}

**Generated:** {{.ResolvedAt}} | **Duration:** {{.DurationMs}}ms | **Mode:** {{if .ModeNoAgent}}No-AI{{else}}AI{{end}}

## Network Information
{{if .IPs}}**IP Addresses:** {{range $i, $ip := .IPv4}}{{if $i}}, {{end}}`{{$ip}}`{{end}}{{if .IPv6}}{{if .IPv4}}, {{end}}{{range $i, $ip := .IPv6}}{{if $i}}, {{end}}`{{$ip}}`{{end}}{{end}}{{end}}
{{if .PTR}}**Reverse DNS:** {{range $i, $ptr := .PTR}}{{if $i}}, {{end}}`{{$ptr}}`{{end}}{{end}}

{{if $geo.Country}}## Geolocation
- **Location:** {{$geo.City}}, {{$geo.Region}}, {{$geo.Country}}
- **Coordinates:** {{$geo.Latitude}}, {{$geo.Longitude}}
- **ISP:** {{$geo.ISP}}
- **Organization:** {{$geo.Org}}
{{if $geo.ASN}}- **ASN:** [{{$geo.ASN}}](https://bgp.he.net/AS{{$geo.ASN}}) - {{$geo.ASName}}{{end}}
{{end}}

{{if $ping.Success}}## Connectivity
- **Ping:** {{$ping.PacketsReceived}}/{{$ping.PacketsSent}} packets received
- **Packet Loss:** {{printf "%.1f" $ping.PacketLossPct}}%
- **RTT:** Min {{$ping.MinRtt}}, Avg {{$ping.AvgRtt}}, Max {{$ping.MaxRtt}}
{{end}}

{{if $whois.Domain}}## Domain Information
- **Domain:** {{$whois.Domain}}
- **Registrar:** {{$whois.Registrar}}
- **Created:** {{$whois.Created}}
- **Expires:** {{$whois.Expires}}
{{end}}

{{if $whois.NetRange}}## Network Information
- **NetRange:** {{$whois.NetRange}}
- **Network Name:** {{$whois.NetName}}
- **Organization:** {{$whois.OrgName}}
{{end}}

{{if $ports.Open}}## Port Scan Results
**Open Ports:** {{range $i, $port := $ports.Open}}{{if $i}}, {{end}}`{{$port}}`{{end}}
{{if $ports.Closed}}**Closed Ports:** {{range $i, $port := $ports.Closed}}{{if $i}}, {{end}}`{{$port}}`{{end}}{{end}}
{{if $ports.Filtered}}**Filtered Ports:** {{range $i, $port := $ports.Filtered}}{{if $i}}, {{end}}`{{$port}}`{{end}}{{end}}
{{end}}

{{if $tls.Subject}}## TLS Certificate
- **Subject:** {{$tls.Subject}}
- **Issuer:** {{$tls.Issuer}}
- **Common Name:** {{$tls.CommonName}}
- **Valid From:** {{$tls.NotBefore}}
- **Valid Until:** {{$tls.NotAfter}}
- **Status:** {{if $tls.Expired}}Expired{{else if $tls.SelfSigned}}Self-signed{{else}}Valid{{end}}
{{if $tls.AltNames}}- **SANs:** {{range $i, $san := $tls.AltNames}}{{if $i}}, {{end}}`{{$san}}`{{end}}{{end}}
{{end}}

{{if .Errors}}## Warnings
{{range $collector, $error := .Errors}}- **{{$collector}}:** {{$error}}
{{end}}{{end}}
```

**Template 3: raw.txt (minimal one-liner format)**
```go
//go:embed raw.txt
var rawText string

{{.Target}}|{{range $i, $ip := .IPv4}}{{if $i}},{{end}}{{$ip}}{{end}}{{if .IPv6}}{{if .IPv4}},{{end}}{{range $i, $ip := .IPv6}}{{if $i}},{{end}}{{$ip}}{{end}}{{end}}|{{.Geo.Country}}|{{.Geo.City}}|{{.Geo.ISP}}|{{.Geo.ASN}}|{{if .Ping.Success}}{{.Ping.AvgRtt}}{{else}}failed{{end}}|{{range $i, $port := .Ports.Open}}{{if $i}},{{end}}{{$port}}{{end}}|{{if .TLS.CommonName}}{{.TLS.CommonName}}{{end}}|{{.DurationMs}}ms
```

**Template loading and usage**
```go
// internal/templates/templates.go
package templates

import (
    "embed"
    "text/template"
)

//go:embed *.txt *.md
var templateFS embed.FS

var (
    SummaryText    = parseTemplate("summary.txt", summaryText)
    SummaryMarkdown = parseTemplate("summary.md", summaryMarkdown)
    RawText        = parseTemplate("raw.txt", rawText)
)

func parseTemplate(name, content string) *template.Template {
    return template.Must(template.New(name).Parse(content))
}
```

**Styling for terminal output**
- Use lipgloss for colors and formatting
- Green for successful data, red for errors, yellow for warnings
- Bold headers, dim for secondary information
- Consistent color scheme across all terminal output

**Template functions**
- Custom template functions for formatting (duration, file size, etc.)
- Conditional rendering with `{{if}}` blocks
- Loop support for arrays (IPs, ports, etc.)
- String formatting with `printf`

**Output consistency**
- All templates show the same core data
- Identical field names and values across formats
- Same error handling and warnings
- Consistent timestamp and duration formatting