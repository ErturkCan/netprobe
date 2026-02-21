# NetProbe Implementation Notes

## Project Overview

NetProbe is a production-quality network latency diagnostic tool written in Go, demonstrating:
- Advanced network programming with UDP and ICMP
- Statistical analysis and percentile computation
- Clean architecture and error handling
- Performance-conscious design patterns
- Professional Go development practices

## Technical Highlights

### 1. Network Protocol Handling

#### UDP Probing (`pkg/probe/udp.go`)
- **Packet Structure**: Combines sequence number, timestamp, and variable payload
  ```
  [Seq (4 bytes)] [Timestamp (8 bytes)] [Payload (variable)]
  ```
- **Bidirectional Communication**: Sends probe, waits for echo response
- **Timeout Handling**: Configurable per-probe timeout with deadline management
- **Error Resilience**: Continues operation even if some probes fail

#### ICMP Probing (`pkg/probe/icmp.go`)
- **Raw Socket Usage**: Uses `golang.org/x/net/icmp` for ICMP support
- **Sequence Tracking**: Sequence numbers for detecting packet loss
- **Platform Awareness**: Handles IPv4 ICMP (IPv6 support extensible)

### 2. Statistical Analysis

#### Jitter Calculation (`pkg/stats/jitter.go`)
- **RFC 3550 Compliance**: Implements exact algorithm specified in RFC 3550
  ```
  J = J + (|D(i-1,i)| - J) / 16
  ```
- **Memory Efficient**: O(1) space complexity (only tracks state, not history)
- **Streaming**: Processes samples incrementally without buffering
- **Magnitude Assessment**: Categorizes jitter as Low/Moderate/High

#### Latency Histogram (`pkg/stats/histogram.go`)
- **Lazy Sorting**: Defers sorting until percentile calculation needed
- **On-Demand Computation**: `isDirty` flag tracks if recalculation needed
- **Linear Interpolation**: Smooth percentile calculation between data points
- **Comprehensive Statistics**:
  - Count, Min, Max, Mean, StdDev
  - Percentiles: p50, p90, p99, p99.9
  - Arbitrary percentile computation

### 3. Anomaly Detection

#### Bufferbloat Detection (`pkg/detect/bufferbloat.go`)
- **Two-Phase Measurement**:
  1. Baseline latency under idle conditions
  2. Latency measurement while introducing background load
- **Quantitative Analysis**:
  - Calculates increase ratios for p50, p99, and max latency
  - Compares idle vs. loaded conditions
- **Severity Classification**:
  ```
  Severe:   p99 > 5.0x increase
  Moderate: p99 > 3.0x increase  
  Mild:     p99 > 1.5x OR p50 > 2.0x increase
  None:     Below thresholds
  ```

### 4. Output Formatting

#### JSON Output (`pkg/output/json.go`)
- **Structured Serialization**: Uses encoding/json for type-safe marshaling
- **Automation Friendly**: Timestamps, metrics in standard units (milliseconds)
- **Extensible Format**: Easy to add new fields without breaking consumers
- **Optional Fields**: Uses `omitempty` for conditional inclusion

#### Table Output (`pkg/output/table.go`)
- **Aligned Columns**: Consistent width formatting for readability
- **Section Organization**: Groups related metrics together
- **Separator Lines**: Visual clarity with dashed separators
- **Terminal Optimized**: Works well with piping and logging

### 5. Go Idioms and Best Practices

#### Error Handling
```go
// Error wrapping with context
if err != nil {
    return nil, fmt.Errorf("failed to resolve address: %w", err)
}

// Graceful degradation
for i := 0; i < count; i++ {
    result := probe(conn, i+1)
    if result.Error != nil {
        failures++
    }
    results = append(results, result)
}
```

#### Resource Management
```go
// Deferred cleanup
conn, err := net.DialUDP("udp", nil, addr)
if err != nil {
    return nil, err
}
defer conn.Close()  // Always cleaned up
```

#### Type-Driven Design
```go
// Configuration objects
type UDPProbeConfig struct {
    Target      string
    Port        int
    Count       int
    Interval    time.Duration
    PayloadSize int
    Timeout     time.Duration
}

// Builder pattern
func NewUDPProber(config UDPProbeConfig) *UDPProber {
    // Apply defaults
    if config.Count == 0 { config.Count = 10 }
    // Validation
    if config.Timeout == 0 { config.Timeout = 3 * time.Second }
    return &UDPProber{config: config}
}
```

#### Receiver Methods
```go
// Data operations as receiver methods
func (h *LatencyHistogram) AddSample(rtt time.Duration) {
    h.samples = append(h.samples, rtt.Microseconds())
    h.isDirty = true
}

func (h *LatencyHistogram) Percentile(p float64) time.Duration {
    h.ensureSorted()  // Lazy computation
    // ... interpolation logic
}
```

### 6. Performance Considerations

#### Memory Efficiency
- **Pre-allocated Slices**: Histogram allocates with initial capacity
  ```go
  samples: make([]int64, 0, capacity)  // Pre-allocates
  ```
- **Lazy Sorting**: Avoids O(n log n) if percentiles not requested
- **O(1) Jitter**: No history buffer, just running state
- **Binary Format**: Compact network packet representation

#### CPU Optimization
- **Minimal Allocations**: Buffer reuse in network operations
- **Direct Computations**: Standard deviation calculated in single pass
- **Early Exit**: Conditions checked before expensive operations

#### Timing Precision
- **Microsecond Resolution**: Uses `time.UnixMicro()` for high precision
- **Direct OS Calls**: Minimal overhead between measurement points
- **Consistent Clock**: Single `time.Now()` call per direction

### 7. CLI Design

#### Subcommand Architecture
```go
switch subcommand {
case "probe":    probeCommand(args)
case "analyze":  analyzeCommand(args)
case "listen":   listenCommand(args)
}
```

#### Flag Management
- **Separate Flag Sets**: Each subcommand has own FlagSet
- **Default Values**: Sensible defaults for all options
- **Clear Help Text**: Examples provided in usage output
- **Type Safety**: Using flag package's typed parameters

### 8. Code Organization

```
netprobe/
├── cmd/                    # Executable entry points
│   ├── netprobe/          # Main CLI tool
│   └── listener/          # Supporting server
├── pkg/                    # Public packages
│   ├── probe/             # Network probe implementations
│   ├── stats/             # Statistical analysis
│   ├── detect/            # Anomaly detection
│   └── output/            # Result formatting
├── internal/              # Internal utilities (not part of public API)
│   └── timing.go
├── go.mod                 # Module definition
├── README.md              # User documentation
└── .gitignore
```

**Rationale**:
- `pkg/` exports for external use (if published)
- `internal/` for implementation details
- `cmd/` for executable binaries
- Clear separation of concerns

### 9. Error Scenarios Handled

1. **Network Errors**
   - Target unreachable
   - DNS resolution failures
   - Timeout on response
   - Connection refused

2. **Configuration Errors**
   - Missing required flags
   - Invalid port numbers
   - Invalid payload sizes
   - Invalid probe types

3. **Operational Errors**
   - Insufficient privileges (ICMP)
   - File descriptor exhaustion
   - Malformed responses
   - System resource limits

4. **Data Errors**
   - Empty sample sets
   - Invalid percentile ranges
   - Negative durations
   - Jitter calculation edge cases

### 10. Extensibility

The architecture supports:
- **New Probe Types**: Add to `pkg/probe/` with `Probe()` interface
- **New Statistics**: Add methods to histogram in `pkg/stats/`
- **New Detectors**: Add to `pkg/detect/` following bufferbloat pattern
- **New Output Formats**: Add formatters to `pkg/output/`
- **New CLI Commands**: Add to subcommand switch in `cmd/netprobe/main.go`

## Testing Strategy

### Manual Testing
```bash
# Local UDP probing
go run ./cmd/listener -port 12345 &
go run ./cmd/netprobe probe -type udp -target localhost

# Remote ICMP
go run ./cmd/netprobe probe -type icmp -target 8.8.8.8

# JSON output
go run ./cmd/netprobe probe -type udp -target localhost -output json | jq

# Load testing
go run ./cmd/netprobe analyze -target localhost
```

### Verification Points
- Latency values are positive durations
- Percentiles maintain p50 ≤ p90 ≤ p99 ≤ p99.9
- Jitter increases monotonically with variance
- Bufferbloat severity correlates with latency increase
- JSON output is valid and parseable
- Table output is human-readable

## Deployment Notes

### Build
```bash
go build -o bin/netprobe ./cmd/netprobe
go build -o bin/netprobe-listener ./cmd/listener
```

### Runtime Requirements
- Linux/macOS/Windows with network access
- ICMP: Root/admin privileges on most systems
- UDP: No elevated privileges needed

### Configuration
- All parameters via command-line flags
- No configuration files required
- Reasonable defaults for all options

## Future Enhancement Opportunities

1. **TCP Probing**: Application-layer latency measurement
2. **IPv6 Support**: Full IPv6 ICMP and UDP probing
3. **Multi-target**: Concurrent probing to multiple targets
4. **Historical Tracking**: Time-series data collection
5. **Real-time Graphing**: Live latency visualization
6. **Custom Protocols**: User-defined probe formats
7. **Rate Limiting**: Token bucket for bandwidth control
8. **Database Backend**: Persistent result storage

---

This implementation demonstrates production-quality Go development with emphasis on:
- Correctness and reliability
- Performance and efficiency
- Clean code and maintainability
- Comprehensive documentation
- Extensible architecture
