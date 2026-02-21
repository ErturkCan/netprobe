package probe

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// ICMPProbeConfig holds configuration for ICMP probes
type ICMPProbeConfig struct {
	Target   string        // Target host or IP
	Count    int           // Number of probes to send
	Interval time.Duration // Time between probes
	Timeout  time.Duration // Timeout for responses
	PacketID int           // ICMP packet ID
}

// ICMPProbeResult holds results from a single ICMP probe
type ICMPProbeResult struct {
	Sequence int
	RTT      time.Duration
	Success  bool
	Error    error
}

// ICMPProber performs ICMP echo (ping) probes
type ICMPProber struct {
	config ICMPProbeConfig
}

// NewICMPProber creates a new ICMP prober
func NewICMPProber(config ICMPProbeConfig) *ICMPProber {
	if config.Count == 0 {
		config.Count = 10
	}
	if config.Interval == 0 {
		config.Interval = 1 * time.Second
	}
	if config.Timeout == 0 {
		config.Timeout = 3 * time.Second
	}
	if config.PacketID == 0 {
		config.PacketID = os.Getpid() & 0xffff
	}

	return &ICMPProber{config: config}
}

// Probe performs a series of ICMP echo probes
func (p *ICMPProber) Probe() ([]ICMPProbeResult, error) {
	results := make([]ICMPProbeResult, 0, p.config.Count)

	// Resolve target
	addr, err := net.ResolveIPAddr("ip4", p.config.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %w", err)
	}

	// Create ICMP connection
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, fmt.Errorf("failed to create ICMP listener: %w", err)
	}
	defer conn.Close()

	// Send probes
	for i := 0; i < p.config.Count; i++ {
		if i > 0 {
			time.Sleep(p.config.Interval)
		}

		result := p.sendProbe(conn, addr, i+1)
		results = append(results, result)
	}

	return results, nil
}

// sendProbe sends a single ICMP echo request and measures RTT
func (p *ICMPProber) sendProbe(conn *icmp.PacketConn, addr *net.IPAddr, sequence int) ICMPProbeResult {
	result := ICMPProbeResult{
		Sequence: sequence,
	}

	// Create ICMP echo request
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   p.config.PacketID,
			Seq:  sequence,
			Data: []byte("netprobe"),
		},
	}

	// Marshal message
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		result.Error = fmt.Errorf("failed to marshal ICMP message: %w", err)
		return result
	}

	// Send request
	sendTime := time.Now()
	_, err = conn.WriteTo(msgBytes, addr)
	if err != nil {
		result.Error = fmt.Errorf("send failed: %w", err)
		return result
	}

	// Receive response with timeout
	conn.SetReadDeadline(time.Now().Add(p.config.Timeout))
	reply := make([]byte, 1500)
	_, _, err = conn.ReadFrom(reply)
	if err != nil {
		result.Error = fmt.Errorf("receive failed: %w", err)
		return result
	}

	receiveTime := time.Now()
	result.RTT = receiveTime.Sub(sendTime)
	result.Success = true

	return result
}
