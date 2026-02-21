package stats

import (
	"time"
)

// JitterCalculator computes RFC 3550 interarrival jitter
// RFC 3550 defines jitter as the mean deviation (smoothed absolute value)
// of the difference between consecutive RTP timestamps
type JitterCalculator struct {
	lastTimestamp int64  // Last sequence number or timestamp
	interarrival  int64  // Last interarrival time
	jitter        int64  // Current jitter estimate
	count         int    // Number of measurements
	initialized   bool   // Whether first measurement has been taken
}

// NewJitterCalculator creates a new jitter calculator
func NewJitterCalculator() *JitterCalculator {
	return &JitterCalculator{
		initialized: false,
	}
}

// AddSample adds an RTT sample and updates jitter estimate
// Implements RFC 3550 algorithm: J = J + (|D(i-1,i)| - J) / 16
// where D is the interarrival delay
func (jc *JitterCalculator) AddSample(rtt time.Duration) {
	rttMicros := rtt.Microseconds()

	if !jc.initialized {
		jc.lastTimestamp = rttMicros
		jc.initialized = true
		return
	}

	// Calculate interarrival delay
	arrival := rttMicros
	d := arrival - jc.lastTimestamp
	if d < 0 {
		d = -d
	}

	// Update jitter estimate using exponential smoothing
	// J = J + (|D| - J) / 16
	diff := d - jc.jitter
	if diff < 0 {
		diff = -diff
	}
	jc.jitter = jc.jitter + diff/16

	jc.lastTimestamp = arrival
	jc.count++
}

// Jitter returns the current jitter estimate in microseconds
func (jc *JitterCalculator) Jitter() int64 {
	return jc.jitter
}

// JitterDuration returns the current jitter estimate as time.Duration
func (jc *JitterCalculator) JitterDuration() time.Duration {
	return time.Duration(jc.jitter) * time.Microsecond
}

// Count returns the number of samples processed
func (jc *JitterCalculator) Count() int {
	return jc.count
}

// Reset resets the jitter calculator
func (jc *JitterCalculator) Reset() {
	jc.lastTimestamp = 0
	jc.interarrival = 0
	jc.jitter = 0
	jc.count = 0
	jc.initialized = false
}

// JitterStats holds jitter statistics
type JitterStats struct {
	Estimate  time.Duration // Jitter estimate
	Count     int           // Number of samples
	Magnitude string        // Qualitative assessment: "Low", "Moderate", "High"
}

// CalculateJitterStats calculates jitter statistics from RTT samples
func CalculateJitterStats(rtts []time.Duration) JitterStats {
	jc := NewJitterCalculator()
	for _, rtt := range rtts {
		jc.AddSample(rtt)
	}

	jitterDur := jc.JitterDuration()
	magnitude := assessJitterMagnitude(jitterDur.Milliseconds())

	return JitterStats{
		Estimate:  jitterDur,
		Count:     jc.Count(),
		Magnitude: magnitude,
	}
}

// assessJitterMagnitude provides qualitative assessment of jitter level
func assessJitterMagnitude(jitterMs int64) string {
	switch {
	case jitterMs < 1:
		return "Low"
	case jitterMs < 10:
		return "Moderate"
	default:
		return "High"
	}
}
