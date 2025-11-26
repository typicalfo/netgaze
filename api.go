package main

import (
	"context"
	"time"

	"github.com/typicalfo/netgaze/internal/collector"
	"github.com/typicalfo/netgaze/internal/model"
)

// Report is the main result type returned by collectors.
// It is an alias of the internal model.Report to keep
// external callers decoupled from internal packages.
type Report = model.Report

// Options controls how a netgaze run is executed.
// AI is not used in this version; all runs are
// deterministic, offline collector executions.
type Options struct {
	// EnablePorts enables the common-port scan collector.
	EnablePorts bool

	// Timeout is the overall timeout for all collectors.
	// If zero or negative, DefaultTimeout is used.
	Timeout time.Duration
}

// DefaultTimeout is the fallback timeout used when
// Options.Timeout is not set.
const DefaultTimeout = 15 * time.Second

// Run executes the netgaze collectors for the given target
// and returns a populated Report. This is the primary entry
// point for using netgaze as a Go package.
func Run(ctx context.Context, target string, opts Options) (*Report, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if opts.Timeout <= 0 {
		opts.Timeout = DefaultTimeout
	}

	report, err := collector.Collect(ctx, target, collector.Options{
		EnablePorts: opts.EnablePorts,
		NoAgent:     true,
		Timeout:     opts.Timeout,
	})
	if err != nil {
		return nil, err
	}

	return report, nil
}
