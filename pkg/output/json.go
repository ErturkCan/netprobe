package output

import (
	"encoding/json"
	"io"
	"time"

	"github.com/ErturkCan/netprobe/pkg/stats"
)

// ProbeResultJSON represents a single probe result in JSON format
type ProbeResultJSON struct {
	Sequence   int    `json:"sequence"`
	RTTMs      float64 `json:"rtt_ms"`
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	PayloadLen int    `json:"payload_len,omitempty"`
}

// HistogramStatsJSON represents histogram statistics in JSON format
type HistogramStatsJSON struct {
	Count      int     `json:"count"`
	MinMs      float64 `json:"min_ms"`
	MaxMs      float64 `json:"max_ms"`
	MeanMs     float64 `json:"mean_ms"`
	StdDevMs   float64 `json:"stddev_ms"`
	P50Ms      float64 `json:"p50_ms"`
	P90Ms      float64 `json:"p90_ms"`
	P99Ms      float64 `json:"p99_ms"`
	P999Ms     float64 `json:"p999_ms"`
}

// JitterStatsJSON represents jitter statistics in JSON format
type JitterStatsJSON struct {
	EstimateMs float64 `json:"estimate_ms"`
	Count      int     `json:"count"`
	Magnitude  string  `json:"magnitude"`
}

// ProbeReportJSON represents a complete probe report
type ProbeReportJSON struct {
	Timestamp    int64              `json:"timestamp"`
	ProbeType    string             `json:"probe_type"`
	Target       string             `json:"target"`
	ProbeResults []ProbeResultJSON  `json:"probe_results"`
	Statistics   HistogramStatsJSON `json:"statistics"`
	Jitter       JitterStatsJSON    `json:"jitter,omitempty"`
}

// WriteProbeResultsJSON writes probe results as JSON
func WriteProbeResultsJSON(w io.Writer, probeType, target string, results interface{}, histStats *stats.HistogramStats, jitterStats *stats.JitterStats) error {
	report := ProbeReportJSON{
		Timestamp: time.Now().Unix(),
		ProbeType: probeType,
		Target:    target,
	}

	// Convert results based on type
	switch v := results.(type) {
	case []interface{}:
		report.ProbeResults = make([]ProbeResultJSON, len(v))
		for i, r := range v {
			pj := ProbeResultJSON{}
			if m, ok := r.(map[string]interface{}); ok {
				if rtt, ok := m["rtt"].(time.Duration); ok {
					pj.RTTMs = rtt.Seconds() * 1000
				}
				if success, ok := m["success"].(bool); ok {
					pj.Success = success
				}
				if errMsg, ok := m["error"].(string); ok {
					pj.Error = errMsg
				}
			}
			report.ProbeResults[i] = pj
		}
	}

	// Add statistics if provided
	if histStats != nil {
		report.Statistics = HistogramStatsJSON{
			Count:    histStats.Count,
			MinMs:    histStats.Min.Seconds() * 1000,
			MaxMs:    histStats.Max.Seconds() * 1000,
			MeanMs:   histStats.Mean.Seconds() * 1000,
			StdDevMs: histStats.StdDev.Seconds() * 1000,
			P50Ms:    histStats.P50.Seconds() * 1000,
			P90Ms:    histStats.P90.Seconds() * 1000,
			P99Ms:    histStats.P99.Seconds() * 1000,
			P999Ms:   histStats.P999.Seconds() * 1000,
		}
	}

	// Add jitter stats if provided
	if jitterStats != nil {
		report.Jitter = JitterStatsJSON{
			EstimateMs: jitterStats.Estimate.Seconds() * 1000,
			Count:      jitterStats.Count,
			Magnitude:  jitterStats.Magnitude,
		}
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// BufferbloatResultJSON represents bufferbloat detection results
type BufferbloatResultJSON struct {
	Timestamp        int64   `json:"timestamp"`
	Target           string  `json:"target"`
	IdleP50Ms        float64 `json:"idle_p50_ms"`
	IdleP99Ms        float64 `json:"idle_p99_ms"`
	IdleMaxMs        float64 `json:"idle_max_ms"`
	LoadP50Ms        float64 `json:"load_p50_ms"`
	LoadP99Ms        float64 `json:"load_p99_ms"`
	LoadMaxMs        float64 `json:"load_max_ms"`
	P50Increase      float64 `json:"p50_increase_ratio"`
	P99Increase      float64 `json:"p99_increase_ratio"`
	MaxIncrease      float64 `json:"max_increase_ratio"`
	IsBufferbloated  bool    `json:"is_bufferbloated"`
	Severity         string  `json:"severity"`
	Explanation      string  `json:"explanation"`
}

// WriteBufferbloatResultJSON writes bufferbloat results as JSON
func WriteBufferbloatResultJSON(w io.Writer, target string, result interface{}) error {
	timestamp := time.Now().Unix()

	jsonResult := BufferbloatResultJSON{
		Timestamp: timestamp,
		Target:    target,
	}

	// Try to extract fields from result if it's a map
	if m, ok := result.(map[string]interface{}); ok {
		if v, ok := m["idle_p50"].(time.Duration); ok {
			jsonResult.IdleP50Ms = v.Seconds() * 1000
		}
		if v, ok := m["idle_p99"].(time.Duration); ok {
			jsonResult.IdleP99Ms = v.Seconds() * 1000
		}
		if v, ok := m["idle_max"].(time.Duration); ok {
			jsonResult.IdleMaxMs = v.Seconds() * 1000
		}
		if v, ok := m["load_p50"].(time.Duration); ok {
			jsonResult.LoadP50Ms = v.Seconds() * 1000
		}
		if v, ok := m["load_p99"].(time.Duration); ok {
			jsonResult.LoadP99Ms = v.Seconds() * 1000
		}
		if v, ok := m["load_max"].(time.Duration); ok {
			jsonResult.LoadMaxMs = v.Seconds() * 1000
		}
		if v, ok := m["p50_increase"].(float64); ok {
			jsonResult.P50Increase = v
		}
		if v, ok := m["p99_increase"].(float64); ok {
			jsonResult.P99Increase = v
		}
		if v, ok := m["max_increase"].(float64); ok {
			jsonResult.MaxIncrease = v
		}
		if v, ok := m["is_bufferbloated"].(bool); ok {
			jsonResult.IsBufferbloated = v
		}
		if v, ok := m["severity"].(string); ok {
			jsonResult.Severity = v
		}
		if v, ok := m["explanation"].(string); ok {
			jsonResult.Explanation = v
		}
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonResult)
}
