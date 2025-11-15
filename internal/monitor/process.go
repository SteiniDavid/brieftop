package monitor

import (
	"fmt"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

type ProcessInfo struct {
	PID           int32
	PPID          int32
	Name          string
	CPUPercent    float64
	MemoryBytes   uint64
	MemoryMB      float64
	Children      []ChildInfo
	Expanded      bool
	LastUpdate    time.Time
	ParentCPU     float64  // Store original parent CPU for display
	ParentMemory  uint64   // Store original parent memory for display
}

type ChildInfo struct {
	PID         int32
	Name        string
	CPUPercent  float64
	MemoryBytes uint64
	IsThread    bool
}

type SystemMetrics struct {
	CPUPercent      float64
	CPUCores        int
	MemoryTotal     uint64
	MemoryUsed      uint64
	MemoryAvailable uint64
	MemoryCached    uint64
	MemoryBuffers   uint64
	MemoryPercent   float64
	SwapTotal       uint64
	SwapUsed        uint64
	SwapPercent     float64
}

type Monitor struct {
	processes    map[int32]*ProcessInfo
	lastCPUTimes map[int32]float64
	config       ConfigInterface
}

type ConfigInterface interface {
	GetCPUThreshold() float64
	GetMemoryThreshold() uint64
	GetRefreshRate() time.Duration
}

func New(config ConfigInterface) *Monitor {
	return &Monitor{
		processes:    make(map[int32]*ProcessInfo),
		lastCPUTimes: make(map[int32]float64),
		config:       config,
	}
}

func (m *Monitor) GetFilteredProcesses() ([]*ProcessInfo, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %w", err)
	}

	var filtered []*ProcessInfo
	allProcesses := make(map[int32]*ProcessInfo)
	qualifyingProcesses := make(map[int32]*ProcessInfo)
	childrenMap := make(map[int32][]int32) // parent PID -> children PIDs

	// First pass: collect all process info
	for _, p := range processes {
		info, err := m.getProcessInfo(p)
		if err != nil {
			continue
		}
		allProcesses[info.PID] = info
		
		// Build parent-child mapping
		if info.PPID != 0 {
			childrenMap[info.PPID] = append(childrenMap[info.PPID], info.PID)
		}
		
		// Check if process meets our thresholds
		if info.CPUPercent >= m.config.GetCPUThreshold() || info.MemoryBytes >= m.config.GetMemoryThreshold() {
			qualifyingProcesses[info.PID] = info
		}
	}

	// Second pass: build hierarchies and collect children
	for pid, info := range allProcesses {
		if childPIDs, hasChildren := childrenMap[pid]; hasChildren {
			for _, childPID := range childPIDs {
				if childInfo, exists := allProcesses[childPID]; exists {
					// Determine if this is a thread or child process
					isThread := m.isThread(childInfo, info)
					
					child := ChildInfo{
						PID:         childInfo.PID,
						Name:        childInfo.Name,
						CPUPercent:  childInfo.CPUPercent,
						MemoryBytes: childInfo.MemoryBytes,
						IsThread:    isThread,
					}
					info.Children = append(info.Children, child)
					
					// If child qualifies, ensure parent is included
					if _, childQualifies := qualifyingProcesses[childPID]; childQualifies {
						if _, parentQualifies := qualifyingProcesses[pid]; !parentQualifies {
							qualifyingProcesses[pid] = info
						}
					}
				}
			}
		}
	}

	// Aggregate resources for processes with children
	for _, info := range qualifyingProcesses {
		if len(info.Children) > 0 {
			// Store original parent values
			info.ParentCPU = info.CPUPercent
			info.ParentMemory = info.MemoryBytes
			
			// Sum all children CPU and memory
			totalCPU := 0.0
			totalMemory := uint64(0)
			
			for _, child := range info.Children {
				totalCPU += child.CPUPercent
				totalMemory += child.MemoryBytes
			}
			
			// Add parent's own usage to totals
			totalCPU += info.ParentCPU
			totalMemory += info.ParentMemory
			
			// Set the totals as the displayed values
			info.CPUPercent = totalCPU
			info.MemoryBytes = totalMemory
			info.MemoryMB = float64(totalMemory) / (1024 * 1024)
		}
	}

	// Third pass: collect top-level processes that qualify
	for _, info := range qualifyingProcesses {
		// Only include processes that don't have a qualifying parent
		if _, parentExists := qualifyingProcesses[info.PPID]; !parentExists {
			// Ensure MemoryMB is set for processes without children
			if len(info.Children) == 0 {
				info.MemoryMB = float64(info.MemoryBytes) / (1024 * 1024)
			}
			filtered = append(filtered, info)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CPUPercent > filtered[j].CPUPercent
	})

	return filtered, nil
}

func (m *Monitor) getProcessInfo(p *process.Process) (*ProcessInfo, error) {
	pid := p.Pid
	
	name, err := p.Name()
	if err != nil {
		return nil, err
	}

	ppid, err := p.Ppid()
	if err != nil {
		ppid = 0
	}

	cpuPercent, err := p.CPUPercent()
	if err != nil {
		cpuPercent = 0
	}

	memInfo, err := p.MemoryInfo()
	if err != nil {
		return nil, err
	}

	info := &ProcessInfo{
		PID:         pid,
		PPID:        ppid,
		Name:        name,
		CPUPercent:  cpuPercent,
		MemoryBytes: memInfo.RSS,
		LastUpdate:  time.Now(),
		Expanded:    false,
		Children:    make([]ChildInfo, 0),
	}

	if existing, exists := m.processes[pid]; exists {
		info.Expanded = existing.Expanded
	}

	m.processes[pid] = info
	return info, nil
}

// isThread determines if a process is likely a thread vs a child process
// This is a heuristic since the distinction can be OS-dependent
func (m *Monitor) isThread(child, parent *ProcessInfo) bool {
	// Heuristics for identifying threads:
	// 1. Same executable name as parent
	// 2. Low memory usage relative to parent (threads share memory)
	// 3. Certain naming patterns
	
	if child.Name == parent.Name {
		return true
	}
	
	// Check for common thread naming patterns
	if len(child.Name) > len(parent.Name) && 
	   child.Name[:len(parent.Name)] == parent.Name {
		return true
	}
	
	// If child uses significantly less memory, likely a thread
	if parent.MemoryBytes > 0 && 
	   float64(child.MemoryBytes)/float64(parent.MemoryBytes) < 0.1 {
		return true
	}
	
	return false
}

func (m *Monitor) ToggleExpanded(pid int32) {
	if info, exists := m.processes[pid]; exists {
		info.Expanded = !info.Expanded
	}
}

func (m *Monitor) GetResourceLevel(cpuPercent float64, memoryMB float64) ResourceLevel {
	if cpuPercent >= 50 || memoryMB >= 500 {
		return High
	} else if cpuPercent >= 20 || memoryMB >= 200 {
		return Medium
	}
	return Low
}

type ResourceLevel int

const (
	Low ResourceLevel = iota
	Medium
	High
)

func (rl ResourceLevel) String() string {
	switch rl {
	case Low:
		return "Low"
	case Medium:
		return "Medium"
	case High:
		return "High"
	default:
		return "Unknown"
	}
}

func (m *Monitor) GetSystemMetrics() (*SystemMetrics, error) {
	metrics := &SystemMetrics{}

	// Get CPU metrics
	cpuPercentages, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercentages) > 0 {
		metrics.CPUPercent = cpuPercentages[0]
	}

	// Get CPU core count
	cpuCounts, err := cpu.Counts(true) // true for logical cores
	if err == nil {
		metrics.CPUCores = cpuCounts
	}

	// Get memory metrics
	vmem, err := mem.VirtualMemory()
	if err == nil {
		metrics.MemoryTotal = vmem.Total
		metrics.MemoryUsed = vmem.Used
		metrics.MemoryAvailable = vmem.Available
		metrics.MemoryCached = vmem.Cached
		metrics.MemoryBuffers = vmem.Buffers
		metrics.MemoryPercent = vmem.UsedPercent
	}

	// Get swap metrics
	swap, err := mem.SwapMemory()
	if err == nil {
		metrics.SwapTotal = swap.Total
		metrics.SwapUsed = swap.Used
		metrics.SwapPercent = swap.UsedPercent
	}

	return metrics, nil
}