# Addendum 14 – Error Handling & Graceful Degradation Strategy

**Philosophy**: Never fail completely. Always provide partial results with clear error communication.

**Error Classification**
1. **Critical Errors** - Stop collection entirely
2. **Collector Errors** - Store in Report.Errors, continue other collectors
3. **Warnings** - Non-critical issues, log but don't affect output

**Critical Errors (stop execution)**
- DNS resolution failure (required dependency for other collectors)
- Context timeout (global 15s exceeded)
- Invalid target format
- Missing required permissions for basic operations

**Collector Error Handling**
```go
// Pattern for all collectors
func collectX(ctx context.Context, target string, r *model.Report) error {
    // Try collection with timeout
    ctx, cancel := context.WithTimeout(ctx, collectorTimeout)
    defer cancel()
    
    result, err := doCollection(ctx, target)
    if err != nil {
        // Store error but don't fail entire collection
        r.Errors["collector_name"] = formatError(err)
        return nil // Continue with other collectors
    }
    
    // Store successful results
    r.CollectorField = result
    return nil
}
```

**Error Message Standards**
- User-friendly, not technical
- Include timeout information
- Suggest remediation when helpful
- Consistent formatting across collectors

**Specific Error Scenarios**

**DNS Errors**
```
"dns": "all DNS lookups failed or timed out"
"dns": "DNS server timeout"
"dns": "invalid hostname format"
```

**Ping Errors**
```
"ping": "ping timeout after 5s"
"ping": "permission denied (try running as administrator)"
"ping": "network unreachable"
```

**Port Scan Errors**
```
"ports": "port scan permission denied (run with sudo for SYN scan)"
"ports": "port scan timeout after 10s"
"ports": "network unreachable during scan"
```

**TLS Errors**
```
"tls": "TLS connection refused"
"tls": "TLS handshake timeout"
"tls": "no TLS certificates presented"
"tls": "certificate parsing failed"
```

**Geolocation/ASN Errors**
```
"geo": "geolocation API request failed"
"geo": "geolocation rate limited"
"asn": "ASN lookup failed"
"asn": "ASN lookup timeout"
```

**WHOIS Errors**
```
"whois": "WHOIS query timeout"
"whois": "no WHOIS data available"
"whois": "WHOIS server unreachable"
```

**Traceroute Errors**
```
"traceroute": "traceroute timeout after 30 hops"
"traceroute": "permission denied (try running as administrator)"
"traceroute": "network unreachable"
```

**Graceful Degradation Examples**

**Scenario 1: Corporate Network**
- DNS: Success
- Ping: "ICMP blocked by firewall"
- Traceroute: "traceroute blocked by firewall"
- WHOIS: Success
- ASN/Geo: Success
- Ports: Success (filtered ports shown)
- TLS: Success
- **Result**: Full report minus connectivity data

**Scenario 2: Remote Host Timeout**
- DNS: Success
- Ping: "ping timeout after 5s"
- Traceroute: "traceroute timeout after 30 hops"
- WHOIS: Success
- ASN/Geo: Success
- Ports: "port scan timeout after 10s"
- TLS: Skipped (no port 443 scan)
- **Result**: Basic network info, no service availability

**Scenario 3: Rate Limiting**
- DNS: Success
- Ping: Success
- Traceroute: Success
- WHOIS: "WHOIS rate limited"
- ASN/Geo: "geolocation rate limited"
- Ports: Success
- TLS: Success
- **Result**: Technical data without third-party enrichment

**Error Display in TUI**
- Red indicator for failed collectors
- Hover over indicator shows error message
- Summary tab shows warning count
- Raw data tab includes error details

**Error Display in Templates**
- Warnings section at bottom of output
- Collector name + brief error message
- No technical stack traces
- Consistent format across output types

**Logging Strategy**
```go
// Structured logging for debugging
logger.Debug("collector_failed", 
    "collector", "ping",
    "target", target,
    "error", err,
    "duration", time.Since(start),
)
```

**Recovery Mechanisms**
- Automatic fallback to unprivileged modes
- Retry once for transient network errors
- Alternative data sources when primary fails
- Cache successful results during single run

**User Experience Principles**
- Always show something useful
- Be transparent about limitations
- Provide actionable error messages
- Maintain consistent performance even with partial failures