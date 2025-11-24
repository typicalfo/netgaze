# netgaze – Final MVP Planning Document (v1.0 – Locked for Implementation)

## Project Overview
`netgaze` is a fast, single-binary, Charmbracelet-powered TUI that takes an IP or hostname/URL and instantly runs all common passive/active reconnaissance tools in parallel, then presents the results beautifully.

Two distinct modes:
1. Full mode – with AI augmentation (Google ADK + OpenRouter/grok-4.1)
2. `--no-agent` mode – zero AI, fully deterministic, templated output, works offline after data collection

## Core Goals
- Sub-12 second total runtime for average target
- Zero external binary dependencies
- Single static Go binary
- Graceful degradation on any failed collector
- Works behind corporate proxies (respects HTTP_PROXY/NO_PROXY)

## CLI Interface
```
netgaze <ip|domain|url> [flags]

Flags:
  --ports          Enable light port scan (common ports)
  --no-agent, -A   Disable AI entirely (faster, deterministic)
  --output string  text (default), md, json, raw   (only active with --no-agent or when piping)
  --timeout duration   Global timeout (default 15s)
  --json           Legacy alias for --output json
```

Examples:
```bash
netgaze 1.1.1.1
netgaze google.com --ports
netgaze suspicious.site --no-agent --output md > report.md
netgaze 8.8.8.8 --no-agent --output json > intel.json
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
| Common port scan  | Top ~20 common ports when --ports               | projectdiscovery/naabu (config: -top-ports 100 or custom list) | 10s |
| TLS cert grab     | If 443 open, pull cert subject/CN/expiry        | crypto/tls                                | 4s      |

Common ports list (hard-coded):
`22,53,80,110,135,139,143,443,993,995,1723,3306,3389,5900,8080,8443,10000`

## AI Integration (only when NOT --no-agent)
- Provider: OpenRouter (model: grok-4.1 free tier)
- Toolkit: google/agent-toolkit-go
- Env var: OPENROUTER_API_KEY (required only in AI mode)
- Single agent with four custom tools:
  - summarize_findings(json_report)
  - detect_anomalies(json_report)
  - suggest_next_steps(json_report)
  - answer_question(user_question, json_report)
- All raw data passed as structured JSON to the agent
- Streaming output in TUI

## TUI Layout (bubbletea)
- Header: target + elapsed time
- Three tabs (1-3 keys):
  1. Summary
     - With AI → live-streamed agent response
     - With --no-agent → static Go text/template
  2. Raw Data → beautifully formatted sections (lipgloss tables)
  3. Ask → live chat with agent (hidden entirely in --no-agent mode)
- Progress spinner during collection
- Fully navigable after completion (q or Ctrl+C to quit)

## No-AI Templated Output (--no-agent)
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
│   ├── agent/
│   │   ├── agent.go             # ADK init + OpenRouter config
│   │   └── tools.go             # custom tool definitions
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
- Full run <12s average, <8s in --no-agent mode
- Works without OPENROUTER_API_KEY when --no-agent used
- All output formats (--output md/json/text/raw) identical in data
- Zero panics on unreachable hosts or timeouts
- Single static binary (no cgo if possible for broader OS support)
