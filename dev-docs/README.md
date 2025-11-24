# netgaze Development Documentation

This directory contains comprehensive planning and implementation documentation for the netgaze network reconnaissance tool.

## Documentation Index

[01-report-struct-schema.md](01-report-struct-schema.md) - Core data structure definitions for all collected network information

[02-dns-collector-details.md](02-dns-collector-details.md) - DNS resolution implementation (A/AAAA/MX/NS/TXT/CNAME/PTR records)

[03-ping-collector-details.md](03-ping-collector-details.md) - ICMP ping implementation with packet loss and RTT statistics

[04-traceroute-collector-details.md](04-traceroute-collector-details.md) - Network path tracing with hop-by-hop analysis

[05-whois-collector-details.md](05-whois-collector-details.md) - WHOIS data collection and parsing for domains/IPs

[06-asn-bgp-collector-details.md](06-asn-bgp-collector-details.md) - ASN and BGP information via Team Cymru DNS lookup

[07-geolocation-collector-details.md](07-geolocation-collector-details.md) - IP geolocation and ISP data from ip-api.com

[08-port-scan-collector-details.md](08-port-scan-collector-details.md) - Port scanning implementation (opt-in, top 20 common ports)

[09-tls-collector-details.md](09-tls-collector-details.md) - TLS certificate collection from HTTPS endpoints

[10-collector-orchestrator.md](10-collector-orchestrator.md) - Parallel collection coordination with errgroup and error handling

[11-adk-agent-setup.md](11-adk-agent-setup.md) - AI agent integration with OpenRouter and custom tools

[12-template-contents.md](12-template-contents.md) - Output templates for text, markdown, and raw formats

[13-exact-common-ports-list.md](13-exact-common-ports-list.md) - Final hard-coded list of 20 most common ports

[14-error-handling-graceful-degradation.md](14-error-handling-graceful-degradation.md) - Error management and graceful degradation strategy

[15-cli-flag-definitions.md](15-cli-flag-definitions.md) - spf13/cobra CLI interface with config command

[16-tui-component-breakdown.md](16-tui-component-breakdown.md) - TUI architecture and state machine

[17-output-rendering-logic.md](17-output-rendering-logic.md) - Output format handling and piping detection

[18-build-distribution-notes.md](18-build-distribution-notes.md) - Build process and GitHub distribution strategy

[19-tui-implementation.md](19-tui-implementation.md) - Complete TUI implementation with Bubble Tea code samples

[20-tui-template-system.md](20-tui-template-system.md) - Comprehensive TUI styling system using Lip Gloss

[development-guidelines.md](development-guidelines.md) - Development requirements and no-special-characters policy

[PLAN.md](PLAN.md) - Master project overview and requirements

## About This Documentation

These documents are designed for developers implementing netgaze and serve as reference for advanced AI-assisted coding. The planning covers every aspect of the application from data collection to user interface, with detailed technical specifications and code examples.

### Key Features Covered

- **Network Intelligence**: Parallel data collection from multiple sources
- **AI Integration**: Optional OpenRouter-powered analysis with custom tools
- **Professional TUI**: Bubble Tea + Lip Gloss based terminal interface
- **Multiple Output Formats**: Text, markdown, JSON, and raw templates
- **Graceful Degradation**: Robust error handling with partial results
- **Cross-Platform**: Single static binary with broad OS support

### Development Approach

The documentation follows a modular approach where each component is independently specified but designed to work together seamlessly. All code examples use the latest Charmbracelet ecosystem patterns and follow the no-special-characters requirement for maximum compatibility.

For implementation guidance, agents should use the context7 MCP server to access the latest Charmbracelet documentation as needed.