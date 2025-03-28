package capture

import (
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
}

// Capture represents a packet capture session
type Capture struct {
	iface string
	stop  chan struct{}
}

// NewCapture creates a new packet capture session
func NewCapture(iface string) *Capture {
	return &Capture{
		iface: iface,
		stop:  make(chan struct{}),
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

// calculateWindowSize determines the appropriate aggregation window based on transfer rate
func calculateWindowSize(bytes int64, seconds float64) time.Duration {
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

// Start begins capturing packets
func (c *Capture) Start() (chan types.AggregatedInfo, error) {
	handle, err := pcap.OpenLive(c.iface, 65536, true, pcap.BlockForever)
	if err != nil {
		return nil, fmt.Errorf("error opening interface: %v", err)
	}

	packets := make(chan types.AggregatedInfo)
	aggregation := make(map[string]*types.AggregatedInfo)
	var mu sync.Mutex

	go func() {
		defer handle.Close()
		defer close(packets)

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
				mu.Lock()
				agg, exists := aggregation[key]
				if !exists {
					agg = &types.AggregatedInfo{
						StartTime:   packet.Metadata().Timestamp,
						EndTime:     packet.Metadata().Timestamp,
						Source:      ip.SrcIP.String(),
						Destination: ip.DstIP.String(),
						Protocol:    transportLayer.LayerType().String(),
						Port:        transportLayer.TransportFlow().Src().String(),
					}
					aggregation[key] = agg
				} else {
					agg.EndTime = packet.Metadata().Timestamp
					agg.TotalBytes += int64(len(packet.Data()))
					agg.Packets++
				}
				mu.Unlock()

			case <-ticker.C:
				// Send aggregated packets
				mu.Lock()
				for key, agg := range aggregation {
					duration := agg.EndTime.Sub(agg.StartTime).Seconds()
					if duration >= 1.0 { // Only send if we have at least 1 second of data
						windowSize := calculateWindowSize(agg.TotalBytes, duration)
						if duration >= windowSize.Seconds() {
							packets <- *agg
							delete(aggregation, key)
						}
					}
				}
				mu.Unlock()
			}
		}
	}()

	return packets, nil
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
