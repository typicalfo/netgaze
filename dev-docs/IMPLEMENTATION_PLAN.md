# netgaze Implementation Plan

## Project Overview
This plan implements netgaze, a fast network reconnaissance TUI tool built in Go with the Charmbracelet ecosystem. The tool runs multiple network data collectors in parallel and presents results through a beautiful terminal interface with optional AI augmentation.

## Implementation Phases

### Phase 1: Project Foundation and Core Structure
**Timeline: Days 1-2**

#### 1.1 Project Setup
- [ ] Initialize Go module with required dependencies
- [ ] Set up directory structure per PLAN.md specifications
- [ ] Configure build system and versioning
- [ ] Set up git hooks and CI/CD foundation

#### 1.2 Core Data Models
- [ ] Implement `internal/model/types.go` with the complete Report struct
- [ ] Add helper types (TraceHop, etc.)
- [ ] Create JSON marshaling utilities
- [ ] Add validation methods for target input

#### 1.3 CLI Foundation
- [ ] Implement `cmd/root.go` with cobra setup
- [ ] Add all CLI flags and validation
- [ ] Implement config command and file management
- [ ] Add version command with build info

### Phase 2: Network Data Collectors
**Timeline: Days 3-6**

#### 2.1 Collector Infrastructure
- [ ] Implement `internal/collector/collector.go` orchestrator
- [ ] Set up errgroup-based parallel execution
- [ ] Add timeout and context management
- [ ] Implement graceful degradation error handling

#### 2.2 DNS Collector (`internal/collector/dns.go`)
- [ ] Implement A/AAAA record resolution
- [ ] Add MX, NS, TXT, CNAME record collection
- [ ] Implement reverse DNS (PTR) lookup
- [ ] Add 3-second timeout with error handling

#### 2.3 Ping Collector (`internal/collector/ping.go`)
- [ ] Integrate prometheus-community/pro-bing
- [ ] Implement 5-packet ICMP echo with statistics
- [ ] Add RTT calculations (min/avg/max/stddev)
- [ ] Implement packet loss percentage

#### 2.4 Traceroute Collector (`internal/collector/traceroute.go`)
- [ ] Integrate pixelbender/go-traceroute
- [ ] Implement UDP-based tracing with fallback
- [ ] Add hop-by-hop analysis with RTT
- [ ] Handle timeout and unreachable hops

#### 2.5 WHOIS Collector (`internal/collector/whois.go`)
- [ ] Integrate likexian/whois
- [ ] Implement domain and IP WHOIS queries
- [ ] Parse key fields (registrar, dates, abuse contacts)
- [ ] Add 6-second timeout

#### 2.6 ASN/BGP Collector (`internal/collector/asn.go`)
- [ ] Integrate ammario/ipisp for Team Cymru lookup
- [ ] Implement ASN and organization detection
- [ ] Add BGP information collection
- [ ] Handle DNS-based lookup failures

#### 2.7 Geolocation Collector (`internal/collector/geo.go`)
- [ ] Implement ip-api.com integration
- [ ] Add JSON parsing for location data
- [ ] Include ISP, organization, and timezone info
- [ ] Handle API failures gracefully

#### 2.8 Port Scan Collector (`internal/collector/ports.go`)
- [ ] Integrate projectdiscovery/naabu
- [ ] Implement top 20 common ports scan
- [ ] Add port state classification (open/closed/filtered)
- [ ] Only run when --ports flag is specified

#### 2.9 TLS Certificate Collector (`internal/collector/tls.go`)
- [ ] Implement crypto/tls certificate grabbing
- [ ] Add certificate parsing (subject, issuer, SANs)
- [ ] Check expiration and self-signed status
- [ ] Only run when port 443 is open

### Phase 3: Output and Template System
**Timeline: Days 7-8**

#### 3.1 Template Infrastructure
- [ ] Implement `internal/templates/templates.go` with go:embed
- [ ] Create template loading and rendering system
- [ ] Add Lip Gloss styling integration
- [ ] Implement template error handling

#### 3.2 Output Templates
- [ ] Create `internal/templates/summary.txt` for terminal output
- [ ] Create `internal/templates/summary.md` for markdown
- [ ] Create `internal/templates/raw.txt` for minimal format
- [ ] Ensure all templates use identical Report struct

#### 3.3 Output Rendering Logic
- [ ] Implement output format detection and validation
- [ ] Add piping detection for non-TUI output
- [ ] Create JSON output with proper formatting
- [ ] Handle --output flag with --no-agent mode

### Phase 4: TUI Implementation
**Timeline: Days 9-11**

#### 4.1 TUI Foundation (`internal/ui/model.go`)
- [ ] Implement Bubble Tea model and state machine
- [ ] Add tab navigation (Summary, Raw Data, Ask)
- [ ] Implement progress tracking during collection
- [ ] Add keyboard navigation and quit handling

#### 4.2 TUI Components (`internal/ui/components/`)
- [ ] Create header component with target and timer
- [ ] Implement summary view with template rendering
- [ ] Add raw data view with formatted tables
- [ ] Create AI chat interface for Ask tab

#### 4.3 TUI Styling (`internal/ui/view.go`)
- [ ] Implement comprehensive Lip Gloss styling system
- [ ] Add color schemes and themes
- [ ] Create responsive layout for different terminal sizes
- [ ] Add progress indicators and status messages

#### 4.4 TUI Integration
- [ ] Connect TUI to collector orchestrator
- [ ] Implement real-time progress updates
- [ ] Add error display and graceful degradation UI
- [ ] Handle terminal resize events

### Phase 5: AI Agent Integration
**Timeline: Days 12-13**

#### 5.1 Agent Infrastructure (`internal/agent/agent.go`)
- [ ] Integrate google/agent-toolkit-go
- [ ] Set up OpenRouter API client with grok-4.1
- [ ] Implement agent initialization and configuration
- [ ] Add API key management and validation

#### 5.2 Custom Tools (`internal/agent/tools.go`)
- [ ] Implement `summarize_findings` tool
- [ ] Implement `detect_anomalies` tool
- [ ] Implement `suggest_next_steps` tool
- [ ] Implement `answer_question` tool

#### 5.3 AI Integration
- [ ] Connect agent to TUI Ask tab
- [ ] Implement streaming response display
- [ ] Add error handling for API failures
- [ ] Ensure graceful fallback when AI unavailable

### Phase 6: Integration and Testing
**Timeline: Days 14-15**

#### 6.1 Integration Testing
- [ ] Test all collectors with various targets
- [ ] Verify parallel execution and timeouts
- [ ] Test graceful degradation scenarios
- [ ] Validate output format consistency

#### 6.2 Performance Optimization
- [ ] Optimize collector timeouts and parallelism
- [ ] Ensure sub-12 second average runtime
- [ ] Optimize memory usage and allocations
- [ ] Test with slow/unreachable targets

#### 6.3 Error Handling Validation
- [ ] Test all failure modes and edge cases
- [ ] Verify graceful degradation behavior
- [ ] Test with network timeouts and failures
- [ ] Validate error message clarity

### Phase 7: Build and Distribution
**Timeline: Day 16**

#### 7.1 Build System
- [ ] Set up GitHub Actions for automated builds
- [ ] Configure cross-compilation for multiple platforms
- [ ] Add version injection and build metadata
- [ ] Create release automation

#### 7.2 Distribution
- [ ] Create GitHub releases with binaries
- [ ] Add Homebrew formula support
- [ ] Create installation documentation
- [ ] Set up update notification system

## Technical Requirements

### Dependencies
```go
// Core CLI and TUI
github.com/spf13/cobra
github.com/charmbracelet/bubbletea
github.com/charmbracelet/lipgloss

// Network collectors
github.com/prometheus-community/pro-bing
github.com/pixelbender/go-traceroute
github.com/likexian/whois
github.com/ammario/ipisp
github.com/projectdiscovery/naabu

// AI Integration
github.com/google/agent-toolkit-go

// Standard library
golang.org/x/sync/errgroup
```

### Performance Targets
- Total runtime: <12 seconds average
- No-agent mode: <8 seconds average
- Memory usage: <50MB typical
- Binary size: <25MB static

### Quality Standards
- Zero panics in production
- Graceful degradation for any collector failure
- Comprehensive error handling with user-friendly messages
- No special characters in output (copy/paste safe)

## Success Criteria

### Functional Requirements
- [ ] All collectors run in parallel with proper timeouts
- [ ] TUI displays real-time progress and results
- [ ] AI integration provides intelligent analysis
- [ ] Multiple output formats work correctly
- [ ] Graceful degradation handles all failure modes

### Non-Functional Requirements
- [ ] Single static binary with no external dependencies
- [ ] Cross-platform compatibility (Linux, macOS, Windows)
- [ ] Sub-12 second performance target achieved
- [ ] Professional TUI with responsive design
- [ ] Robust error handling with helpful messages

### Integration Requirements
- [ ] Works behind corporate proxies
- [ ] Respects HTTP_PROXY/NO_PROXY environment variables
- [ ] Works offline in --no-agent mode after data collection
- [ ] Handles all input types (IP, domain, URL) correctly

## Implementation Notes

### Development Approach
1. **Incremental Development**: Each phase builds upon previous work
2. **Parallel Testing**: Test collectors independently before integration
3. **Continuous Integration**: Automated testing at each phase
4. **Documentation**: Update documentation as implementation progresses

### Risk Mitigation
1. **Network Dependencies**: Handle all external service failures gracefully
2. **Performance**: Monitor and optimize throughout development
3. **Compatibility**: Test on multiple platforms early and often
4. **AI Integration**: Ensure graceful fallback when AI services unavailable

### Quality Assurance
1. **Unit Tests**: Test each collector and component independently
2. **Integration Tests**: Verify end-to-end functionality
3. **Performance Tests**: Validate timing requirements
4. **User Acceptance**: Ensure CLI and TUI meet usability standards

This implementation plan provides a structured approach to building netgaze while maintaining the high standards outlined in the documentation. Each phase has clear deliverables and success criteria, ensuring systematic progress toward a production-ready network reconnaissance tool.