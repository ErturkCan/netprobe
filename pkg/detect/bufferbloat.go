package detect

import (
	"fmt"
	"sync"
	"time"

	"github.com/ErturkCan/netprobe/pkg/stats"
)

// BufferbloatDetector detects bufferbloat by measuring latency changes under load
type BufferbloatDetector struct {
	probeFn func(count int) ([]time.Duration, error) // Function to get RTT samples
}

// NewBufferbloatDetector creates a new bufferbloat detector
func NewBufferbloatDetector(probeFn func(count int) ([]time.Duration, error)) *BufferbloatDetector {
	return &BufferbloatDetector{
		probeFn: probeFn,
	}
}

// BufferbloatResult holds bufferbloat detection results
type BufferbloatResult struct {
	IdleLatencyP50    time.Duration
	IdleLatencyP99    time.Duration
	IdleLatencyMax    time.Duration
	LoadLatencyP50    time.Duration
	LoadLatencyP99    time.Duration
	LoadLatencyMax    time.Duration
	P50Increase       float64 // Ratio increase
	P99Increase       float64 // Ratio increase
	MaxIncrease       float64 // Ratio increase
	IsBufferbloated   bool    // True if significant increase detected
	Severity          string  // "None", "Mild", "Moderate", "Severe"
	Explanation       string  // Human-readable explanation
}

// Detect performs bufferbloat detection
// It measures latency under idle conditions, then creates load while measuring latency
func (bd *BufferbloatDetector) Detect(idleCount, loadCount int) (BufferbloatResult, error) {
	result := BufferbloatResult{}

	// Measure idle latency
	idleLatencies, err := bd.probeFn(idleCount)
	if err != nil {
		return result, fmt.Errorf("failed to measure idle latency: %w", err)
	}

	idleHist := stats.NewLatencyHistogram(len(idleLatencies))
	idleHist.AddSamples(idleLatencies)

	// Measure loaded latency
	// Start background probing to create load
	loadedLatencies := make([]time.Duration, 0, loadCount)
	var mu sync.Mutex

	// Function to continuously send probes in background (creates load)
	stopLoad := make(chan struct{})
	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopLoad:
				return
			case <-ticker.C:
				// Send probe in background (doesn't accumulate results)
				_ = bd.probeFn(1)
			}
		}
	}()

	// Measure latency while load is being generated
	time.Sleep(100 * time.Millisecond) // Let load start
	for i := 0; i < loadCount; i++ {
		latencies, err := bd.probeFn(1)
		if err == nil && len(latencies) > 0 {
			mu.Lock()
			loadedLatencies = append(loadedLatencies, latencies[0])
			mu.Unlock()
		}
		time.Sleep(50 * time.Millisecond)
	}

	close(stopLoad)
	time.Sleep(100 * time.Millisecond) // Let background load stop

	loadHist := stats.NewLatencyHistogram(len(loadedLatencies))
	loadHist.AddSamples(loadedLatencies)

	// Calculate results
	result.IdleLatencyP50 = idleHist.P50()
	result.IdleLatencyP99 = idleHist.P99()
	result.IdleLatencyMax = idleHist.Max()

	result.LoadLatencyP50 = loadHist.P50()
	result.LoadLatencyP99 = loadHist.P99()
	result.LoadLatencyMax = loadHist.Max()

	// Calculate increase ratios
	if result.IdleLatencyP50.Microseconds() > 0 {
		result.P50Increase = float64(result.LoadLatencyP50.Microseconds()) / float64(result.IdleLatencyP50.Microseconds())
	}
	if result.IdleLatencyP99.Microseconds() > 0 {
		result.P99Increase = float64(result.LoadLatencyP99.Microseconds()) / float64(result.IdleLatencyP99.Microseconds())
	}
	if result.IdleLatencyMax.Microseconds() > 0 {
		result.MaxIncrease = float64(result.LoadLatencyMax.Microseconds()) / float64(result.IdleLatencyMax.Microseconds())
	}

	// Assess bufferbloat severity
	result.IsBufferbloated = result.P50Increase > 1.5 || result.P99Increase > 2.0
	result.Severity, result.Explanation = assessBufferbloat(result)

	return result, nil
}

// assessBufferbloat evaluates bufferbloat severity
func assessBufferbloat(result BufferbloatResult) (string, string) {
	p50Increase := result.P50Increase
	p99Increase := result.P99Increase

	switch {
	case p99Increase > 5.0:
		return "Severe", "Latency increased dramatically under load. Buffer bloat is severe - consider optimizing buffer sizes or traffic shaping."
	case p99Increase > 3.0:
		return "Moderate", "Significant latency increase under load. Moderate buffer bloat detected - network may benefit from QoS improvements."
	case p99Increase > 1.5 || p50Increase > 2.0:
		return "Mild", "Noticeable latency increase under load. Mild buffer bloat - monitor for potential issues."
	default:
		return "None", "Latency remained stable under load. No significant buffer bloat detected."
	}
}
