package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/typicalfo/netgaze/internal/model"
	"golang.org/x/sync/errgroup"
)

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

func contains(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Placeholder functions - will be implemented in separate files
