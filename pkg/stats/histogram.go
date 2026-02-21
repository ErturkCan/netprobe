package stats

import (
	"fmt"
	"sort"
	"time"
)

// LatencyHistogram represents latency measurements with percentile computation
type LatencyHistogram struct {
	samples    []int64 // RTT samples in microseconds
	sorted     []int64 // Sorted samples (computed on demand)
	isDirty    bool    // Whether sorted needs recomputation
	bucketSize int64   // Bucket size in microseconds
}

// NewLatencyHistogram creates a new latency histogram with pre-allocated capacity
func NewLatencyHistogram(capacity int) *LatencyHistogram {
	return &LatencyHistogram{
		samples:    make([]int64, 0, capacity),
		bucketSize: 1000, // Default 1ms buckets
		isDirty:    false,
	}
}

// AddSample adds an RTT sample to the histogram
func (h *LatencyHistogram) AddSample(rtt time.Duration) {
	h.samples = append(h.samples, rtt.Microseconds())
	h.isDirty = true
}

// AddSamples adds multiple RTT samples
func (h *LatencyHistogram) AddSamples(rtts []time.Duration) {
	for _, rtt := range rtts {
		h.AddSample(rtt)
	}
}

// ensureSorted ensures the sorted array is up-to-date
func (h *LatencyHistogram) ensureSorted() {
	if !h.isDirty && len(h.sorted) > 0 {
		return
	}

	h.sorted = make([]int64, len(h.samples))
	copy(h.sorted, h.samples)
	sort.Slice(h.sorted, func(i, j int) bool {
		return h.sorted[i] < h.sorted[j]
	})
	h.isDirty = false
}

// Count returns the number of samples
func (h *LatencyHistogram) Count() int {
	return len(h.samples)
}

// Min returns the minimum latency
func (h *LatencyHistogram) Min() time.Duration {
	if len(h.samples) == 0 {
		return 0
	}
	h.ensureSorted()
	return time.Duration(h.sorted[0]) * time.Microsecond
}

// Max returns the maximum latency
func (h *LatencyHistogram) Max() time.Duration {
	if len(h.samples) == 0 {
		return 0
	}
	h.ensureSorted()
	return time.Duration(h.sorted[len(h.sorted)-1]) * time.Microsecond
}

// Mean returns the mean (average) latency
func (h *LatencyHistogram) Mean() time.Duration {
	if len(h.samples) == 0 {
		return 0
	}

	sum := int64(0)
	for _, sample := range h.samples {
		sum += sample
	}
	mean := sum / int64(len(h.samples))
	return time.Duration(mean) * time.Microsecond
}

// StdDev returns the standard deviation of latency
func (h *LatencyHistogram) StdDev() time.Duration {
	if len(h.samples) < 2 {
		return 0
	}

	mean := h.Mean().Microseconds()
	var sumSquares int64

	for _, sample := range h.samples {
		diff := sample - mean
		sumSquares += diff * diff
	}

	variance := sumSquares / int64(len(h.samples))
	stddev := int64(sqrt(float64(variance)))
	return time.Duration(stddev) * time.Microsecond
}

// Percentile returns the latency at the given percentile (0-100)
func (h *LatencyHistogram) Percentile(p float64) time.Duration {
	if len(h.samples) == 0 {
		return 0
	}
	if p < 0 || p > 100 {
		return 0
	}

	h.ensureSorted()

	// Linear interpolation between indices
	index := (p / 100.0) * float64(len(h.sorted)-1)
	lower := int(index)
	upper := lower + 1
	frac := index - float64(lower)

	if upper >= len(h.sorted) {
		return time.Duration(h.sorted[lower]) * time.Microsecond
	}

	interpolated := float64(h.sorted[lower])*(1-frac) + float64(h.sorted[upper])*frac
	return time.Duration(int64(interpolated)) * time.Microsecond
}

// P50 returns the 50th percentile (median)
func (h *LatencyHistogram) P50() time.Duration {
	return h.Percentile(50)
}

// P90 returns the 90th percentile
func (h *LatencyHistogram) P90() time.Duration {
	return h.Percentile(90)
}

// P99 returns the 99th percentile
func (h *LatencyHistogram) P99() time.Duration {
	return h.Percentile(99)
}

// P999 returns the 99.9th percentile
func (h *LatencyHistogram) P999() time.Duration {
	return h.Percentile(99.9)
}

// Percentiles returns multiple percentiles at once
func (h *LatencyHistogram) Percentiles(percentiles []float64) map[float64]time.Duration {
	result := make(map[float64]time.Duration, len(percentiles))
	for _, p := range percentiles {
		result[p] = h.Percentile(p)
	}
	return result
}

// Stats returns a summary of histogram statistics
type HistogramStats struct {
	Count int
	Min   time.Duration
	Max   time.Duration
	Mean  time.Duration
	StdDev time.Duration
	P50   time.Duration
	P90   time.Duration
	P99   time.Duration
	P999  time.Duration
}

// GetStats returns all statistics at once
func (h *LatencyHistogram) GetStats() HistogramStats {
	return HistogramStats{
		Count:  h.Count(),
		Min:    h.Min(),
		Max:    h.Max(),
		Mean:   h.Mean(),
		StdDev: h.StdDev(),
		P50:    h.P50(),
		P90:    h.P90(),
		P99:    h.P99(),
		P999:   h.P999(),
	}
}

// String returns a formatted string representation
func (s HistogramStats) String() string {
	return fmt.Sprintf(
		"Count: %d, Min: %.3fms, Max: %.3fms, Mean: %.3fms, StdDev: %.3fms, P50: %.3fms, P90: %.3fms, P99: %.3fms, P99.9: %.3fms",
		s.Count,
		s.Min.Seconds()*1000,
		s.Max.Seconds()*1000,
		s.Mean.Seconds()*1000,
		s.StdDev.Seconds()*1000,
		s.P50.Seconds()*1000,
		s.P90.Seconds()*1000,
		s.P99.Seconds()*1000,
		s.P999.Seconds()*1000,
	)
}

// sqrt is a simple integer square root
func sqrt(x float64) float64 {
	if x < 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}
