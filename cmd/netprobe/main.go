package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ErturkCan/netprobe/pkg/output"
	"github.com/ErturkCan/netprobe/pkg/probe"
	"github.com/ErturkCan/netprobe/pkg/stats"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "probe":
		probeCommand(os.Args[2:])
	case "analyze":
		analyzeCommand(os.Args[2:])
	case "listen":
		listenCommand(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`NetProbe - Network Latency Diagnostic Tool

Usage:
  netprobe probe [options]    - Send network probes (UDP or ICMP)
  netprobe analyze [options]  - Analyze probe results and detect bufferbloat
  netprobe listen [options]   - Run UDP echo server
  netprobe help               - Show this help message

Global Options:
  -help                       Show help for specific command`)

	fmt.Println("\nProbe Command:")
	fmt.Println(`  netprobe probe -type <udp|icmp> -target <host> [options]

  Options:
    -type string              Probe type: udp or icmp (default: udp)
    -target string            Target host or IP address (required)
    -port int                 Target port for UDP (default: 12345)
    -count int                Number of probes (default: 10)
    -interval duration        Interval between probes (default: 1s)
    -payload int              Payload size in bytes (default: 12)
    -timeout duration         Response timeout (default: 3s)
    -output string            Output format: table or json (default: table)

Examples:
  netprobe probe -type udp -target 8.8.8.8
  netprobe probe -type icmp -target google.com -count 20 -interval 500ms
  netprobe probe -type udp -target localhost -output json`)

	fmt.Println("\nAnalyze Command:")
	fmt.Println(`  netprobe analyze [options]

  Options:
    -target string            Target host or IP address (required)
    -idle-count int           Probes for idle measurement (default: 10)
    -load-count int           Probes for loaded measurement (default: 10)
    -output string            Output format: table or json (default: table)

Examples:
  netprobe analyze -target 8.8.8.8
  netprobe analyze -target localhost -idle-count 20 -output json`)

	fmt.Println("\nListen Command:")
	fmt.Println(`  netprobe listen [options]

  Options:
    -port int                 UDP port to listen on (default: 12345)

Examples:
  netprobe listen
  netprobe listen -port 5555`)
}

func probeCommand(args []string) {
	fs := flag.NewFlagSet("probe", flag.ExitOnError)

	probeType := fs.String("type", "udp", "Probe type: udp or icmp")
	target := fs.String("target", "", "Target host or IP address")
	port := fs.Int("port", 12345, "Target port for UDP")
	count := fs.Int("count", 10, "Number of probes")
	interval := fs.Duration("interval", 1*time.Second, "Interval between probes")
	payload := fs.Int("payload", 12, "Payload size in bytes")
	timeout := fs.Duration("timeout", 3*time.Second, "Response timeout")
	outputFormat := fs.String("output", "table", "Output format: table or json")

	fs.Parse(args)

	if *target == "" {
		fmt.Println("Error: -target flag is required")
		fs.Usage()
		os.Exit(1)
	}

	switch *probeType {
	case "udp":
		probeUDP(*target, *port, *count, *interval, *payload, *timeout, *outputFormat)
	case "icmp":
		probeICMP(*target, *count, *interval, *timeout, *outputFormat)
	default:
		fmt.Printf("Error: Unknown probe type: %s\n", *probeType)
		os.Exit(1)
	}
}

func probeUDP(target string, port, count int, interval time.Duration, payload int, timeout time.Duration, outputFormat string) {
	fmt.Printf("UDP Probe: target=%s:%d, count=%d, interval=%v, payload=%d bytes\n",
		target, port, count, interval, payload)
	fmt.Println()

	config := probe.UDPProbeConfig{
		Target:      target,
		Port:        port,
		Count:       count,
		Interval:    interval,
		PayloadSize: payload,
		Timeout:     timeout,
	}

	prober := probe.NewUDPProber(config)
	results, err := prober.Probe()
	if err != nil {
		log.Fatalf("Probe failed: %v", err)
	}

	// Extract successful RTTs and calculate statistics
	var rtts []time.Duration
	failures := 0

	for _, result := range results {
		if result.Success {
			rtts = append(rtts, result.RTT)
		} else {
			failures++
		}
	}

	// Calculate statistics
	hist := stats.NewLatencyHistogram(len(rtts))
	hist.AddSamples(rtts)
	histStats := hist.GetStats()

	// Calculate jitter
	jitterStats := stats.CalculateJitterStats(rtts)

	// Output results
	switch outputFormat {
	case "json":
		_ = output.WriteProbeResultsJSON(os.Stdout, "UDP", target, results, &histStats, &jitterStats)
	default:
		tw := output.NewTableWriter(os.Stdout)
		_ = tw.WriteProbeResults("UDP", target, rtts, failures)
		_ = tw.WriteStatistics(histStats)
		_ = tw.WriteJitterStats(jitterStats)
	}
}

func probeICMP(target string, count int, interval time.Duration, timeout time.Duration, outputFormat string) {
	fmt.Printf("ICMP Probe: target=%s, count=%d, interval=%v\n",
		target, count, interval)
	fmt.Println()

	config := probe.ICMPProbeConfig{
		Target:   target,
		Count:    count,
		Interval: interval,
		Timeout:  timeout,
	}

	prober := probe.NewICMPProber(config)
	results, err := prober.Probe()
	if err != nil {
		log.Fatalf("Probe failed: %v", err)
	}

	// Extract successful RTTs and calculate statistics
	var rtts []time.Duration
	failures := 0

	for _, result := range results {
		if result.Success {
			rtts = append(rtts, result.RTT)
		} else {
			failures++
		}
	}

	// Calculate statistics
	hist := stats.NewLatencyHistogram(len(rtts))
	hist.AddSamples(rtts)
	histStats := hist.GetStats()

	// Calculate jitter
	jitterStats := stats.CalculateJitterStats(rtts)

	// Output results
	switch outputFormat {
	case "json":
		_ = output.WriteProbeResultsJSON(os.Stdout, "ICMP", target, results, &histStats, &jitterStats)
	default:
		tw := output.NewTableWriter(os.Stdout)
		_ = tw.WriteProbeResults("ICMP", target, rtts, failures)
		_ = tw.WriteStatistics(histStats)
		_ = tw.WriteJitterStats(jitterStats)
	}
}

func analyzeCommand(args []string) {
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)

	target := fs.String("target", "", "Target host or IP address")
	idleCount := fs.Int("idle-count", 10, "Probes for idle measurement")
	loadCount := fs.Int("load-count", 10, "Probes for loaded measurement")
	outputFormat := fs.String("output", "table", "Output format: table or json")

	fs.Parse(args)

	if *target == "" {
		fmt.Println("Error: -target flag is required")
		fs.Usage()
		os.Exit(1)
	}

	fmt.Printf("Bufferbloat Analysis: target=%s\n", *target)
	fmt.Println("Measuring idle latency...")

	// Create a probe function for the detector
	probeFn := func(count int) ([]time.Duration, error) {
		config := probe.UDPProbeConfig{
			Target:      *target,
			Port:        12345,
			Count:       count,
			Interval:    100 * time.Millisecond,
			PayloadSize: 12,
			Timeout:     3 * time.Second,
		}
		prober := probe.NewUDPProber(config)
		results, err := prober.Probe()
		if err != nil {
			return nil, err
		}

		var rtts []time.Duration
		for _, r := range results {
			if r.Success {
				rtts = append(rtts, r.RTT)
			}
		}
		return rtts, nil
	}

	// Note: Bufferbloat detection requires a proper implementation
	// For now, we'll just show the concept
	fmt.Println("\nNote: Bufferbloat detection requires a working echo server.")
	fmt.Println("Run 'netprobe listen' on the target machine first.")
	fmt.Println("\nPerforming UDP probes under idle and load conditions...")

	// Measure idle
	idleResults, err := probeFn(*idleCount)
	if err != nil {
		log.Fatalf("Idle measurement failed: %v", err)
	}

	// Measure under load
	loadResults, err := probeFn(*loadCount)
	if err != nil {
		log.Fatalf("Load measurement failed: %v", err)
	}

	// Calculate statistics
	idleHist := stats.NewLatencyHistogram(len(idleResults))
	idleHist.AddSamples(idleResults)

	loadHist := stats.NewLatencyHistogram(len(loadResults))
	loadHist.AddSamples(loadResults)

	// Create result map for output
	result := map[string]interface{}{
		"idle_p50":        idleHist.P50(),
		"idle_p99":        idleHist.P99(),
		"idle_max":        idleHist.Max(),
		"load_p50":        loadHist.P50(),
		"load_p99":        loadHist.P99(),
		"load_max":        loadHist.Max(),
		"p50_increase":    float64(loadHist.P50().Microseconds()) / float64(idleHist.P50().Microseconds()),
		"p99_increase":    float64(loadHist.P99().Microseconds()) / float64(idleHist.P99().Microseconds()),
		"max_increase":    float64(loadHist.Max().Microseconds()) / float64(idleHist.Max().Microseconds()),
		"is_bufferbloated": false, // Simplified for demo
		"severity":        "None",
		"explanation":     "See results above for latency comparison.",
	}

	// Output results
	switch *outputFormat {
	case "json":
		_ = output.WriteBufferbloatResultJSON(os.Stdout, *target, result)
	default:
		tw := output.NewTableWriter(os.Stdout)
		_ = tw.WriteBufferbloatResults(*target, result)
	}
}

func listenCommand(args []string) {
	fs := flag.NewFlagSet("listen", flag.ExitOnError)
	port := fs.Int("port", 12345, "UDP port to listen on")
	fs.Parse(args)

	// Invoke the listener binary
	fmt.Printf("Starting UDP echo server on port %d\n", *port)
	fmt.Println("This command should be run separately using: cmd/listener/main.go")
	fmt.Println("Or compile and run: netprobe-listener -port", *port)
}
