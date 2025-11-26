# netgaze

netgaze is a single-binary network reconnaissance tool with a Charmbracelet-powered TUI. Enter an IP, domain, or URL to run parallel passive/active collectors and view results in the terminal.

**NOTE:** netgaze is currently in active development and is not yet fully functional; expect incomplete features and potential breaking changes.

Mode:
- Deterministic, offline templated output (no AI)

## Features

- Parallel data collection: DNS, PTR, ping, traceroute, WHOIS, ASN/BGP, geolocation, port scan (opt-in), TLS certs
- Sub-12s average runtime
- Graceful degradation on failures
- Multiple outputs: TUI, text, markdown, JSON, raw
- Works behind proxies (HTTP_PROXY/NO_PROXY)
- Single static Go binary, no external deps

## Installation

```bash
go install github.com/typicalfo/netgaze@latest
```

Or build from source:
```bash
git clone https://github.com/typicalfo/netgaze
cd netgaze
make build  # Builds as 'ng' in project root
./ng example.com
```

## Quick Start

```bash
ng 1.1.1.1                    # Text output with styling
ng tui google.com --ports      # Interactive TUI mode
ng example.com
ng 8.8.8.8 --output json &gt; intel.json            # JSON for automation
```

Full CLI:
```
ng &lt;target&gt; [flags]           # Styled text output
ng tui &lt;target&gt; [flags]        # Interactive TUI mode
ng to &lt;target&gt; [flags]         # Traceroute JSON output
ng tc &lt;target&gt; [flags]         # Traceroute baseline compare
ng config [action]             # Manage configuration
ng version                     # Show version information

Flags:
  --ports             Scan common ports (opt-in)
  --output string     text/md/json/raw (for piping or automation)
  --no-style          Disable all ANSI styling
  --timeout duration  Global timeout (default 30s)
  --json              Legacy alias for --output json (hidden)
```

AI mode requires `OPENROUTER_API_KEY` env var.

## TUI

Launch with `ng tui <target>` for interactive mode:

- Header: target + timer
- Tabs (1-3): Summary, Raw Data (lipgloss tables), Ask
- Progress spinner during collection
- Navigable post-completion (q/Ctrl+C to quit)

## Data Collectors

| Feature | Package | Timeout |
|---------|---------|---------|
| DNS (A/AAAA/MX/NS/TXT/CNAME/PTR) | stdlib | 3s |
| Ping (ICMP, 5pkts) | pro-bing | 5s |
| Traceroute | go-traceroute | 10s |
| WHOIS | likexian/whois | 6s |
| ASN/BGP | ammario/ipisp | 3s |
| Geolocation | ip-api.com | 4s |
| Ports (top 20, opt-in) | naabu | 10s |
| TLS Cert (443) | crypto/tls | 4s |

Common ports: 22,53,80,110,135,139,143,443,993,995,1723,3306,3389,5900,8080,8443,10000

## No-Agent Output

Embedded templates (text/md/raw) styled with lipgloss at runtime.

## Development

Detailed specs in [dev-docs/](dev-docs/).

Code samples for Bubble Tea and Lip Gloss used in planning are not committed to this repo. See:
- [Bubble Tea examples](https://github.com/charmbracelet/bubbletea/tree/master/examples)
- [Lip Gloss examples](https://github.com/charmbracelet/lipgloss/tree/master/examples)

No special characters in code/templates/output per design.

## License

TBD