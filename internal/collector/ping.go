package collector

import (
	"context"
	"fmt"
	"math"
	"time"

	probing "github.com/prometheus-community/pro-bing"
	"github.com/typicalfo/netgaze/internal/model"
)

func collectPing(ctx context.Context, target string, report *model.Report) error {
	// Create context with 5-second timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Create pinger
	pinger, err := probing.NewPinger(target)
	if err != nil {
		report.Errors["ping"] = fmt.Sprintf("Failed to create pinger: %v", err)
		return fmt.Errorf("failed to create pinger: %w", err)
	}

	// Configure pinger
	pinger.Count = 5
	pinger.Interval = 200 * time.Millisecond
	pinger.Timeout = 4 * time.Second
	pinger.SetPrivileged(false) // Don't require privileged mode

	// Statistics collection
	var rtts []time.Duration
	var packetsSent, packetsReceived int

	// Set up handlers
	pinger.OnRecv = func(pkt *probing.Packet) {
		packetsReceived++
		rtts = append(rtts, pkt.Rtt)
	}

	pinger.OnFinish = func(stats *probing.Statistics) {
		packetsSent = int(stats.PacketsSent)
		packetsReceived = int(stats.PacketsRecv)
		if len(stats.Rtts) > 0 {
			rtts = stats.Rtts
		}
	}

	// Run ping directly (pro-bing has its own timeout handling)
	err = pinger.Run()
	if err != nil {
		report.Errors["ping"] = fmt.Sprintf("Ping failed: %v", err)
		return fmt.Errorf("ping failed: %w", err)
	}

	// Calculate statistics
	report.Ping.PacketsSent = packetsSent
	report.Ping.PacketsReceived = packetsReceived

	if packetsSent > 0 {
		report.Ping.PacketLossPct = float64(packetsSent-packetsReceived) / float64(packetsSent) * 100
	}

	if len(rtts) > 0 {
		// Calculate RTT statistics
		var sum, sumSquares float64
		minRtt, maxRtt := rtts[0], rtts[0]

		for _, rtt := range rtts {
			ms := float64(rtt.Nanoseconds()) / 1e6
			sum += ms
			sumSquares += ms * ms
			if rtt < minRtt {
				minRtt = rtt
			}
			if rtt > maxRtt {
				maxRtt = rtt
			}
		}

		avgRtt := sum / float64(len(rtts))
		variance := (sumSquares / float64(len(rtts))) - (avgRtt * avgRtt)
		stdDev := math.Sqrt(variance)

		report.Ping.MinRtt = formatDuration(minRtt)
		report.Ping.AvgRtt = formatDuration(time.Duration(avgRtt * 1e6))
		report.Ping.MaxRtt = formatDuration(maxRtt)
		report.Ping.StdDevRtt = fmt.Sprintf("%.2fms", stdDev)
	}

	report.Ping.Success = packetsReceived > 0

	return nil
}

func formatDuration(d time.Duration) string {
	ms := float64(d.Nanoseconds()) / 1e6
	if ms < 1 {
		return fmt.Sprintf("%.2fms", ms)
	}
	return fmt.Sprintf("%.1fms", ms)
}
