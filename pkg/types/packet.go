package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// PacketInfo represents information about a captured network packet
type PacketInfo struct {
	Timestamp   time.Time
	Source      string
	Destination string
	Protocol    string
	Port        string
	Bytes       string
}

// String returns a human-readable string representation of the packet
func (p PacketInfo) String() string {
	return fmt.Sprintf("%s => %s %s %s %s bytes", p.Source, p.Destination, p.Protocol, p.Port, p.Bytes)
}

// JSONString returns a JSON-like string representation of the packet
func (p PacketInfo) JSONString() string {
	data := struct {
		Timestamp   string `json:"timestamp"`
		Source      string `json:"source"`
		Destination string `json:"destination"`
		Protocol    string `json:"protocol"`
		Port        string `json:"port"`
		Bytes       string `json:"bytes"`
	}{
		Timestamp:   p.Timestamp.Format("2006-01-02 15:04:05.999"),
		Source:      p.Source,
		Destination: p.Destination,
		Protocol:    p.Protocol,
		Port:        p.Port,
		Bytes:       p.Bytes,
	}
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

// AggregatedInfo represents aggregated packet information
type AggregatedInfo struct {
	StartTime   time.Time
	EndTime     time.Time
	Source      string
	Destination string
	Protocol    string
	Port        string
	TotalBytes  int64
	Packets     int64
}

// String returns a human-readable string representation of the aggregated info
func (a AggregatedInfo) String() string {
	duration := a.EndTime.Sub(a.StartTime).Seconds()
	return fmt.Sprintf("%s => %s %s %s %d bytes (%d packets in %.2fs)",
		a.Source, a.Destination, a.Protocol, a.Port, a.TotalBytes, a.Packets, duration)
}

// JSONString returns a JSON-like string representation of the aggregated info
func (a AggregatedInfo) JSONString() string {
	duration := a.EndTime.Sub(a.StartTime).Seconds()
	data := struct {
		Timestamp   string `json:"timestamp"`
		Duration    string `json:"duration"`
		Source      string `json:"source"`
		Destination string `json:"destination"`
		Protocol    string `json:"protocol"`
		Port        string `json:"port"`
		TotalBytes  int64  `json:"total_bytes"`
		Packets     int64  `json:"packets"`
	}{
		Timestamp:   a.StartTime.Format("2006-01-02 15:04:05.999"),
		Duration:    fmt.Sprintf("%.2fs", duration),
		Source:      a.Source,
		Destination: a.Destination,
		Protocol:    a.Protocol,
		Port:        a.Port,
		TotalBytes:  a.TotalBytes,
		Packets:     a.Packets,
	}
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}
