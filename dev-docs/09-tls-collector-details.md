# Addendum 9 – TLS Certificate Collector Details

**Responsibility**: Populate `Report.TLS` with certificate information from HTTPS endpoints.

**File**: `internal/collector/tls.go`

**Behavior**
- Only runs if port 443 is detected as open (from port scan or manual check)
- Single function: `collectTLS(ctx context.Context, target string, r *model.Report)`
- Uses Go stdlib `crypto/tls` for certificate retrieval
- 4 second timeout
- Opportunistic: failure doesn't stop other collectors

**Exact configuration**
```go
dialer := &net.Dialer{Timeout: 4 * time.Second}
conn, err := tls.DialWithDialer(dialer, "tcp", target+":443", &tls.Config{
    InsecureSkipVerify: true, // we want cert even if invalid
    ServerName:         extractHostname(target),
})
```

**Certificate parsing**
```go
certs := conn.ConnectionState().PeerCertificates
if len(certs) > 0 {
    cert := certs[0] // leaf certificate
    
    r.TLS.Subject = cert.Subject.String()
    r.TLS.Issuer = cert.Issuer.String()
    r.TLS.CommonName = cert.Subject.CommonName
    r.TLS.AltNames = cert.DNSNames
    r.TLS.NotBefore = cert.NotBefore.Format(time.RFC3339)
    r.TLS.NotAfter = cert.NotAfter.Format(time.RFC3339)
    r.TLS.Expired = time.Now().After(cert.NotAfter)
    r.TLS.SelfSigned = cert.Issuer.CommonName == cert.Subject.CommonName
}
```

**Hostname extraction**
- For domains: use the domain directly
- For IPs: use the IP as ServerName (may cause cert validation errors, but we still get cert)
- For URLs: extract hostname from URL

**Error handling**
- Connection refused → set `Error = "TLS connection refused"`
- Timeout → set `Error = "TLS handshake timeout"`
- No certificates → set `Error = "no TLS certificates presented"`
- Add to `r.Errors["tls"]` on complete failure

**Security considerations**
- `InsecureSkipVerify: true` to get cert even for invalid chains
- Don't validate certificate chain for reconnaissance purposes
- Close connection immediately after cert retrieval

**Performance target**
- Complete in ≤ 4 seconds
- Fast failure if port not actually open
- Single TCP connection

**Data quality**
- Extract all Subject Alternative Names
- Properly format dates in ISO 8601
- Handle self-signed certificates gracefully
- Detect expired certificates

**Special cases**
- Wildcard certificates (*.example.com)
- Multi-domain certificates (SANs)
- Certificate transparency information (if available)
- OCSP stapling status (if available)