# brieftop

[![Go Reference](https://pkg.go.dev/badge/github.com/SteiniDavid/brieftop.svg)](https://pkg.go.dev/github.com/SteiniDavid/brieftop)
[![Go Report Card](https://goreportcard.com/badge/github.com/SteiniDavid/brieftop)](https://goreportcard.com/report/github.com/SteiniDavid/brieftop)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A focused Go CLI tool for monitoring CPU and memory intensive processes - showing only the essentials. Unlike generic tools like htop, brieftop specifically shows only processes that are using significant resources.

## Features

- **Resource Filtering**: Only displays processes using >5% CPU or >50MB memory
- **Process Hierarchy**: Groups child processes and threads under their parent processes with visual distinction
- **Color Coding**: Visual indicators based on resource usage levels
  - 🟢 Green: Low usage (CPU <20%, Memory <200MB)
  - 🟡 Yellow: Medium usage (CPU 20-50%, Memory 200-500MB)  
  - 🔴 Red: High usage (CPU >50%, Memory >500MB)
- **Interactive Controls**:
  - `↑/↓`: Navigate through processes
  - `Enter`: Expand/collapse thread details
  - `Space`: Pause/unpause updates
  - `R`: Force refresh
  - `Q`: Quit application

## Installation

### Quick Install (Recommended)

If you have Go 1.21 or later installed:

```bash
go install github.com/SteiniDavid/brieftop@latest
```

This installs `brieftop` to your `$GOPATH/bin` directory. Make sure `$GOPATH/bin` is in your `PATH`.

### Build from Source

```bash
git clone https://github.com/SteiniDavid/brieftop.git
cd brieftop
go mod tidy
go build -o brieftop
```

### Run
```bash
# Run with default settings (>5% CPU or >50MB memory)
./brieftop

# Show help
./brieftop --help

# Custom thresholds and refresh rate
./brieftop --cpu 10 --memory 100 --refresh 2s

# Show version
./brieftop --version
```

## Configuration

### Command Line Options
- `--cpu <float>`: CPU threshold percentage (default: 5.0)
- `--memory <uint>`: Memory threshold in MB (default: 50)  
- `--refresh <duration>`: Refresh rate, e.g. "500ms", "2s" (default: 1s)
- `--help`: Show help information
- `--version`: Show version information

### Default Values
- **CPU Threshold**: 5% per core
- **Memory Threshold**: 50MB
- **Refresh Rate**: 1 second

All thresholds can be customized via command line flags.

## Usage

The interface displays:
1. **Header**: Shows current thresholds and pause status
2. **Process List**: Filtered processes with expandable thread details
3. **Footer**: Keyboard controls reference

### Process Display Format
```
▶ PID        CPU     MEMORY CHILDREN NAME (expands to fill available space)
▼ 1234      35.4%   490.6MB       12 chrome (total of parent + all children)
  ├─ 1234    3.2%    85.4MB          chrome (parent)
  ├─ 1235    8.1%   128.4MB          chrome-renderer (child process)
  ╠═ 1236    2.3%    45.2MB          chrome-gpu-process (thread)
  ├─ 1237    4.8%    71.7MB          chrome-utility-process (child process) 
  ╠═ 1238    1.2%     8.1MB          chrome-background-thread (thread)
  ... (sum of all entries = 35.4% total)
```

**Visual Indicators:**
- `├─` White: Parent process (when expanded)
- `├─` Teal: Child processes (separate processes)
- `╠═` Gray: Threads (shared memory space)

**Resource Aggregation:**
- **Top Level**: Shows sum of parent + all children/threads
- **When Expanded**: Parent process listed first, followed by all children
- **Perfect Math**: All individual entries sum to the top-level total

## Technical Details

### Architecture
- `main.go`: Application entry point
- `internal/config/`: Configuration management
- `internal/monitor/`: Process monitoring and filtering logic
- `internal/ui/`: Terminal user interface components

### Dependencies
- `github.com/shirou/gopsutil/v3`: Cross-platform system information
- `github.com/gdamore/tcell/v2`: Terminal UI framework

### Performance
- Efficient process filtering reduces overhead
- Real-time updates with configurable refresh rates
- Memory-efficient data structures

## Contributing

This is a focused tool designed specifically for monitoring resource-intensive processes. Feature requests should align with this core purpose.

## License

MIT License - see [LICENSE](LICENSE) file for details.