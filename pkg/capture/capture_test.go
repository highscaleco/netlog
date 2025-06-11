package capture

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/highscaleco/netlog/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestIsPublicIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{
			name:     "public IP",
			ip:       "8.8.8.8",
			expected: true,
		},
		{
			name:     "private IP",
			ip:       "192.168.1.1",
			expected: false,
		},
		{
			name:     "loopback IP",
			ip:       "127.0.0.1",
			expected: false,
		},
		{
			name:     "link local IP",
			ip:       "169.254.0.1",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			result := IsPublicIP(ip)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateWindowSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		seconds  float64
		expected time.Duration
	}{
		{
			name:     "high bandwidth",
			bytes:    200 * 1024 * 1024, // 200MB
			seconds:  1,
			expected: 5 * time.Second,
		},
		{
			name:     "medium bandwidth",
			bytes:    20 * 1024 * 1024, // 20MB
			seconds:  1,
			expected: 3 * time.Second,
		},
		{
			name:     "low bandwidth",
			bytes:    2 * 1024 * 1024, // 2MB
			seconds:  1,
			expected: 2 * time.Second,
		},
		{
			name:     "very low bandwidth",
			bytes:    100 * 1024, // 100KB
			seconds:  1,
			expected: time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateWindowSize(tt.bytes, tt.seconds)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewCapture(t *testing.T) {
	capture := NewCapture(
		"eth0",
		DefaultBufferSize,
		DefaultPromiscuous,
		DefaultTimeout,
		DefaultFilter,
		DefaultMaxPacketSize,
		DefaultMaxConnections,
	)

	assert.NotNil(t, capture)
	assert.Equal(t, "eth0", capture.iface)
	assert.Equal(t, DefaultBufferSize, capture.bufferSize)
	assert.Equal(t, DefaultPromiscuous, capture.promiscuous)
	assert.Equal(t, DefaultTimeout, capture.timeout)
	assert.Equal(t, DefaultFilter, capture.filter)
	assert.Equal(t, DefaultMaxPacketSize, capture.maxPacketSize)
	assert.Equal(t, DefaultMaxConnections, capture.maxConnections)
	assert.NotNil(t, capture.packets)
	assert.NotNil(t, capture.stop)
	assert.NotNil(t, capture.aggregatedInfo)
}

func TestCaptureStartStop(t *testing.T) {
	capture := NewCapture(
		"eth0",
		DefaultBufferSize,
		DefaultPromiscuous,
		DefaultTimeout,
		DefaultFilter,
		DefaultMaxPacketSize,
		DefaultMaxConnections,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := capture.Start(ctx)
	assert.NoError(t, err)

	// Wait for a short time to let the goroutines start
	time.Sleep(100 * time.Millisecond)

	capture.Stop()

	// Wait for goroutines to stop
	time.Sleep(100 * time.Millisecond)
}

func TestAggregatePacket(t *testing.T) {
	capture := NewCapture(
		"eth0",
		DefaultBufferSize,
		DefaultPromiscuous,
		DefaultTimeout,
		DefaultFilter,
		DefaultMaxPacketSize,
		DefaultMaxConnections,
	)

	// Create a test packet
	ip := &layers.IPv4{
		SrcIP: net.ParseIP("192.168.1.1"),
		DstIP: net.ParseIP("8.8.8.8"),
	}
	tcp := &layers.TCP{
		SrcPort: 12345,
		DstPort: 80,
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}
	err := gopacket.SerializeLayers(buf, opts, ip, tcp)
	assert.NoError(t, err)

	packet := gopacket.NewPacket(buf.Bytes(), layers.LinkTypeEthernet, gopacket.Default)

	// Process the packet
	capture.mu.Lock()
	key := fmt.Sprintf("%s:%s:%s:%s", ip.SrcIP, ip.DstIP, tcp.LayerType(), tcp.TransportFlow().Src().String())
	agg, exists := capture.aggregatedInfo[key]
	assert.False(t, exists)

	// Update aggregation
	agg = &types.AggregatedInfo{
		StartTime:   packet.Metadata().Timestamp,
		EndTime:     packet.Metadata().Timestamp,
		Source:      ip.SrcIP.String(),
		Destination: ip.DstIP.String(),
		Protocol:    tcp.LayerType().String(),
		Port:        tcp.TransportFlow().Src().String(),
		LastSeen:    time.Now(),
	}
	capture.aggregatedInfo[key] = agg
	capture.mu.Unlock()

	// Verify aggregation
	capture.mu.RLock()
	agg, exists = capture.aggregatedInfo[key]
	assert.True(t, exists)
	assert.Equal(t, ip.SrcIP.String(), agg.Source)
	assert.Equal(t, ip.DstIP.String(), agg.Destination)
	assert.Equal(t, tcp.LayerType().String(), agg.Protocol)
	assert.Equal(t, tcp.TransportFlow().Src().String(), agg.Port)
	capture.mu.RUnlock()
}
