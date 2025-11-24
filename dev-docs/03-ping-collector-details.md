# Addendum 3 – Ping Collector Details

**Responsibility**: Populate `Report.Ping` with ICMP echo statistics.

**File**: `internal/collector/ping.go`

**Behavior**
- Single function: `collectPing(ctx context.Context, target string, r *model.Report)`
- Uses `github.com/prometheus-community/pro-bing` for cross-platform ICMP
- Sends exactly 5 packets with 1s timeout each
- Target is the first resolved IPv4 if available, otherwise IPv6, otherwise original input
- Total timeout: 5 seconds maximum

**Exact configuration**
```go
pinger, err := probing.NewPinger(target)
pinger.Count = 5
pinger.Timeout = 5 * time.Second
pinger.Interval = 1 * time.Second
pinger.SetPrivileged(true) // fallback to unprivileged if fails
```

**Statistics calculation**
- `PacketsSent`: always 5
- `PacketsReceived`: actual responses received
- `PacketLossPct`: ((sent - received) / sent) * 100
- RTT values: converted to "12.4ms" string format
- `Success`: true if at least 1 response received
- `Error`: set on total failure (no responses, permission denied, etc.)

**Error handling**
- Permission denied (requires root) → try unprivileged mode
- Network unreachable → set `Error` and `Success = false`
- Timeout → set `Error = "ping timeout after 5s"`
- Add to `r.Errors["ping"]` only on complete failure

**Platform notes**
- Linux/macOS: privileged ICMP first, fallback to unprivileged
- Windows: uses native ICMP API (no special privileges needed)
- All platforms: same 5-packet behavior

**Performance target**
- Complete in ≤ 5 seconds
- Fail fast if target clearly unreachable