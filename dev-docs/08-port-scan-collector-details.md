# Addendum 8 – Port Scan Collector Details

**Responsibility**: Populate `Report.Ports` with open/closed/filtered port information.

**File**: `internal/collector/ports.go`

**Behavior**
- Only runs when `--ports` flag is explicitly specified (not default)
- Single function: `collectPorts(ctx context.Context, target string, r *model.Report)`
- Uses `github.com/projectdiscovery/naabu` for fast TCP SYN scanning
- 10 second timeout
- Scans top 20 common ports only

**Common ports list (hard-coded)**
```
22,53,80,110,135,139,143,443,993,995,1723,3306,3389,5900,8080,8443,10000,1433,445,992
```

**Exact configuration**
```go
scanner, err := naabu.NewScanner(&naabu.ScannerOptions{
    Host:    target,
    Ports:   commonPortsList,
    Timeout: 10 * time.Second,
    Retries: 1,
    Rate:    1000, // packets per second
    ScanType: naabu.SYN, // SYN scan if privileged, fallback to TCP
})
```

**Port classification**
- Open: successful connection or SYN-ACK response
- Closed: RST response or immediate connection refused
- Filtered: timeout or no response
- Error: scan initialization failure

**Result mapping**
```go
r.Ports.Scanned = commonPortsList
r.Ports.Open = openPorts
r.Ports.Closed = closedPorts
r.Ports.Filtered = filteredPorts
```

**Error handling**
- Permission denied (SYN scan needs root) → fallback to TCP connect scan
- Network unreachable → set `Error = "port scan network unreachable"`
- Timeout → set `Error = "port scan timeout"`
- Add to `r.Errors["ports"]` on complete failure

**Platform considerations**
- Linux/macOS: SYN scan with root, TCP connect without
- Windows: TCP connect scan only (no raw sockets)
- Corporate networks: many ports may show as filtered

**Performance target**
- Complete in ≤ 10 seconds for 20 ports
- Fast failure on clearly unreachable hosts
- Rate limiting to avoid triggering IDS/IPS

**Security considerations**
- Non-invasive scanning (only common ports)
- Respectful scanning rate (1000 packets/sec max)
- No aggressive techniques that could be flagged

**Graceful degradation**
- Partial results acceptable (some ports timeout)
- Continue with other collectors even if scan fails