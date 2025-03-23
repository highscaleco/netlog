package capture

import (
	"net"
	"os"
	"testing"
	"time"
)

func TestIsPublicIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       net.IP
		expected bool
	}{
		{
			name:     "Public IP",
			ip:       net.ParseIP("8.8.8.8"),
			expected: true,
		},
		{
			name:     "Private IP",
			ip:       net.ParseIP("192.168.1.1"),
			expected: false,
		},
		{
			name:     "Localhost",
			ip:       net.ParseIP("127.0.0.1"),
			expected: false,
		},
		{
			name:     "Link Local",
			ip:       net.ParseIP("169.254.0.1"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPublicIP(tt.ip)
			if result != tt.expected {
				t.Errorf("IsPublicIP(%v) = %v, want %v", tt.ip, result, tt.expected)
			}
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
			name:     "Low rate",
			bytes:    1000000, // 1 MB
			seconds:  1.0,
			expected: time.Second,
		},
		{
			name:     "Medium rate",
			bytes:    20000000, // 20 MB
			seconds:  1.0,
			expected: 3 * time.Second,
		},
		{
			name:     "High rate",
			bytes:    200000000, // 200 MB
			seconds:  1.0,
			expected: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateWindowSize(tt.bytes, tt.seconds)
			if result != tt.expected {
				t.Errorf("calculateWindowSize(%v, %v) = %v, want %v", tt.bytes, tt.seconds, result, tt.expected)
			}
		})
	}
}

func TestNewCapture(t *testing.T) {
	capture := NewCapture("test0")
	if capture == nil {
		t.Error("NewCapture returned nil")
	}
	if capture.iface != "test0" {
		t.Errorf("NewCapture iface = %v, want test0", capture.iface)
	}
}

func TestStartStop(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("Test requires root privileges")
	}

	capture := NewCapture("test0")
	if capture == nil {
		t.Fatal("Failed to create capture")
	}

	// Start capture
	packets, err := capture.Start()
	if err != nil {
		t.Fatalf("Failed to start capture: %v", err)
	}

	// Stop capture
	capture.Stop()

	// Verify channel is closed
	select {
	case _, ok := <-packets:
		if ok {
			t.Error("Channel should be closed")
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for channel to close")
	}
}
