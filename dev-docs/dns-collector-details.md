# Addendum 2 – DNS Collector Details

**Responsibility**: Populate all DNS-related fields in `Report`:
- `IPs`, `IPv4`, `IPv6`
- `PTR`
- `CNAME`
- `MX`
- `NS`
- `TXT`

**File**: `internal/collector/dns.go`

**Behavior**
- Single function: `collectDNS(ctx context.Context, target string, r *model.Report)`
- Runs entirely with Go stdlib (`net` package) unless enhanced resolution is needed
- All lookups are parallel inside the function using `sync.WaitGroup` or `errgroup`
- Individual timeouts per lookup type: 3 seconds each
- Graceful: any failed lookup simply leaves that slice empty (no error in `Errors` map unless all fail)

**Resolution strategy**
```go
// 1. If target is IP → skip forward lookups, do reverse only
// 2. If target is hostname/URL → forward resolve first
```

**Exact lookup order and handling**
| Record | Function                  | Result handling                                      |
|--------|---------------------------|------------------------------------------------------|
| A      | net.ResolveIPAddr("ip4", host) | append to IPs + IPv4                              |
| AAAA   | net.ResolveIPAddr("ip6", host) | append to IPs + IPv6                              |
| CNAME  | net.LookupCNAME(host)     | store chain (last non-CNAME wins for A/AAAA)         |
| MX     | net.LookupMX(host)        | "15 mx.example.com" format, sorted by preference     |
| NS     | net.LookupNS(host)        | "ns1.example.com" only                               |
| TXT    | net.LookupTXT(host)       | raw strings, joined if multiple                      |
| PTR    | net.LookupAddr(ip)        | only on final resolved IPs (deduped), multi-value OK |

**Special rules**
- Preserve resolution chain: if CNAME exists, store full chain in `CNAME` slice
- Deduplicate IPs across A and AAAA
- For URLs: strip scheme and path (`http://example.com/path` → `example.com`)
- Always perform reverse DNS on every distinct resolved IP (even if input was IP)
- Use system resolver by default (`net.DefaultResolver`)
- Respect context timeout/cancel from parent collector

**Error handling**
- Individual lookup failure → silently skip that record type
- Total DNS failure (no IPs resolved and input was hostname) → add to `r.Errors["dns"] = "all DNS lookups failed or timed out"`

**Performance target**
- Complete in ≤ 3 seconds even on slow networks
- Maximum 6 concurrent lookups at peak

**No external dependencies** – pure stdlib (no miekg/dns required for MVP)
