package capture

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/highscaleco/netlog/pkg/types"
)

// PacketInfo represents the captured packet information
type PacketInfo struct {
	Timestamp   time.Time
	Source      string
	Destination string
	Protocol    string
	Port        string
	Bytes       int
}

// AggregatedInfo represents aggregated packet information
type AggregatedInfo struct {
	StartTime   time.Time
	EndTime     time.Time
	Source      string
	Destination string
	Protocol    string
	Port        string
	TotalBytes  int
	PacketCount int
	Namespace   string
	Name        string
	Direction   string
	LastSeen    time.Time
}

// Capture represents a packet capture session
type Capture struct {
	iface          string
	bufferSize     int
	promiscuous    bool
	timeout        time.Duration
	filter         string
	maxPacketSize  int
	maxConnections int
	packets        chan types.AggregatedInfo
	stop           chan struct{}
	handle         *pcap.Handle
	mu             sync.RWMutex
	aggregatedInfo map[string]*types.AggregatedInfo
}

const (
	// DefaultInterface is the default network interface to capture on
	DefaultInterface = "eth0"
	// DefaultBufferSize is the default buffer size for packet capture
	DefaultBufferSize = 65536
	// DefaultPromiscuous is the default promiscuous mode setting
	DefaultPromiscuous = true
	// DefaultTimeout is the default timeout for packet capture
	DefaultTimeout = pcap.BlockForever
	// DefaultFilter is the default BPF filter
	DefaultFilter = "tcp or udp"
	// DefaultMaxPacketSize is the default maximum packet size
	DefaultMaxPacketSize = 65536
	// DefaultMaxConnections is the default maximum number of connections to track
	DefaultMaxConnections = 10000
	// DefaultConnectionTimeout is the default timeout for connections
	DefaultConnectionTimeout = 5 * time.Minute
	// DefaultCleanupInterval is the default interval for cleaning up old connections
	DefaultCleanupInterval = 1 * time.Minute
)

// NewCapture creates a new packet capture session
func NewCapture(iface string, bufferSize int, promiscuous bool, timeout time.Duration, filter string, maxPacketSize, maxConnections int) *Capture {
	return &Capture{
		iface:          iface,
		bufferSize:     bufferSize,
		promiscuous:    promiscuous,
		timeout:        timeout,
		filter:         filter,
		maxPacketSize:  maxPacketSize,
		maxConnections: maxConnections,
		packets:        make(chan types.AggregatedInfo, 1000),
		stop:           make(chan struct{}),
		aggregatedInfo: make(map[string]*types.AggregatedInfo),
	}
}

// IsPublicIP checks if an IP address is public
func IsPublicIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return false
	}

	// Check if it's a private network
	if ip.IsPrivate() {
		return false
	}

	return true
}

// CalculateWindowSize calculates the window size based on the bytes and seconds
func CalculateWindowSize(bytes int64, seconds float64) time.Duration {
	bytesPerSecond := float64(bytes) / seconds
	megabytesPerSecond := bytesPerSecond / (1024 * 1024)

	switch {
	case megabytesPerSecond > 100:
		return 5 * time.Second
	case megabytesPerSecond > 10:
		return 3 * time.Second
	case megabytesPerSecond > 1:
		return 2 * time.Second
	default:
		return time.Second
	}
}

// Start starts capturing packets
func (c *Capture) Start(ctx context.Context) error {
	// Start cleanup goroutine
	go c.cleanupLoop(ctx)

	// Start packet processing goroutine
	go c.processPackets(ctx)

	// Start packet capture
	_ = c.handle.LinkType() // Ignore the return value as it's not an error
	return nil
}

// cleanupLoop runs the cleanup function periodically
func (c *Capture) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(DefaultCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.cleanup()
		}
	}
}

// cleanup removes old connections from the aggregated info map
func (c *Capture) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, info := range c.aggregatedInfo {
		if now.Sub(info.EndTime) > DefaultConnectionTimeout {
			delete(c.aggregatedInfo, key)
		}
	}
}

// processPackets processes packets and updates the aggregated info map
func (c *Capture) processPackets(ctx context.Context) {
	handle, err := pcap.OpenLive(c.iface, 65536, true, pcap.BlockForever)
	if err != nil {
		fmt.Printf("Error opening interface: %v\n", err)
		return
	}
	defer handle.Close()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-c.stop:
			return
		case packet := <-packetSource.Packets():
			// Process packet
			ipLayer := packet.NetworkLayer()
			if ipLayer == nil {
				continue
			}

			ip, ok := ipLayer.(*layers.IPv4)
			if !ok {
				continue
			}

			// Only process packets with public IPs
			if !IsPublicIP(ip.SrcIP) && !IsPublicIP(ip.DstIP) {
				continue
			}

			// Get transport layer info
			transportLayer := packet.TransportLayer()
			if transportLayer == nil {
				continue
			}

			// Create connection key
			key := fmt.Sprintf("%s:%s:%s:%s", ip.SrcIP, ip.DstIP, transportLayer.LayerType(), transportLayer.TransportFlow().Src().String())

			// Update aggregation
			c.mu.Lock()
			agg, exists := c.aggregatedInfo[key]
			if !exists {
				// Try to get namespace and name from source IP first
				ofipSrc, errSrc := types.GetNamespaceAndNameByIPv4(ip.SrcIP.String())
				ofipDst, errDst := types.GetNamespaceAndNameByIPv4(ip.DstIP.String())

				// Set namespace, name, and direction based on which IP is in our cluster
				var namespace, name, direction string
				if errSrc == nil && ofipSrc != nil && ofipSrc.Namespace != "" {
					namespace = ofipSrc.Namespace
					name = ofipSrc.Name
					direction = "outbound"
				} else if errDst == nil && ofipDst != nil && ofipDst.Namespace != "" {
					namespace = ofipDst.Namespace
					name = ofipDst.Name
					direction = "inbound"
				}

				agg = &types.AggregatedInfo{
					StartTime:   packet.Metadata().Timestamp,
					EndTime:     packet.Metadata().Timestamp,
					Source:      ip.SrcIP.String(),
					Destination: ip.DstIP.String(),
					Protocol:    transportLayer.LayerType().String(),
					Port:        transportLayer.TransportFlow().Src().String(),
					Namespace:   namespace,
					Name:        name,
					Direction:   direction,
					LastSeen:    time.Now(),
				}
				c.aggregatedInfo[key] = agg
			} else {
				agg.EndTime = packet.Metadata().Timestamp
				agg.TotalBytes += int64(len(packet.Data()))
				agg.Packets++
				agg.LastSeen = time.Now()
			}
			c.mu.Unlock()

		case <-ticker.C:
			// Send aggregated packets
			c.mu.Lock()
			for key, agg := range c.aggregatedInfo {
				duration := agg.EndTime.Sub(agg.StartTime).Seconds()
				if duration >= 1.0 { // Only send if we have at least 1 second of data
					windowSize := CalculateWindowSize(int64(agg.TotalBytes), duration)
					if duration >= windowSize.Seconds() {
						c.packets <- *agg
						delete(c.aggregatedInfo, key)
					}
				}
			}
			c.mu.Unlock()
		}
	}
}

// Stop stops the packet capture
func (c *Capture) Stop() {
	close(c.stop)
}

// processPacket extracts relevant information from a packet
// func (c *Capture) processPacket(packet gopacket.Packet) (PacketInfo, error) {
// 	networkLayer := packet.NetworkLayer()
// 	if networkLayer == nil {
// 		return PacketInfo{}, fmt.Errorf("no network layer found")
// 	}

// 	transportLayer := packet.TransportLayer()
// 	if transportLayer == nil {
// 		return PacketInfo{}, fmt.Errorf("no transport layer found")
// 	}

// 	return PacketInfo{
// 		Timestamp:   packet.Metadata().Timestamp,
// 		Source:      networkLayer.NetworkFlow().Src().String(),
// 		Destination: networkLayer.NetworkFlow().Dst().String(),
// 		Protocol:    transportLayer.LayerType().String(),
// 		Port:        transportLayer.TransportFlow().Dst().String(),
// 		Bytes:       len(packet.Data()),
// 	}, nil
// }

// Packets returns the channel for receiving aggregated packets
func (c *Capture) Packets() chan types.AggregatedInfo {
	return c.packets
}
