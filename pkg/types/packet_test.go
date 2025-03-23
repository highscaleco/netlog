package types

import (
	"testing"
	"time"
)

func TestPacketInfo(t *testing.T) {
	now := time.Now()
	packet := PacketInfo{
		Timestamp:   now,
		Source:      "192.168.1.100",
		Destination: "8.8.8.8",
		Protocol:    "TCP",
		Port:        "443",
		Bytes:       "1000",
	}

	// Test string representation
	expected := "192.168.1.100 => 8.8.8.8 TCP 443 1000 bytes"
	if packet.String() != expected {
		t.Errorf("PacketInfo.String() = %v, want %v", packet.String(), expected)
	}

	// Test JSON-like string representation
	jsonStr := packet.JSONString()
	if jsonStr == "" {
		t.Error("PacketInfo.JSONString() returned empty string")
	}
}

func TestAggregatedInfo(t *testing.T) {
	now := time.Now()
	agg := AggregatedInfo{
		StartTime:   now,
		EndTime:     now.Add(5 * time.Second),
		Source:      "192.168.1.100",
		Destination: "8.8.8.8",
		Protocol:    "TCP",
		Port:        "443",
		TotalBytes:  500000000,
		Packets:     50000,
	}

	// Test string representation
	expected := "192.168.1.100 => 8.8.8.8 TCP 443 500000000 bytes (50000 packets in 5.00s)"
	if agg.String() != expected {
		t.Errorf("AggregatedInfo.String() = %v, want %v", agg.String(), expected)
	}

	// Test JSON-like string representation
	jsonStr := agg.JSONString()
	if jsonStr == "" {
		t.Error("AggregatedInfo.JSONString() returned empty string")
	}
}
