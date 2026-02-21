package probe

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/ErturkCan/netprobe/internal"
)

// UDPProbeConfig holds configuration for UDP probes
type UDPProbeConfig struct {
	Target      string        // Target host or IP
	Port        int           // Target port
	Count       int           // Number of probes to send
	Interval    time.Duration // Time between probes
	PayloadSize int           // Size of payload in bytes (minimum 12 for timestamp)
	Timeout     time.Duration // Timeout for responses
}

// UDPProbeResult holds results from a single probe
type UDPProbeResult struct {
	Sequence   uint32
	RTT        time.Duration
	PayloadLen int
	Success    bool
	Error      error
}

// UDPProber performs UDP echo probes
type UDPProber struct {
	config UDPProbeConfig
}

// NewUDPProber creates a new UDP prober
func NewUDPProber(config UDPProbeConfig) *UDPProber {
	if config.Count == 0 {
		config.Count = 10
	}
	if config.Port == 0 {
		config.Port = 12345
	}
	if config.Interval == 0 {
		config.Interval = 1 * time.Second
	}
	if config.PayloadSize < 12 {
		config.PayloadSize = 12
	}
	if config.Timeout == 0 {
		config.Timeout = 3 * time.Second
	}

	return &UDPProber{config: config}
}

// Probe performs a series of UDP echo probes
func (p *UDPProber) Probe() ([]UDPProbeResult, error) {
	results := make([]UDPProbeResult, 0, p.config.Count)

	// Resolve target address
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", p.config.Target, p.config.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %w", err)
	}

	// Create UDP connection
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial UDP: %w", err)
	}
	defer conn.Close()

	// Set read deadline for all operations
	baseDeadline := time.Now().Add(time.Duration(p.config.Count) * (p.config.Interval + p.config.Timeout))
	if err := conn.SetReadDeadline(baseDeadline); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Send probes
	for i := 0; i < p.config.Count; i++ {
		if i > 0 {
			time.Sleep(p.config.Interval)
		}

		result := p.sendProbe(conn, uint32(i+1))
		results = append(results, result)
	}

	return results, nil
}

// sendProbe sends a single UDP probe and measures RTT
func (p *UDPProber) sendProbe(conn *net.UDPConn, sequence uint32) UDPProbeResult {
	result := UDPProbeResult{
		Sequence: sequence,
	}

	// Prepare payload: [4 bytes sequence][8 bytes timestamp][variable payload]
	payload := make([]byte, p.config.PayloadSize)
	binary.BigEndian.PutUint32(payload[0:4], sequence)
	binary.BigEndian.PutUint64(payload[4:12], uint64(internal.NowNano()))

	// Send probe
	sendTime := time.Now()
	_, err := conn.Write(payload)
	if err != nil {
		result.Error = fmt.Errorf("send failed: %w", err)
		return result
	}

	// Receive response
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		result.Error = fmt.Errorf("receive failed: %w", err)
		return result
	}

	receiveTime := time.Now()
	result.RTT = receiveTime.Sub(sendTime)
	result.PayloadLen = n
	result.Success = true

	return result
}
