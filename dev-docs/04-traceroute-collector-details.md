# Addendum 4 – Traceroute Collector Details

**Responsibility**: Populate `Report.Trace` with hop-by-hop path analysis.

**File**: `internal/collector/traceroute.go`

**Behavior**
- Single function: `collectTraceroute(ctx context.Context, target string, r *model.Report)`
- Uses `github.com/pixelbender/go-traceroute` for UDP-based traceroute
- Max 30 hops, 3 probes per hop, 2s timeout per probe
- Target selection: same priority as ping (IPv4 → IPv6 → original)

**Exact configuration**
```go
opts := traceroute.Options{
    MaxHops:     30,
    ProbesPerHop: 3,
    Timeout:     2 * time.Second,
    Port:        33434, // starting UDP port
}
```

**Hop data structure**
For each successful hop:
```go
hop := TraceHop{
    Hop:     hopNumber,
    IP:      responderIP.String(),
    Host:    reverseLookup(responderIP), // optional, best effort
    RTT:     formatDuration(rtt), // "12.4ms" or "*" for timeout
    Timeout: rtt == 0,
}
```

**Output handling**
- Include all hops with at least one successful probe
- Use average RTT from successful probes
- `*` for completely timed out hops
- Stop early if destination reached

**Error handling**
- Permission denied → try unprivileged mode
- Network unreachable → set `Error` and `Success = false`
- Timeout → set `Error = "traceroute timeout after 30 hops"`
- Add to `r.Errors["traceroute"]` only on complete failure

**Platform considerations**
- Linux/macOS: UDP traceroute with fallback to TCP if needed
- Windows: uses ICMP-based traceroute (different library behavior)
- Corporate networks: may show * * * for firewalled hops (expected)

**Performance target**
- Complete in ≤ 10 seconds
- Graceful degradation with partial results acceptable