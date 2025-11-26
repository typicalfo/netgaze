package collector

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/typicalfo/netgaze/internal/model"
	"golang.org/x/sync/errgroup"
)

func collectDNS(ctx context.Context, target string, report *model.Report) error {
	// Create context with 3-second timeout
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// First resolve to IP addresses (A and AAAA records)
	ips, err := resolveIPs(ctx, target)
	if err != nil {
		report.Errors["dns"] = fmt.Sprintf("IP resolution failed: %v", err)
		return fmt.Errorf("IP resolution failed: %w", err)
	}

	// Store IP addresses
	report.IPs = ips
	for _, ip := range ips {
		if ip.To4() != nil {
			report.IPv4 = append(report.IPv4, ip.String())
		} else {
			report.IPv6 = append(report.IPv6, ip.String())
		}
	}

	// If we have IPs, do reverse DNS (PTR) lookup
	if len(ips) > 0 {
		ptrs, err := resolvePTR(ctx, ips[0]) // Use first IP for PTR
		if err != nil {
			report.Errors["dns_ptr"] = fmt.Sprintf("PTR lookup failed: %v", err)
		} else {
			report.PTR = ptrs
		}
	}

	// Resolve other record types in parallel
	g, ctx := errgroup.WithContext(ctx)

	// CNAME records
	g.Go(func() error {
		cname, err := resolveCNAME(ctx, target)
		if err != nil {
			report.Errors["dns_cname"] = fmt.Sprintf("CNAME lookup failed: %v", err)
		} else if cname != "" {
			report.CNAME = []string{cname}
		}
		return nil
	})

	// MX records
	g.Go(func() error {
		mx, err := resolveMX(ctx, target)
		if err != nil {
			report.Errors["dns_mx"] = fmt.Sprintf("MX lookup failed: %v", err)
		} else {
			report.MX = mx
		}
		return nil
	})

	// NS records
	g.Go(func() error {
		ns, err := resolveNS(ctx, target)
		if err != nil {
			report.Errors["dns_ns"] = fmt.Sprintf("NS lookup failed: %v", err)
		} else {
			report.NS = ns
		}
		return nil
	})

	// TXT records
	g.Go(func() error {
		txt, err := resolveTXT(ctx, target)
		if err != nil {
			report.Errors["dns_txt"] = fmt.Sprintf("TXT lookup failed: %v", err)
		} else {
			report.TXT = txt
		}
		return nil
	})

	// Wait for all DNS lookups
	g.Wait()

	return nil
}

func resolveIPs(ctx context.Context, target string) ([]net.IP, error) {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 2 * time.Second,
			}
			return d.DialContext(ctx, network, address)
		},
	}

	ips, err := resolver.LookupIPAddr(ctx, target)
	if err != nil {
		return nil, err
	}

	var result []net.IP
	for _, ip := range ips {
		result = append(result, ip.IP)
	}

	return result, nil
}

func resolvePTR(ctx context.Context, ip net.IP) ([]string, error) {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 2 * time.Second,
			}
			return d.DialContext(ctx, network, address)
		},
	}

	names, err := resolver.LookupAddr(ctx, ip.String())
	if err != nil {
		return nil, err
	}

	return names, nil
}

func resolveCNAME(ctx context.Context, target string) (string, error) {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 2 * time.Second,
			}
			return d.DialContext(ctx, network, address)
		},
	}

	cname, err := resolver.LookupCNAME(ctx, target)
	if err != nil {
		return "", err
	}

	// If CNAME is the same as target, there's no CNAME
	if cname == target+"." {
		return "", nil
	}

	// Remove trailing dot
	return strings.TrimSuffix(cname, "."), nil
}

func resolveMX(ctx context.Context, target string) ([]string, error) {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 2 * time.Second,
			}
			return d.DialContext(ctx, network, address)
		},
	}

	mxRecords, err := resolver.LookupMX(ctx, target)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, mx := range mxRecords {
		result = append(result, fmt.Sprintf("%d %s", mx.Pref, mx.Host))
	}

	return result, nil
}

func resolveNS(ctx context.Context, target string) ([]string, error) {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 2 * time.Second,
			}
			return d.DialContext(ctx, network, address)
		},
	}

	nsRecords, err := resolver.LookupNS(ctx, target)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, ns := range nsRecords {
		result = append(result, ns.Host)
	}

	return result, nil
}

func resolveTXT(ctx context.Context, target string) ([]string, error) {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 2 * time.Second,
			}
			return d.DialContext(ctx, network, address)
		},
	}

	txtRecords, err := resolver.LookupTXT(ctx, target)
	if err != nil {
		return nil, err
	}

	return txtRecords, nil
}
