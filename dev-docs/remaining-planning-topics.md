Here are the remaining addendums we still need to complete before the plan is 100% ready for coding agents:

3. Ping Collector Details  
4. Traceroute Collector Details  
5. WHOIS Collector Details  
6. ASN / BGP Collector Details  
7. Geolocation Collector Details  
8. Port Scan Collector Details  
9. TLS Certificate Collector Details  
10. Collector Orchestrator (collector/collector.go) – how errgroup, timeouts, and error aggregation work  
11. ADK Agent Setup & Tool Signatures (agent/agent.go + tools.go) – exact tool schemas, prompts, and OpenRouter config  
12. Template Contents (internal/templates/) – exact contents of summary.txt, summary.md, raw.txt  
13. Exact Common Ports List (final hard-coded list)  
14. Error Handling & Graceful Degradation Strategy (across all collectors + UI)  
15. CLI Flag Definitions (exact urfave/cli or cobra spec)  
16. TUI Component Breakdown & State Machine (ui/model.go – tabs, progress, ask mode)  
17. Output Rendering Logic (--output md/json/text/raw + piping detection)  
18. Build & Distribution Notes (go build flags, static binary, cgo considerations)
