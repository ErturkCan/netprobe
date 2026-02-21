package internal

import (
	"time"
)

// HighResTimer provides high-resolution timing utilities
type HighResTimer struct {
	start time.Time
}

// NewHighResTimer creates a new high-resolution timer
func NewHighResTimer() *HighResTimer {
	return &HighResTimer{
		start: time.Now(),
	}
}

// Elapsed returns the duration since timer creation in nanoseconds
func (t *HighResTimer) Elapsed() time.Duration {
	return time.Since(t.start)
}

// ElapsedMillis returns elapsed time in milliseconds
func (t *HighResTimer) ElapsedMillis() float64 {
	return float64(t.Elapsed().Microseconds()) / 1000.0
}

// ElapsedMicros returns elapsed time in microseconds
func (t *HighResTimer) ElapsedMicros() int64 {
	return t.Elapsed().Microseconds()
}

// Reset resets the timer
func (t *HighResTimer) Reset() {
	t.start = time.Now()
}

// NowNano returns current time as nanoseconds since epoch
func NowNano() int64 {
	return time.Now().UnixNano()
}

// NowMicro returns current time as microseconds since epoch
func NowMicro() int64 {
	return time.Now().UnixMicro()
}

// DurationMicros converts duration to microseconds
func DurationMicros(d time.Duration) int64 {
	return d.Microseconds()
}

// DurationMillis converts duration to milliseconds
func DurationMillis(d time.Duration) float64 {
	return float64(d.Microseconds()) / 1000.0
}
