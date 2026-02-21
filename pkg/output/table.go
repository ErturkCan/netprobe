package output

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ErturkCan/netprobe/pkg/stats"
)

// TableWriter provides human-readable table output
type TableWriter struct {
	w io.Writer
}

// NewTableWriter creates a new table writer
func NewTableWriter(w io.Writer) *TableWriter {
	return &TableWriter{w: w}
}

// WriteProbeResults writes probe results in table format
func (tw *TableWriter) WriteProbeResults(probeType, target string, rtts []time.Duration, failures int) error {
	fmt.Fprintf(tw.w, "=== %s Probe Results ===\n", strings.ToUpper(probeType))
	fmt.Fprintf(tw.w, "Target: %s\n", target)
	fmt.Fprintf(tw.w, "Probes sent: %d\n", len(rtts)+failures)
	fmt.Fprintf(tw.w, "Successful: %d\n", len(rtts))
	fmt.Fprintf(tw.w, "Failed: %d\n", failures)

	if len(rtts) > 0 {
		fmt.Fprintf(tw.w, "Loss rate: %.1f%%\n", float64(failures)/float64(len(rtts)+failures)*100)
	}

	fmt.Fprintln(tw.w)

	// Write table header
	fmt.Fprintf(tw.w, "%-8s %-12s\n", "Seq", "RTT (ms)")
	fmt.Fprintf(tw.w, "%-8s %-12s\n", strings.Repeat("-", 8), strings.Repeat("-", 12))

	// Write individual results
	for i, rtt := range rtts {
		fmt.Fprintf(tw.w, "%-8d %-12.3f\n", i+1, rtt.Seconds()*1000)
	}

	fmt.Fprintln(tw.w)

	return nil
}

// WriteStatistics writes statistics in table format
func (tw *TableWriter) WriteStatistics(stats stats.HistogramStats) error {
	fmt.Fprintln(tw.w, "=== Statistics ===")

	// Write statistics table
	fmt.Fprintf(tw.w, "%-15s %-15s\n", "Metric", "Value")
	fmt.Fprintf(tw.w, "%-15s %-15s\n", strings.Repeat("-", 15), strings.Repeat("-", 15))

	fmt.Fprintf(tw.w, "%-15s %-15d\n", "Count", stats.Count)
	fmt.Fprintf(tw.w, "%-15s %-15.3fms\n", "Min", stats.Min.Seconds()*1000)
	fmt.Fprintf(tw.w, "%-15s %-15.3fms\n", "Max", stats.Max.Seconds()*1000)
	fmt.Fprintf(tw.w, "%-15s %-15.3fms\n", "Mean", stats.Mean.Seconds()*1000)
	fmt.Fprintf(tw.w, "%-15s %-15.3fms\n", "StdDev", stats.StdDev.Seconds()*1000)

	fmt.Fprintln(tw.w)
	fmt.Fprintln(tw.w, "=== Percentiles ===")

	fmt.Fprintf(tw.w, "%-15s %-15s\n", "Percentile", "Latency")
	fmt.Fprintf(tw.w, "%-15s %-15s\n", strings.Repeat("-", 15), strings.Repeat("-", 15))

	fmt.Fprintf(tw.w, "%-15s %-15.3fms\n", "p50", stats.P50.Seconds()*1000)
	fmt.Fprintf(tw.w, "%-15s %-15.3fms\n", "p90", stats.P90.Seconds()*1000)
	fmt.Fprintf(tw.w, "%-15s %-15.3fms\n", "p99", stats.P99.Seconds()*1000)
	fmt.Fprintf(tw.w, "%-15s %-15.3fms\n", "p99.9", stats.P999.Seconds()*1000)

	fmt.Fprintln(tw.w)

	return nil
}

// WriteJitterStats writes jitter statistics in table format
func (tw *TableWriter) WriteJitterStats(js stats.JitterStats) error {
	fmt.Fprintln(tw.w, "=== Jitter Analysis ===")

	fmt.Fprintf(tw.w, "%-20s %-15s\n", "Metric", "Value")
	fmt.Fprintf(tw.w, "%-20s %-15s\n", strings.Repeat("-", 20), strings.Repeat("-", 15))

	fmt.Fprintf(tw.w, "%-20s %-15.3fms\n", "Jitter (RFC3550)", js.Estimate.Seconds()*1000)
	fmt.Fprintf(tw.w, "%-20s %-15d\n", "Samples", js.Count)
	fmt.Fprintf(tw.w, "%-20s %-15s\n", "Magnitude", js.Magnitude)

	fmt.Fprintln(tw.w)

	return nil
}

// WriteBufferbloatResults writes bufferbloat detection results in table format
func (tw *TableWriter) WriteBufferbloatResults(target string, result interface{}) error {
	fmt.Fprintln(tw.w, "=== Bufferbloat Detection Results ===")
	fmt.Fprintf(tw.w, "Target: %s\n", target)

	// Extract values from result map
	var (
		idleP50, idleP99, idleMax     float64
		loadP50, loadP99, loadMax     float64
		p50Inc, p99Inc, maxInc        float64
		isBloated                     bool
		severity, explanation         string
	)

	if m, ok := result.(map[string]interface{}); ok {
		if v, ok := m["idle_p50"].(time.Duration); ok {
			idleP50 = v.Seconds() * 1000
		}
		if v, ok := m["idle_p99"].(time.Duration); ok {
			idleP99 = v.Seconds() * 1000
		}
		if v, ok := m["idle_max"].(time.Duration); ok {
			idleMax = v.Seconds() * 1000
		}
		if v, ok := m["load_p50"].(time.Duration); ok {
			loadP50 = v.Seconds() * 1000
		}
		if v, ok := m["load_p99"].(time.Duration); ok {
			loadP99 = v.Seconds() * 1000
		}
		if v, ok := m["load_max"].(time.Duration); ok {
			loadMax = v.Seconds() * 1000
		}
		if v, ok := m["p50_increase"].(float64); ok {
			p50Inc = v
		}
		if v, ok := m["p99_increase"].(float64); ok {
			p99Inc = v
		}
		if v, ok := m["max_increase"].(float64); ok {
			maxInc = v
		}
		if v, ok := m["is_bufferbloated"].(bool); ok {
			isBloated = v
		}
		if v, ok := m["severity"].(string); ok {
			severity = v
		}
		if v, ok := m["explanation"].(string); ok {
			explanation = v
		}
	}

	fmt.Fprintln(tw.w)
	fmt.Fprintln(tw.w, "=== Idle Conditions ===")
	fmt.Fprintf(tw.w, "%-15s %-15s\n", "Metric", "Latency")
	fmt.Fprintf(tw.w, "%-15s %-15s\n", strings.Repeat("-", 15), strings.Repeat("-", 15))
	fmt.Fprintf(tw.w, "%-15s %-15.3fms\n", "p50", idleP50)
	fmt.Fprintf(tw.w, "%-15s %-15.3fms\n", "p99", idleP99)
	fmt.Fprintf(tw.w, "%-15s %-15.3fms\n", "Max", idleMax)

	fmt.Fprintln(tw.w)
	fmt.Fprintln(tw.w, "=== Under Load ===")
	fmt.Fprintf(tw.w, "%-15s %-15s\n", "Metric", "Latency")
	fmt.Fprintf(tw.w, "%-15s %-15s\n", strings.Repeat("-", 15), strings.Repeat("-", 15))
	fmt.Fprintf(tw.w, "%-15s %-15.3fms\n", "p50", loadP50)
	fmt.Fprintf(tw.w, "%-15s %-15.3fms\n", "p99", loadP99)
	fmt.Fprintf(tw.w, "%-15s %-15.3fms\n", "Max", loadMax)

	fmt.Fprintln(tw.w)
	fmt.Fprintln(tw.w, "=== Latency Increase Ratios ===")
	fmt.Fprintf(tw.w, "%-15s %-15s\n", "Metric", "Increase")
	fmt.Fprintf(tw.w, "%-15s %-15s\n", strings.Repeat("-", 15), strings.Repeat("-", 15))
	fmt.Fprintf(tw.w, "%-15s %-15.2fx\n", "p50", p50Inc)
	fmt.Fprintf(tw.w, "%-15s %-15.2fx\n", "p99", p99Inc)
	fmt.Fprintf(tw.w, "%-15s %-15.2fx\n", "Max", maxInc)

	fmt.Fprintln(tw.w)
	fmt.Fprintln(tw.w, "=== Assessment ===")
	fmt.Fprintf(tw.w, "Bufferbloated: %v\n", isBloated)
	fmt.Fprintf(tw.w, "Severity: %s\n", severity)
	fmt.Fprintf(tw.w, "Explanation: %s\n", explanation)

	fmt.Fprintln(tw.w)

	return nil
}

// WriteSeparator writes a visual separator
func (tw *TableWriter) WriteSeparator() error {
	fmt.Fprintln(tw.w, strings.Repeat("=", 60))
	return nil
}
