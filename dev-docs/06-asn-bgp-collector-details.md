# Addendum 6 – ASN / BGP Collector Details

**Responsibility**: Populate `Report.Geo.ASN`, `Report.Geo.ASName`, and related BGP fields.

**File**: `internal/collector/asn.go`

**Behavior**
- Single function: `collectASN(ctx context.Context, target string, r *model.Report)`
- Uses Team Cymru DNS-based ASN lookup via `github.com/ammario/ipisp`
- 3 second timeout
- Queries for first resolved IPv4, then IPv6 if needed

**Team Cymru DNS query format**
```
# For IPv4: 1.2.3.4 → 4.3.2.1.origin.asn.cymru.com
# For IPv6: reverse nibble format + origin6.asn.cymru.com
```

**Expected response parsing**
```
"197074 | 197074 | 208.84.121.0/24 | US | arin | 2020-01-28"
# Format: ASN | AS-NAME | IP-BLOCK | COUNTRY | RIR | ALLOC-DATE
```

**Field mapping**
```go
r.Geo.ASN = response.ASN
r.Geo.ASName = response.ASName
r.Geo.Org = response.ASName // fallback for geolocation
r.Geo.CountryCode = response.CountryCode
```

**Library usage**
```go
client := ipisp.NewDNSClient()
info, err := client.LookupIP(ctx, net.ParseIP(ip))
```

**Error handling**
- DNS resolution failure → set `Error = "ASN lookup failed"`
- No ASN data → leave fields empty, no error (some IPs have no ASN)
- Timeout → set `Error = "ASN lookup timeout"`
- Add to `r.Errors["asn"]` only on complete failure

**Fallback strategy**
- If Team Cymru fails, try ip-api.com (same as geolocation)
- If both fail, leave ASN fields empty but continue collection

**Performance target**
- Complete in ≤ 3 seconds
- Very lightweight DNS query

**Data quality notes**
- ASN data is highly reliable for commercial IP ranges
- Residential IPs may show ISP ASN instead of specific organization
- Some cloud providers may show generic ASN (AWS, Google, etc.)