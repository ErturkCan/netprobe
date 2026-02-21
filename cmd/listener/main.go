package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	port := flag.Int("port", 12345, "UDP port to listen on")
	flag.Parse()

	// Create UDP listener
	addr := net.UDPAddr{
		Port: *port,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("Failed to listen on UDP %d: %v", *port, err)
	}
	defer conn.Close()

	log.Printf("UDP Echo Server listening on :%d", *port)
	log.Println("Ready to receive probes. Press Ctrl+C to stop.")

	buffer := make([]byte, 4096)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Read error: %v", err)
			continue
		}

		// Extract sequence and send time from payload
		var sequence uint32
		var sendTime int64

		if n >= 12 {
			sequence = binary.BigEndian.Uint32(buffer[0:4])
			sendTime = int64(binary.BigEndian.Uint64(buffer[4:12]))
		}

		// Get current timestamp
		now := time.Now().UnixNano()

		// Log the probe
		rtt := (now - sendTime) / 1000 // Convert to microseconds
		fmt.Printf("[%s] Seq=%d Payload=%d bytes RTT=%.3fms\n",
			remoteAddr.IP.String(),
			sequence,
			n,
			float64(rtt)/1000.0,
		)

		// Echo the packet back
		_, err = conn.WriteToUDP(buffer[:n], remoteAddr)
		if err != nil {
			log.Printf("Write error: %v", err)
		}
	}
}
