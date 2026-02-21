# NetProbe - Network Latency Diagnostic Tool

A comprehensive Go-based network latency measurement and analysis tool designed for network engineers and system administrators. NetProbe provides detailed insights into network performance including latency distribution, jitter analysis, and bufferbloat detection.

## Features

- **UDP Probing**: Send timestamped UDP packets to measure round-trip time with high precision
- **ICMP Ping**: Traditional ICMP echo requests with sequence number tracking for packet loss detection
- **Latency Statistics**: Calculate percentiles (p50, p90, p99, p99.9), mean, standard deviation
- **Jitter Analysis**: RFC 3550 compliant interarrival jitter calculation
- **Bufferbloat Detection**: Measure latency degradation under load to identify buffer bloat
- **Flexible Output**: Human-readable tables or structured JSON for automation

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
└── go.mod
```

## Usage

```bash
# Build
go build -o bin/netprobe ./cmd/netprobe
go build -o bin/netprobe-listener ./cmd/listener

# UDP probe
./bin/netprobe probe -type udp -target localhost -count 100

# ICMP ping
sudo ./bin/netprobe probe -type icmp -target 8.8.8.8 -count 50

# Bufferbloat detection
./bin/netprobe detect -target localhost -duration 10s

# JSON output
./bin/netprobe probe -type udp -target localhost -format json
```

## Jitter Analysis

RFC 3550 compliant interarrival jitter calculation with exponentially weighted moving average. Jitter score categorization: Excellent (<1ms), Good (1-5ms), Fair (5-20ms), Poor (>20ms).

## Bufferbloat Detection

Measures latency degradation under synthetic load by comparing baseline latency against loaded latency. Reports bloat ratio with severity classification.

## Testing

```bash
go test ./...
go test -race ./...
go test -bench=. -benchmem ./...
```

## License

MIT License - See LICENSE file
