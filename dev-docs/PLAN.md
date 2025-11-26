# netgaze – Final MVP Planning Document (v1.0 – Locked for Implementation)

## Project Overview
`netgaze` is a fast, single-binary, Charmbracelet-powered TUI that takes an IP or hostname/URL and instantly runs all common passive/active reconnaissance tools in parallel, then presents the results beautifully.

Two distinct modes:
1. Full mode – with AI augmentation (Google ADK + OpenRouter/grok-4.1)
2. Offline templated output mode – deterministic, works without external services

## Core Goals
- Sub-12 second total runtime for average target
- Zero external binary dependencies
- Single static Go binary
- Graceful degradation on any failed collector
- Works behind corporate proxies (respects HTTP_PROXY/NO_PROXY)

## CLI Interface
```
ng <ip|domain|url> [flags]           # Default text output
ng tui <ip|domain|url> [flags]        # Interactive TUI mode

Flags:
  --ports          Enable port scan of common ports (not enabled by default)
  --ai, -A         Enable AI-augmented mode (requires OPENROUTER_API_KEY)
  --output string  text (default), md, json, raw   (only active in non-AI mode or when piping)
  --timeout duration   Global timeout (default 15s)
  --json           Legacy alias for --output json
```

Examples:
```bash
ng 1.1.1.1                           # Text output with styling
ng tui google.com --ports             # Interactive TUI mode
ng -ai suspicious.site --output md > report.md
ng 8.8.8.8 --output json > intel.json
```

## Data Collection (all run in parallel)
| Feature           | Implementation                                  | Package                                   | Timeout |
|-------------------|-------------------------------------------------|-------------------------------------------|---------|
| DNS (A/AAAA/MX/NS/TXT/CNAME) | net.Resolver + context                    | stdlib / miekg/dns if needed              | 3s      |
| Reverse DNS (PTR) | net.LookupAddr                                  | stdlib                                    | 3s      |
| Ping              | ICMP echo, 5 packets                            | github.com/prometheus-community/pro-bing  | 5s      |
| Traceroute        | UDP-based with fallback                                 | github.com/pixelbender/go-traceroute      | 10s     |
| WHOIS             | Domain/IP whois query                           | github.com/likexian/whois                 | 6s      |
| ASN + BGP         | Team Cymru DNS lookup                           | github.com/ammario/ipisp                  | 3s      |
| Geolocation       | ip-api.com (free tier, no key)                  | net/http + JSON                           | 4s      |
| Common port scan  | Top 20 common ports only when --ports explicitly specified | projectdiscovery/naabu (config: -top-ports 100 or custom list) | 10s |
| TLS cert grab     | If 443 open, pull cert subject/CN/expiry        | crypto/tls                                | 4s      |

Common ports list (hard-coded):
`22,53,80,110,135,139,143,443,993,995,1723,3306,3389,5900,8080,8443,10000`

## Future Integration (reserved)
- Provider: OpenRouter (model: grok-4.1 free tier)
- Toolkit: (not used in this version)
- Env var: OPENROUTER_API_KEY (required only in AI mode)
- Planned integration with external toolkit (removed in this version):
  - summarize_findings(json_report)
  - detect_anomalies(json_report)
  - suggest_next_steps(json_report)
  - answer_question(user_question, json_report)
- All raw data passed as structured JSON to any future integration layer
- Streaming output in TUI

## TUI Layout (bubbletea)
- Header: target + elapsed time
- Three tabs (1-3 keys):
  1. Summary
      - Static Go text/template summary
   2. Raw Data → beautifully formatted sections (lipgloss tables)
   3. Ask → reserved for future interactive help (disabled in this version)
- Progress spinner during collection
- Fully navigable after completion (q or Ctrl+C to quit)

## Non-AI Templated Output
Templates are embedded at compile time (//go:embed):
- internal/templates/summary.txt     → colored terminal (lipgloss-styled at runtime)
- internal/templates/summary.md      → GitHub-flavored markdown
- internal/templates/raw.txt        → minimal one-liner format

All templates receive the exact same Report struct used for JSON and AI.

## Directory Structure
```
netgaze/
├── main.go
├── cmd/
│   └── root.go                  # cobra/viper or urfave/cli setup
├── internal/
│   ├── collector/
│   │   ├── collector.go         # orchestrates everything via errgroup
│   │   ├── dns.go
│   │   ├── ping.go
│   │   ├── traceroute.go
│   │   ├── whois.go
│   │   ├── asn.go
│   │   ├── geo.go
│   │   └── ports.go
│   ├── model/
│   │   └── types.go             # full Report struct + sub-structs
│   ├── integration/            # reserved for any future integrations
│   │   └── README.md
│   ├── ui/
│   │   ├── model.go
│   │   ├── view.go
│   │   └── components/
│   └── templates/
│       ├── templates.go         # //go:embed declarations
│       ├── summary.txt
│       ├── summary.md
│       └── raw.txt
├── go.mod
└── README.md
```

## Success Criteria
- Full run <12s average
- Works without any AI or external LLM keys
- All output formats (--output md/json/text/raw) identical in data
- Zero panics on unreachable hosts or timeouts
- Single static binary (no cgo if possible for broader OS support)

## Development Requirements
- **No special characters**: Do not use emoji, unicode symbols, or special characters (✅, ❌, ⚠️, etc.) in any code, templates, documentation, or output. Users will copy/paste output frequently, and these characters complicate downstream processing. Use plain text alternatives (e.g., "Success", "Error", "Warning" instead of symbols).
