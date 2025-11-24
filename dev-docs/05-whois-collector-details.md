# Addendum 5 – WHOIS Collector Details

**Responsibility**: Populate `Report.WhoisRaw` and parsed `Report.Whois` fields.

**File**: `internal/collector/whois.go`

**Behavior**
- Single function: `collectWhois(ctx context.Context, target string, r *model.Report)`
- Uses `github.com/likexian/whois` for cross-platform WHOIS queries
- 6 second total timeout
- Attempts both domain and IP WHOIS based on input type

**Query strategy**
```go
// 1. If target is hostname/domain → domain WHOIS first
// 2. Always try IP WHOIS on resolved IPs (first IPv4, then IPv6)
// 3. Store raw response, then parse top fields
```

**Parsing rules**
Extract these fields from raw WHOIS using regex patterns:
- `Domain`: from "Domain Name:" or similar
- `Registrar`: from "Registrar:" or "Sponsoring Registrar:"
- `Created`: from "Creation Date:", "Created:", "Registered:"
- `Expires`: from "Registry Expiry Date:", "Expiration Date:"
- `Registrant`: from "Registrant Name:", "Registrant Organization:"
- `NetRange`: from "NetRange:", "inetnum:"
- `NetName`: from "NetName:", "network:"
- `OrgName`: from "Organization:", "OrgName:"
- `Country`: from "Country:", "Jurisdiction:"
- `AbuseEmails`: extract all emails containing "abuse"

**Error handling**
- WHOIS server timeout → set `Error = "WHOIS query timeout"`
- No WHOIS data → set `Error = "no WHOIS data available"`
- Parse failures → store raw data, leave parsed fields empty
- Add to `r.Errors["whois"]` on complete failure

**Data quality**
- Normalize dates to ISO 8601 format when possible
- Clean whitespace and standardize field names
- Handle multiple abuse emails (comma-separated in raw, array in parsed)

**Performance target**
- Complete in ≤ 6 seconds
- Cache WHOIS server responses in memory during single run

**Special cases**
- IP ranges: show netblock info instead of domain
- Privacy protection: may show "REDACTED FOR PRIVACY"
- TLD-specific formats: handle different WHOIS formats gracefully