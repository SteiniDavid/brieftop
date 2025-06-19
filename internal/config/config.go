package config

import "time"

type Config struct {
	CPUThreshold    float64
	MemoryThreshold uint64
	RefreshRate     time.Duration
	ShowThreads     bool
}

func New() *Config {
	return &Config{
		CPUThreshold:    5.0,           // 5% CPU
		MemoryThreshold: 50 * 1024 * 1024, // 50MB in bytes
		RefreshRate:     time.Second,
		ShowThreads:     true,
	}
}

func (c *Config) SetCPUThreshold(threshold float64) {
	c.CPUThreshold = threshold
}

func (c *Config) SetMemoryThreshold(threshold uint64) {
	c.MemoryThreshold = threshold
}

func (c *Config) SetRefreshRate(rate time.Duration) {
	c.RefreshRate = rate
}

func (c *Config) GetCPUThreshold() float64 {
	return c.CPUThreshold
}

func (c *Config) GetMemoryThreshold() uint64 {
	return c.MemoryThreshold
}

func (c *Config) GetRefreshRate() time.Duration {
	return c.RefreshRate
}