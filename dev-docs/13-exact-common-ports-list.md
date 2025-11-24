# Addendum 13 – Exact Common Ports List

**Final hard-coded port list for --ports flag**

**Primary 20 common ports**
```
22,53,80,110,135,139,143,443,993,995,1723,3306,3389,5900,8080,8443,10000,1433,445,992
```

**Port service mapping**
| Port | Service | Description |
|------|---------|-------------|
| 22   | SSH     | Secure Shell |
| 53   | DNS     | Domain Name System |
| 80   | HTTP    | Web traffic |
| 110  | POP3    | Post Office Protocol v3 |
| 135  | RPC     | Windows RPC |
| 139  | NetBIOS | Windows file sharing |
| 143  | IMAP    | Internet Message Access Protocol |
| 443  | HTTPS   | Secure web traffic |
| 993  | IMAPS   | Secure IMAP |
| 995  | POP3S   | Secure POP3 |
| 1723 | PPTP    | Point-to-Point Tunneling Protocol |
| 3306 | MySQL   | Database |
| 3389 | RDP     | Remote Desktop Protocol |
| 445  | SMB     | Windows file sharing |
| 5900 | VNC     | Virtual Network Computing |
| 8080 | HTTP-ALT| Alternative web port |
| 8443 | HTTPS-ALT| Alternative secure web |
| 992  | TelnetS | Secure Telnet |
| 10000| Webmin  | Web-based admin interface |
| 1433 | MSSQL   | Microsoft SQL Server |

**Selection criteria**
- Most common services found on internet-facing servers
- Includes standard management ports (SSH, RDP, VNC)
- Covers web services (HTTP/HTTPS variants)
- Database services (MySQL, MSSQL)
- Mail services (POP3, IMAP)
- Windows-specific services (RPC, SMB, NetBIOS)
- VPN services (PPTP)

**Implementation details**
```go
// internal/collector/ports.go
var commonPorts = []int{
    22, 53, 80, 110, 135, 139, 143, 443, 993, 995,
    1723, 3306, 3389, 5900, 8080, 8443, 10000, 1433, 445, 992,
}
```

**Scanning strategy**
- TCP SYN scan when privileged (root/administrator)
- TCP connect scan fallback for unprivileged users
- Single probe per port for speed
- 1-second timeout per port
- Total scan time: ~20 seconds maximum, typically <10 seconds

**Security considerations**
- Non-invasive: only checks port availability
- No protocol-specific payloads or banner grabbing
- Rate-limited to avoid triggering IDS/IPS
- Respects corporate network policies

**Future extensibility**
- Hard-coded list for MVP simplicity
- Could add --custom-ports flag later
- Port categories (web, db, management) for future filtering
- Service detection via banner grabbing (optional future feature)

**Rationale for exclusions**
- Excluded high-risk ports (no exploit attempts)
- Excluded uncommon services (reduces scan time)
- Focused on reconnaissance, not vulnerability assessment
- Avoids ports that might trigger aggressive security responses