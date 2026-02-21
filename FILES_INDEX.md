# NetProbe - Complete Files Index

## Project Location
`/sessions/compassionate-youthful-curie/repos/netprobe/`

## File Manifest

### Core Module Files

#### 1. go.mod
- **Lines:** 3
- **Purpose:** Go module definition
- **Content:** Module path (github.com/ErturkCan/netprobe), Go 1.21 requirement

#### 2. .gitignore  
- **Lines:** 36
- **Purpose:** Version control exclusions
- **Includes:** Build artifacts, IDE files, test output, vendor directories

### Command-Line Applications

#### 3. cmd/netprobe/main.go
- **Lines:** 330
- **Purpose:** Main CLI entry point with subcommands
- **Subcommands:**
  - `probe`: Send network probes (UDP or ICMP)
  - `analyze`: Detect bufferbloat under load
  - `listen`: Run UDP echo server
- **Features:**
  - Flag-based argument parsing
  - Comprehensive help text with examples
  - Error handling and validation
  - Both table and JSON output support

#### 4. cmd/listener/main.go
- **Lines:** 67
- **Purpose:** UDP echo server implementation
- **Features:**
  - Listens on configurable UDP port
  - Timestamps received packets (nanosecond precision)
  - Echoes packets back to sender
  - Displays packet information (sequence, size, RTT)
  - Continuous operation until interrupted

### Network Probing

#### 5. pkg/probe/udp.go
- **Lines:** 126
- **Purpose:** UDP-based network probing
- **Key Types:**
  - `UDPProbeConfig`: Configuration for UDP probes
  - `UDPProbeResult`: Individual probe result
  - `UDPProber`: Main prober implementation
- **Key Methods:**
  - `NewUDPProber()`: Create configured prober
  - `Probe()`: Execute series of probes
  - `sendProbe()`: Send single probe and measure RTT
- **Features:**
  - Timestamped packets (sequence + nanosecond timestamp)
  - Configurable: target, port, count, interval, payload size, timeout
  - Per-probe success/failure reporting
  - High-resolution RTT measurement

#### 6. pkg/probe/icmp.go
- **Lines:** 129
- **Purpose:** ICMP echo (ping) implementation
- **Key Types:**
  - `ICMPProbeConfig`: Configuration for ICMP probes
  - `ICMPProbeResult`: Individual probe result
  - `ICMPProber`: ICMP prober implementation
- **Key Methods:**
  - `NewICMPProber()`: Create configured prober
  - `Probe()`: Execute series of ping requests
  - `sendProbe()`: Send single ICMP echo and measure RTT
- **Features:**
  - Raw socket ICMP echo requests
  - Sequence number tracking for loss detection
  - Packet ID management
  - Timeout handling per probe

### Statistical Analysis

#### 7. pkg/stats/jitter.go
- **Lines:** 114
- **Purpose:** RFC 3550 interarrival jitter calculation
- **Key Types:**
  - `JitterCalculator`: Streaming jitter calculator
  - `JitterStats`: Jitter statistics result
- **Key Methods:**
  - `NewJitterCalculator()`: Create calculator
  - `AddSample()`: Process single RTT sample
  - `Jitter()`: Get jitter in microseconds
  - `JitterDuration()`: Get jitter as time.Duration
- **Algorithm:** J = J + (|D(i-1,i)| - J) / 16
- **Features:**
  - O(1) memory complexity
  - Streaming sample processing
  - Qualitative magnitude assessment
  - Fully RFC 3550 compliant

#### 8. pkg/stats/histogram.go
- **Lines:** 217
- **Purpose:** Latency histogram with percentile computation
- **Key Types:**
  - `LatencyHistogram`: Histogram implementation
  - `HistogramStats`: Complete statistics result
- **Key Methods:**
  - `NewLatencyHistogram()`: Create pre-allocated histogram
  - `AddSample()` / `AddSamples()`: Add RTT samples
  - `Min()`, `Max()`, `Mean()`, `StdDev()`: Basic stats
  - `Percentile()`, `P50()`, `P90()`, `P99()`, `P999()`: Percentiles
  - `GetStats()`: Get all statistics at once
- **Features:**
  - Pre-allocated bucket storage
  - Lazy sorting (on-demand computation)
  - Linear interpolation for percentiles
  - Arbitrary percentile calculation support
  - Standard deviation calculation

### Anomaly Detection

#### 9. pkg/detect/bufferbloat.go
- **Lines:** 134
- **Purpose:** Bufferbloat detection algorithm
- **Key Types:**
  - `BufferbloatDetector`: Main detector
  - `BufferbloatResult`: Detection results
- **Key Methods:**
  - `NewBufferbloatDetector()`: Create detector
  - `Detect()`: Perform bufferbloat analysis
- **Algorithm:**
  1. Measure baseline latency under idle conditions
  2. Generate background load
  3. Measure latency while load is applied
  4. Calculate latency increase ratios
  5. Classify severity based on thresholds
- **Severity Levels:**
  - Severe: p99 > 5.0x increase
  - Moderate: p99 > 3.0x increase
  - Mild: p99 > 1.5x OR p50 > 2.0x increase
  - None: Below all thresholds

### Output Formatting

#### 10. pkg/output/json.go
- **Lines:** 178
- **Purpose:** JSON output formatting
- **Key Types:**
  - `ProbeResultJSON`: Single probe result
  - `HistogramStatsJSON`: Statistics in JSON format
  - `JitterStatsJSON`: Jitter stats in JSON format
  - `ProbeReportJSON`: Complete report
  - `BufferbloatResultJSON`: Bufferbloat results
- **Key Functions:**
  - `WriteProbeResultsJSON()`: Format probe results as JSON
  - `WriteBufferbloatResultJSON()`: Format bufferbloat results as JSON
- **Features:**
  - Structured JSON with proper tags
  - Timestamp inclusion
  - Metrics in standard units (milliseconds)
  - Optional field handling (omitempty)
  - Pretty-printing with indentation

#### 11. pkg/output/table.go
- **Lines:** 188
- **Purpose:** Human-readable table output
- **Key Types:**
  - `TableWriter`: Main table formatter
- **Key Methods:**
  - `NewTableWriter()`: Create writer
  - `WriteProbeResults()`: Format individual probe results
  - `WriteStatistics()`: Format histogram statistics
  - `WriteJitterStats()`: Format jitter analysis
  - `WriteBufferbloatResults()`: Format bufferbloat detection
  - `WriteSeparator()`: Visual separation
- **Features:**
  - Aligned columns with consistent width
  - Section organization with headers
  - Dashed separators for clarity
  - Terminal-optimized formatting
  - Loss rate calculation

### Utilities

#### 12. internal/timing.go
- **Lines:** 57
- **Purpose:** High-resolution timing utilities
- **Key Types:**
  - `HighResTimer`: Timer with elapsed time tracking
- **Key Functions:**
  - `NewHighResTimer()`: Create timer
  - `NowNano()`: Current time in nanoseconds
  - `NowMicro()`: Current time in microseconds
  - `DurationMicros()`: Convert duration to microseconds
  - `DurationMillis()`: Convert duration to milliseconds
- **Key Methods:**
  - `Elapsed()`: Get elapsed time as duration
  - `ElapsedMillis()`: Get elapsed in milliseconds
  - `ElapsedMicros()`: Get elapsed in microseconds
  - `Reset()`: Reset timer

### Documentation

#### 13. README.md
- **Lines:** 395
- **Purpose:** User-facing documentation
- **Sections:**
  - Feature overview
  - Architecture diagram
  - Building instructions
  - Usage guide with examples
  - Sample output (table and JSON)
  - Component reference
  - Performance characteristics
  - Use case scenarios
  - Requirements and deployment
  - Future enhancement ideas

#### 14. IMPLEMENTATION_NOTES.md
- **Lines:** ~400 (approximate)
- **Purpose:** Technical deep-dive documentation
- **Sections:**
  - Technical highlights
  - Network protocol handling details
  - Statistical analysis algorithms
  - Anomaly detection methodology
  - Output formatting approach
  - Go idioms and best practices
  - Performance considerations
  - Code organization rationale
  - Error scenario handling
  - Testing strategy
  - Deployment notes
  - Future enhancement opportunities

#### 15. PROJECT_SUMMARY.txt
- **Lines:** ~100 (approximate)
- **Purpose:** Quick project overview
- **Contains:**
  - File listing with descriptions
  - Statistics and metrics
  - Features implemented
  - Architecture highlights
  - Verification checklist

#### 16. COMPLETION_REPORT.md
- **Lines:** ~300 (approximate)
- **Purpose:** Delivery verification
- **Contains:**
  - Deliverables checklist
  - Code quality metrics
  - Key implementation details
  - Project structure
  - Usage examples
  - Portfolio value demonstration
  - Verification checklist

#### 17. FILES_INDEX.md
- **Purpose:** This file
- **Contains:** Complete file-by-file reference

---

## Statistics Summary

| Metric | Count |
|--------|-------|
| Total Files | 16 |
| Go Source Files | 10 |
| Configuration Files | 1 |
| Documentation Files | 5 |
| Total Lines of Code | ~1,543 |
| Documentation Lines | ~1,200 |
| Total Project Size | ~156 KB |

---

## File Dependencies

```
cmd/netprobe/main.go
├── pkg/probe/udp.go
├── pkg/probe/icmp.go
├── pkg/stats/jitter.go
├── pkg/stats/histogram.go
└── pkg/output/
    ├── json.go
    └── table.go

cmd/listener/main.go
└── (no internal dependencies)

pkg/stats/jitter.go
└── (standard library only)

pkg/stats/histogram.go
├── (standard library only)
└── (uses time.Duration)

pkg/output/json.go
├── pkg/stats/histogram.go
└── pkg/stats/jitter.go

pkg/output/table.go
└── pkg/stats/histogram.go

internal/timing.go
└── (standard library only)
```

---

## Required Imports

### Standard Library Used
- encoding/binary (network byte order)
- encoding/json (JSON marshaling)
- flag (CLI arguments)
- fmt (formatting)
- io (I/O interfaces)
- log (logging)
- net (network operations)
- os (OS operations)
- sort (sorting)
- strings (string manipulation)
- sync (synchronization)
- time (timing and durations)

### External Imports
- golang.org/x/net/icmp (ICMP support)
- golang.org/x/net/ipv4 (IPv4 ICMP types)

---

## Building the Project

```bash
cd /sessions/compassionate-youthful-curie/repos/netprobe

# Initialize module (if needed)
go mod download

# Build main CLI
go build -o bin/netprobe ./cmd/netprobe

# Build listener
go build -o bin/netprobe-listener ./cmd/listener

# Run directly
go run ./cmd/netprobe probe -type udp -target localhost
```

---

## File Purposes at a Glance

| Purpose | Files |
|---------|-------|
| Network I/O | udp.go, icmp.go, listener/main.go |
| Statistics | jitter.go, histogram.go |
| Analysis | bufferbloat.go |
| Formatting | json.go, table.go |
| Utilities | timing.go |
| CLI | netprobe/main.go |
| Config | go.mod, .gitignore |
| Docs | README.md, IMPLEMENTATION_NOTES.md, etc. |

---

**Last Updated:** 2025-02-21  
**Project Status:** Complete and Ready for Use
