# netgaze

netgaze is a fast, single-binary network reconnaissance tool with a Charmbracelet-powered TUI. Enter an IP, domain, or URL to run parallel passive/active collectors and view results in a beautiful terminal interface.

Supports two modes:
- Full mode with AI analysis (OpenRouter/grok)
- `--no-agent` mode for deterministic, offline templated output

## Features

- Parallel data collection: DNS, PTR, ping, traceroute, WHOIS, ASN/BGP, geolocation, port scan (opt-in), TLS certs
- Sub-12s average runtime (&lt;8s no-agent)
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
go build -o netgaze ./cmd/root.go  # Update path as implemented
./netgaze example.com
```

## Quick Start

```bash
netgaze 1.1.1.1
netgaze google.com --ports
netgaze suspicious.site --no-agent --output md &gt; report.md
netgaze 8.8.8.8 --no-agent --output json &gt; intel.json
```

Full CLI:
```
netgaze &lt;target&gt; [flags]
  --ports          Scan common ports (opt-in)
  --no-agent, -A   Disable AI (faster, deterministic)
  --output string  text/md/json/raw (no-agent or piping)
  --timeout duration  Global timeout (default 15s)
  --json           Alias for --output json
```

AI mode requires `OPENROUTER_API_KEY` env var.

## TUI

- Header: target + timer
- Tabs (1-3): Summary (AI/templated), Raw Data (lipgloss tables), Ask (agent chat)
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