# Addendum 10 – Collector Orchestrator

**Responsibility**: Coordinate all collectors in parallel with proper error handling and timeouts.

**File**: `internal/collector/collector.go`

**Behavior**
- Single function: `Collect(ctx context.Context, target string, opts Options) (*model.Report, error)`
- Uses `golang.org/x/sync/errgroup` for parallel execution
- Global timeout of 15 seconds (configurable via --timeout flag)
- Graceful degradation: any collector failure doesn't stop others

**Exact orchestrator structure**
```go
type Options struct {
    EnablePorts bool
    NoAgent     bool
    Timeout     time.Duration
}

func Collect(ctx context.Context, target string, opts Options) (*model.Report, error) {
    ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
    defer cancel()
    
    report := &model.Report{
        Target:     target,
        ResolvedAt: time.Now().UTC(),
        ModeNoAgent: opts.NoAgent,
        Errors:     make(map[string]string),
    }
    
    g, ctx := errgroup.WithContext(ctx)
    
    // DNS first (needed by other collectors)
    g.Go(func() error { return collectDNS(ctx, target, report) })
    if err := g.Wait(); err != nil {
        return nil, fmt.Errorf("DNS resolution failed: %w", err)
    }
    
    // Reset errgroup for parallel collectors
    g, ctx = errgroup.WithContext(ctx)
    
    // Always run these collectors
    g.Go(func() error { return collectPing(ctx, target, report) })
    g.Go(func() error { return collectTraceroute(ctx, target, report) })
    g.Go(func() error { return collectWhois(ctx, target, report) })
    g.Go(func() error { return collectASN(ctx, target, report) })
    g.Go(func() error { return collectGeo(ctx, target, report) })
    
    // Port scan only when explicitly requested
    if opts.EnablePorts {
        g.Go(func() error { return collectPorts(ctx, target, report) })
    }
    
    // Wait for all collectors
    if err := g.Wait(); err != nil {
        // Log but don't fail - individual errors are in report.Errors
    }
    
    // TLS collection (depends on port scan results)
    if opts.EnablePorts && len(report.Ports.Open) > 0 && contains(report.Ports.Open, 443) {
        if err := collectTLS(ctx, target, report); err != nil {
            report.Errors["tls"] = err.Error()
        }
    }
    
    report.DurationMs = time.Since(report.ResolvedAt).Milliseconds()
    return report, nil
}
```

**Error handling strategy**
- Individual collector errors stored in `report.Errors[collectorName]`
- Only fail the entire collection if DNS resolution fails (required for other collectors)
- Context cancellation propagates to all collectors on timeout
- No panics: all collectors must handle their own errors gracefully

**Timeout hierarchy**
- Global timeout: 15s (configurable)
- Individual collector timeouts: as specified in their details
- Context cancellation: immediate on global timeout

**Performance optimization**
- DNS runs first and serially (dependency for other collectors)
- All other collectors run in parallel
- Memory allocation minimized (single Report instance)
- No unnecessary string copying or JSON marshaling during collection

**Logging and debugging**
- Structured logging with collector name and timing
- Debug mode shows individual collector start/end times
- Error aggregation for user-friendly display

**Graceful degradation examples**
- Ping fails → continue with other collectors
- WHOIS times out → show raw data from other sources
- Port scan fails → skip TLS collection
- Geolocation fails → show network data without location

**Return value**
- Always returns a Report (even partial)
- Only returns error for critical failures (DNS, context timeout)
- Report.Errors map contains per-collector error details