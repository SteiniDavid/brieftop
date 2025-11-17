package monitor

import "testing"

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    uint64
		expected string
	}{
		{"Zero bytes", 0, "0 B"},
		{"Bytes", 512, "512 B"},
		{"1 KB", 1024, "1.0 KB"},
		{"1.5 KB", 1536, "1.5 KB"},
		{"1 MB", 1024 * 1024, "1.0 MB"},
		{"100 MB", 100 * 1024 * 1024, "100.0 MB"},
		{"1 GB", 1024 * 1024 * 1024, "1.0 GB"},
		{"1.5 GB", 1536 * 1024 * 1024, "1.5 GB"},
		{"2 TB", 2 * 1024 * 1024 * 1024 * 1024, "2.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatBytes(%d) = %s; expected %s", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestFormatCPU(t *testing.T) {
	tests := []struct {
		name     string
		percent  float64
		expected string
	}{
		{"Zero percent", 0.0, "0.0%"},
		{"Low percent", 5.2, "5.2%"},
		{"Medium percent", 45.7, "45.7%"},
		{"High percent", 99.9, "99.9%"},
		{"Full percent", 100.0, "100.0%"},
		{"Over 100", 150.5, "150.5%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCPU(tt.percent)
			if result != tt.expected {
				t.Errorf("FormatCPU(%f) = %s; expected %s", tt.percent, result, tt.expected)
			}
		})
	}
}
