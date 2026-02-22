# NetProbe - Network Latency Diagnostic Tool

A comprehensive Go-based network latency measurement and analysis tool designed for network engineers and system administrators. NetProbe provides detailed insights into network performance including latency distribution, jitter analysis, and bufferbloat detection.

## Features

- **UDP Probing**: Send timestamped UDP packets to measure round-trip time with high precision
- **ICMP Ping**: Traditional ICMP echo requests with sequence number tracking for packet loss detection
- **Latency Statistics**: Calculate percentiles (p50, p90, p99, p99.9), mean, standard deviation
- **Jitter Analysis**: RFC 3550 compliant interarrival jitter calculation
- **Bufferbloat Detection**: Measure latency degradation under load to identify buffer bloat
- **Flexible Output**: Human-readable tables or structured JSON for automation
- **High-Resolution Timing**: Microsecond-precision timing using Go's native timing API

## Architecture

```
netprobe/
├── cmd/
│   ├── netprobe/          # Main CLI application
│   └── listener/          # UDP echo server for probing
├── pkg/
│   ├── probe/             # Probe implementations (UDP, ICMP)
│   ├── stats/             # Statistical analysis (jitter, histogram)
│   ├── detect/            # Network anomaly detection (bufferbloat)
│   └── output/            # Output formatters (JSON, table)
├── internal/              # Internal utilities (timing)
└── go.mod               # Module definition
```

## Building

```bash
# Build main netprobe CLI
go build -o bin/netprobe ./cmd/netprobe

# Build UDP listener
go build -o bin/netprobe-listener ./cmd/listener
```

## Usage

### 1. UDP Probing

Send UDP echo probes to a target:

```bash
# Basic UDP probe (10 probes to localhost)
./bin/netprobe probe -type udp -target localhost

# Detailed configuration
./bin/netprobe probe -type udp \
  -target 8.8.8.8 \
  -port 12345 \
  -count 20 \
  -interval 500ms \
  -payload 64 \
  -timeout 3s \
  -output table

# JSON output for automation
./bin/netprobe probe -type udp -target 8.8.8.8 -output json
```

**Flags:**
- `-type`: Probe type (udp or icmp)
- `-target`: Target host or IP address (required)
- `-port`: UDP port (default: 12345)
- `-count`: Number of probes (default: 10)
- `-interval`: Time between probes (default: 1s)
- `-payload`: Payload size in bytes (default: 12)
- `-timeout`: Response timeout (default: 3s)
- `-output`: Output format: table or json (default: table)

### 2. ICMP Ping

Send ICMP echo requests (requires appropriate network permissions):

```bash
# Basic ping
./bin/netprobe probe -type icmp -target google.com

# Detailed configuration
./bin/netprobe probe -type icmp \
  -target 8.8.8.8 \
  -count 30 \
  -interval 100ms \
  -output json
```

Note: ICMP probing may require elevated privileges on some systems.

### 3. Run UDP Echo Server

Start a UDP echo server to receive probes:

```bash
# Run on default port 12345
./bin/netprobe-listener

# Run on custom port
./bin/netprobe-listener -port 5555
```

The listener will:
- Receive UDP packets from probes
- Timestamp them with high-resolution timing
- Echo them back to the sender
- Display packet information including sequence number and RTT

### 4. Bufferbloat Detection

Analyze network for bufferbloat by measuring latency under load:

```bash
# Basic analysis
./bin/netprobe analyze -target localhost

# Detailed configuration
./bin/netprobe analyze \
  -target 8.8.8.8 \
  -idle-count 15 \
  -load-count 15 \
  -output json
```

**Flags:**
- `-target`: Target host (required)
- `-idle-count`: Number of probes for idle measurement (default: 10)
- `-load-count`: Number of probes for loaded measurement (default: 10)
- `-output`: Output format: table or json (default: table)

## Sample Output

### UDP Probe Results (Table Format)

```
UDP Probe: target=localhost:12345, count=10, interval=1s, payload=12 bytes

=== UDP Probe Results ===
Target: localhost:12345
Probes sent: 10
Successful: 10
Failed: 0
Loss rate: 0.0%

Seq      RTT (ms)
-------- ----------
1        0.234
2        0.198
3        0.201
4        0.215
5        0.189
6        0.203
7        0.197
8        0.219
9        0.202
10       0.196

=== Statistics ===
Metric          Value
--------------- ---------------
Count           10
Min             0.189ms
Max             0.234ms
Mean            0.205ms
StdDev          0.014ms

=== Percentiles ===
Percentile      Latency
--------------- ---------------
p50             0.201ms
p90             0.221ms
p99             0.233ms
p99.9           0.234ms

=== Jitter Analysis ===
Metric               Value
-------------------- ---------------
Jitter (RFC3550)     0.015ms
Samples              9
Magnitude            Low
```

### UDP Probe Results (JSON Format)

```json
{
  "timestamp": 1704067200,
  "probe_type": "UDP",
  "target": "localhost:12345",
  "probe_results": [
    {
      "sequence": 1,
      "rtt_ms": 0.234,
      "success": true,
      "payload_len": 12
    },
    {
      "sequence": 2,
      "rtt_ms": 0.198,
      "success": true,
      "payload_len": 12
    }
  ],
  "statistics": {
    "count": 10,
    "min_ms": 0.189,
    "max_ms": 0.234,
    "mean_ms": 0.205,
    "stddev_ms": 0.014,
    "p50_ms": 0.201,
    "p90_ms": 0.221,
    "p99_ms": 0.233,
    "p999_ms": 0.234
  },
  "jitter": {
    "estimate_ms": 0.015,
    "count": 9,
    "magnitude": "Low"
  }
}
```

## Component Reference

### Probes

#### UDP Probe (`pkg/probe/udp.go`)
- Sends timestamped UDP packets with configurable payload
- Measures round-trip time by comparing send and receive timestamps
- Supports custom port, packet size, and count
- Provides detailed per-probe results including success/failure

**Packet Format:**
```
Bytes 0-3:    Sequence number (big-endian uint32)
Bytes 4-11:   Send timestamp in nanoseconds (big-endian uint64)
Bytes 12+:    Variable payload
```

#### ICMP Probe (`pkg/probe/icmp.go`)
- Uses raw ICMP sockets (requires root on Linux)
- Traditional ping-style echo requests
- Measures RTT with packet ID and sequence number tracking
- Useful for detecting packet loss at network layer

### Statistics

#### Jitter Calculator (`pkg/stats/jitter.go`)
- Implements RFC 3550 interarrival jitter algorithm
- Calculates running smoothed absolute value of delay variations
- Formula: J = J + (|D(i-1,i)| - J) / 16
- Returns jitter in microseconds with qualitative assessment

#### Latency Histogram (`pkg/stats/histogram.go`)
- Pre-allocated bucket storage for efficient memory usage
- Sorted-order computation only on demand
- Supports arbitrary percentile calculation via linear interpolation
- Computes mean, standard deviation, min, max
- Provides percentiles: p50, p90, p99, p99.9

### Anomaly Detection

#### Bufferbloat Detector (`pkg/detect/bufferbloat.go`)
- Measures latency under idle conditions (baseline)
- Generates background load while measuring latency
- Calculates latency increase ratios for p50, p99, and max
- Classifies severity: None, Mild, Moderate, Severe
- Thresholds:
 - Severe: p99 > 5.0x increase
 - Moderate: p99 > 3.0x increase
 - Mild: p99 > 1.5x OR p50 > 2.0x increase
 - None: Below thresholds

### Output Formatters

#### Table Output (`pkg/output/table.go`)
- Human-readable aligned columns
- Separate sections for results, statistics, percentiles, and jitter
- Formatted with separators and proper alignment
- Suitable for terminal display and log files

#### JSON Output (`pkg/output/json.go`)
- Structured JSON with timestamp
- Includes all probe results with timestamps
- Statistics embedded in response
- Easy integration with monitoring systems
- Can be piped to `jq` for further processing

### Utilities

#### High-Resolution Timing (`internal/timing.go`)
- Wrapper around Go's `time` package for nanosecond precision
- Functions: `NowNano()`, `NowMicro()`, `DurationMicros()`, `DurationMillis()`
- HighResTimer struct for measuring elapsed time
- Microsecond-level accuracy suitable for latency measurement

## Performance Characteristics

- **UDP Probes**: Sub-millisecond latency measurement accuracy
- **Memory Usage**: O(n) where n is number of samples; pre-allocated
- **CPU**: Minimal; timing operations are fast
- **Network**: Configurable probes per second (interval-based)

## Use Cases

1. **Network Baseline Profiling**: Establish normal latency characteristics
2. **Problem Diagnosis**: Identify high jitter, packet loss, or bufferbloat
3. **Monitoring Integration**: JSON output feeds into monitoring systems
4. **Performance Testing**: Validate network improvements
5. **ISP Quality Verification**: Test against well-known targets (8.8.8.8, 1.1.1.1)
6. **Application Tuning**: Understand network constraints for app design

## Requirements

- Go 1.21 or later
- Linux/macOS/Windows with network access
- ICMP probing requires elevated privileges (Linux: `sudo`, macOS: root, Windows: Administrator)
- UDP echo server on target for UDP probe testing

## Project Structure

```
netprobe/
├── cmd/
│   ├── netprobe/
│   │   └── main.go                 # CLI entry point, subcommand routing
│   └── listener/
│       └── main.go                 # UDP echo server implementation
├── pkg/
│   ├── probe/
│   │   ├── udp.go                  # UDP probing with RTT measurement
│   │   └── icmp.go                 # ICMP echo probing
│   ├── stats/
│   │   ├── jitter.go               # RFC 3550 jitter calculation
│   │   └── histogram.go            # Percentile-based latency analysis
│   ├── detect/
│   │   └── bufferbloat.go          # Bufferbloat detection algorithm
│   └── output/
│       ├── json.go                 # JSON formatting and marshaling
│       └── table.go                # Human-readable table output
├── internal/
│   └── timing.go                   # High-resolution timing utilities
├── go.mod                          # Go module definition
├── .gitignore                      # Git ignore patterns
└── README.md                       # This file
```

## Example Workflow

```bash
# Terminal 1: Start echo server
./bin/netprobe-listener -port 12345

# Terminal 2: Run baseline UDP probes
./bin/netprobe probe -type udp -target localhost -count 30 -interval 100ms

# Run bufferbloat analysis
./bin/netprobe analyze -target localhost -idle-count 20 -load-count 20

# Export results as JSON for further analysis
./bin/netprobe probe -type udp -target 8.8.8.8 -output json > results.json
cat results.json | jq '.statistics | .p99_ms'
```

## Error Handling

All components include comprehensive error handling:
- Network errors are wrapped with context
- Invalid parameters produce clear error messages
- Timeout handling for unresponsive targets
- Graceful degradation for partial failures

## Future Enhancements

- TCP probing for application-layer latency
- Multi-target concurrent probing
- Real-time graphing of latency trends
- Packet loss detection and analysis
- IPv6 support enhancements
- Custom probe protocols
- Rate limiting and backoff strategies

## License

This is a portfolio project demonstrating network systems expertise in Go.

## Author

Erturk Can

---

**Note**: For production use on public networks, ensure compliance with your network usage policies and obtain necessary permissions before running probing tools against remote targets.
