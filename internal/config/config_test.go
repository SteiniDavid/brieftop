package config

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	cfg := New()

	if cfg == nil {
		t.Fatal("New() returned nil")
	}

	// Test default values
	if cfg.CPUThreshold != 5.0 {
		t.Errorf("Expected CPUThreshold to be 5.0, got %v", cfg.CPUThreshold)
	}

	expectedMemory := uint64(50 * 1024 * 1024)
	if cfg.MemoryThreshold != expectedMemory {
		t.Errorf("Expected MemoryThreshold to be %d, got %d", expectedMemory, cfg.MemoryThreshold)
	}

	if cfg.RefreshRate != time.Second {
		t.Errorf("Expected RefreshRate to be 1 second, got %v", cfg.RefreshRate)
	}

	if !cfg.ShowThreads {
		t.Error("Expected ShowThreads to be true")
	}
}

func TestSetCPUThreshold(t *testing.T) {
	cfg := New()

	testCases := []float64{0.0, 10.5, 50.0, 100.0}
	for _, threshold := range testCases {
		cfg.SetCPUThreshold(threshold)
		if cfg.GetCPUThreshold() != threshold {
			t.Errorf("Expected CPUThreshold to be %v, got %v", threshold, cfg.GetCPUThreshold())
		}
	}
}

func TestSetMemoryThreshold(t *testing.T) {
	cfg := New()

	testCases := []uint64{0, 1024, 1024 * 1024, 1024 * 1024 * 1024}
	for _, threshold := range testCases {
		cfg.SetMemoryThreshold(threshold)
		if cfg.GetMemoryThreshold() != threshold {
			t.Errorf("Expected MemoryThreshold to be %d, got %d", threshold, cfg.GetMemoryThreshold())
		}
	}
}

func TestSetRefreshRate(t *testing.T) {
	cfg := New()

	testCases := []time.Duration{
		100 * time.Millisecond,
		500 * time.Millisecond,
		time.Second,
		2 * time.Second,
		5 * time.Second,
	}

	for _, rate := range testCases {
		cfg.SetRefreshRate(rate)
		if cfg.GetRefreshRate() != rate {
			t.Errorf("Expected RefreshRate to be %v, got %v", rate, cfg.GetRefreshRate())
		}
	}
}
