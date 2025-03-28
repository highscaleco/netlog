package types

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/highscaleco/netlog/pkg/k8s"
	"github.com/highscaleco/netlog/pkg/redis"
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
	Namespace   string
	Name        string
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
	ofipSrc, errSrc := GetNamespaceAndNameByIPv4(a.Source)
	ofipDst, errDst := GetNamespaceAndNameByIPv4(a.Destination)

	// Reset namespace and name only if both calls fail
	if errSrc != nil && errDst != nil {
		a.Namespace = ""
		a.Name = ""
	} else {
		if errSrc == nil && ofipSrc != nil && ofipSrc.Namespace != "" {
			a.Namespace = ofipSrc.Namespace
			a.Name = ofipSrc.Name
		} else if errDst == nil && ofipDst != nil && ofipDst.Namespace != "" {
			a.Namespace = ofipDst.Namespace
			a.Name = ofipDst.Name
		}
	}

	// Return empty string if no namespace is found
	if a.Namespace == "" {
		return ""
	}

	duration := a.EndTime.Sub(a.StartTime).Seconds()
	return fmt.Sprintf("%s %s %s %s => %s %s %s %d bytes (%d packets in %.2fs)",
		a.StartTime, a.Namespace, a.Name, a.Source, a.Destination, a.Protocol, a.Port, a.TotalBytes, a.Packets, duration)
}

// JSONString returns a JSON-like string representation of the aggregated info
func (a AggregatedInfo) JSONString() string {
	ofipSrc, errSrc := GetNamespaceAndNameByIPv4(a.Source)
	ofipDst, errDst := GetNamespaceAndNameByIPv4(a.Destination)

	// Reset namespace and name only if both calls fail
	if errSrc != nil && errDst != nil {
		a.Namespace = ""
		a.Name = ""
	} else {
		if errSrc == nil && ofipSrc != nil && ofipSrc.Namespace != "" {
			a.Namespace = ofipSrc.Namespace
			a.Name = ofipSrc.Name
		} else if errDst == nil && ofipDst != nil && ofipDst.Namespace != "" {
			a.Namespace = ofipDst.Namespace
			a.Name = ofipDst.Name
		}
	}

	// Return empty string if no namespace is found
	if a.Namespace == "" {
		return ""
	}

	duration := a.EndTime.Sub(a.StartTime).Seconds()
	data := struct {
		Timestamp   string `json:"timestamp"`
		Namespace   string `json:"namespace"`
		Name        string `json:"name"`
		Duration    string `json:"duration"`
		Source      string `json:"source"`
		Destination string `json:"destination"`
		Protocol    string `json:"protocol"`
		Port        string `json:"port"`
		TotalBytes  int64  `json:"total_bytes"`
		Packets     int64  `json:"packets"`
	}{
		Timestamp:   a.StartTime.Format("2006-01-02 15:04:05.999"),
		Namespace:   a.Namespace,
		Name:        a.Name,
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

type OFIP struct {
	Namespace string
	Name      string
}

func GetNamespaceAndNameByIPv4(ipv4 string) (*OFIP, error) {
	if ipv4 == "" {
		return nil, fmt.Errorf("ipv4 cannot be empty")
	}

	// Try to get from Redis first
	info, err := redis.GetIP(ipv4)
	if err == nil && info.Namespace != "" {
		return &OFIP{
			Namespace: info.Namespace,
			Name:      info.Name,
		}, nil
	}

	// If Redis fails or no data found, try K8s
	ofip, err := k8s.GetOFIPByIPv4(ipv4)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace and name by ipv4: %w", err)
	}

	// Parse the ofip string (format: namespace-name)
	parts := regexp.MustCompile(`-`).Split(ofip, 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ofip format: %s", ofip)
	}

	// Store in Redis for future use
	redisInfo := redis.IPInfo{
		Namespace: parts[0],
		Name:      parts[1],
	}
	if err := redis.SetIP(ipv4, redisInfo); err != nil {
		// Log the error but don't fail the operation
		fmt.Printf("warning: failed to cache IP info in Redis: %v\n", err)
	}

	return &OFIP{
		Namespace: parts[0],
		Name:      parts[1],
	}, nil
}
