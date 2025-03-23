package main

import (
	"testing"
)

func TestFormatFlag(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{
			name:     "Valid JSON format",
			format:   "json",
			expected: "json",
		},
		{
			name:     "Valid text format",
			format:   "text",
			expected: "text",
		},
		{
			name:     "Invalid format",
			format:   "invalid",
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			FormatFlag = tt.format
			if FormatFlag != tt.expected {
				t.Errorf("FormatFlag = %v, want %v", FormatFlag, tt.expected)
			}
		})
	}
}

func TestInterfaceFlag(t *testing.T) {
	tests := []struct {
		name     string
		iface    string
		expected string
	}{
		{
			name:     "Default interface",
			iface:    "",
			expected: "en1",
		},
		{
			name:     "Custom interface",
			iface:    "eth0",
			expected: "eth0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.iface != "" {
				InterfaceFlag = tt.iface
			}
			if InterfaceFlag != tt.expected {
				t.Errorf("InterfaceFlag = %v, want %v", InterfaceFlag, tt.expected)
			}
		})
	}
}
