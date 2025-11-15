# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**brieftop** is a focused Go CLI tool for monitoring CPU and memory intensive processes. Unlike generic monitoring tools, it specifically filters to show only processes using significant resources (>5% CPU or >50MB memory by default).

Key features:
- Resource filtering with customizable thresholds
- Process hierarchy showing parent/child relationships and threads
- Color-coded TUI based on resource usage levels
- Interactive controls for navigation and process inspection

## Build and Run Commands

```bash
# Install dependencies
go mod tidy

# Build the binary
go build -o brieftop

# Run with default settings (>5% CPU or >50MB memory)
./brieftop

# Run with custom thresholds
./brieftop --cpu 10 --memory 100 --refresh 2s

# Show help
./brieftop --help

# Show version
./brieftop --version
```

## Architecture Overview

The codebase follows a clean layered architecture with clear separation of concerns:

### 1. Entry Point (`main.go`)
- Parses command-line flags (--cpu, --memory, --refresh, --help, --version)
- Initializes the Config, Monitor, and Display components
- Sets up signal handling for graceful shutdown
- Starts the main event loop

### 2. Configuration Layer (`internal/config/`)
- **Purpose**: Centralized configuration management
- **Key Type**: `Config` struct with CPU threshold, memory threshold, refresh rate
- **Pattern**: Simple getter/setter methods for thread-safe configuration access
- **Interface**: `ConfigInterface` allows dependency injection and testing

### 3. Monitoring Layer (`internal/monitor/`)
- **Purpose**: Process data collection, filtering, and hierarchy building
- **Key Types**:
  - `Monitor`: Main monitoring engine
  - `ProcessInfo`: Represents a process with aggregated resource usage
  - `ChildInfo`: Represents child processes or threads
  - `ResourceLevel`: Enum for Low/Medium/High resource usage (used for color coding)

- **Core Algorithm** (`GetFilteredProcesses()`):
  1. **First pass**: Collect all process info and build parent-child mapping
  2. **Second pass**: Build process hierarchies, distinguish threads from child processes
  3. **Resource aggregation**: For processes with children, sum CPU/memory across entire tree
  4. **Third pass**: Filter to top-level processes that meet thresholds
  5. Sort by CPU usage (descending)

- **Thread Detection Heuristic** (`isThread()`):
  - Same executable name as parent
  - Common naming patterns (prefix match)
  - Low memory usage relative to parent (<10%)
  - This is heuristic-based since thread vs. child process distinction is OS-dependent

- **Important**: Parent process stores both aggregated totals (`CPUPercent`, `MemoryBytes`) and original values (`ParentCPU`, `ParentMemory`) for proper display when expanded

### 4. UI Layer (`internal/ui/`)

Three main components work together to render the TUI:

#### `display.go` - Main Display Engine
- **Pattern**: Runs three concurrent goroutines:
  1. `updateLoop()`: Periodically fetches process data from Monitor
  2. `inputLoop()`: Handles keyboard events via tcell
  3. `render()`: Main render loop (50ms refresh) with mutex-protected state

- **State Management**: Uses `sync.RWMutex` to protect shared state (processes list, selected index, pause flag)

- **Rendering Pipeline**:
  - `renderHeader()`: Shows thresholds, pause status, column headers
  - `renderProcesses()`: Renders process tree with expansion logic
  - `renderFooter()`: Displays keyboard controls and process count

- **Hierarchy Display**: When a process is expanded, shows:
  1. Aggregated parent line (sum of all children)
  2. Parent process itself with original values (marked "parent")
  3. Child processes (prefix: `├─`, teal color, marked "child")
  4. Threads (prefix: `╠═`, gray color, marked "thread")

#### `input.go` - Input Handling
- **Controls**:
  - `q/Q` or `Esc` or `Ctrl+C`: Quit
  - `Space`: Toggle pause/unpause
  - `r/R`: Force refresh
  - `↑/↓`: Navigate through processes (wraps around)
  - `Enter`: Expand/collapse selected process
  - `Home/End`: Jump to first/last process

#### `colors.go` - Visual Theme
- **ColorScheme**: Dark navy theme with RGB colors for terminal
- **Resource Level Colors**:
  - Green (<20% CPU, <200MB): Low usage
  - Orange (20-50% CPU, 200-500MB): Medium usage
  - Red (>50% CPU, >500MB): High usage
- **Hierarchy Colors**:
  - White: Parent process (when expanded)
  - Teal: Child processes
  - Gray: Threads
- **Helper Functions**: `GetStatusIcon()` returns expand/collapse arrows (▶/▼) or activity dots (●/○)

## Key Dependencies

- `github.com/shirou/gopsutil/v3`: Cross-platform system and process information
  - Used for: `process.Processes()`, CPU percent, memory info, PPID

- `github.com/gdamore/tcell/v2`: Terminal UI framework
  - Used for: Screen management, event handling, cell-based rendering, colors

## Development Patterns

### Thread Safety
- Display state uses `sync.RWMutex` for concurrent access
- Read locks for rendering, write locks for state updates
- Monitor maintains process state across refreshes to preserve expansion state

### Process Hierarchy Math
The parent + children resource display must sum correctly:
- Top-level line shows: `parent CPU + sum(all children CPU)`
- When expanded, the individual entries must add up to the top-level total
- `ParentCPU` and `ParentMemory` store the original values before aggregation

### Interface Usage
Both `config.Config` and `monitor.Monitor` are accessed via interfaces (`ConfigInterface`) in the UI layer. This allows for:
- Easier testing with mock implementations
- Loose coupling between layers
- Clear API contracts

### Error Handling
- Process collection errors are silently skipped (individual processes may be inaccessible)
- Critical errors (screen init, etc.) are fatal and logged
- Graceful degradation when process info is unavailable

## File Organization

```
brieftop/
├── main.go                      # Entry point, CLI parsing, initialization
├── internal/
│   ├── config/
│   │   └── config.go           # Configuration management
│   ├── monitor/
│   │   ├── process.go          # Process data collection and filtering
│   │   └── system.go           # Utility formatters (bytes, CPU)
│   └── ui/
│       ├── display.go          # Main TUI rendering engine
│       ├── input.go            # Keyboard input handling
│       └── colors.go           # Color scheme and visual helpers
├── go.mod                       # Go module definition
└── README.md                    # User-facing documentation
```

## Testing Considerations

Currently there are no test files. When adding tests:
- Mock `ConfigInterface` for unit testing Monitor and Display
- Test process filtering logic with synthetic process data
- Test thread detection heuristics with various process configurations
- Test hierarchy building with complex parent-child relationships
- Consider table-driven tests for resource level thresholds
