package main

import (
	"brieftop/internal/config"
	"brieftop/internal/monitor"
	"brieftop/internal/ui"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Command line flags
	var (
		cpuThreshold    = flag.Float64("cpu", 5.0, "CPU threshold percentage (processes using more than this will be shown)")
		memoryThreshold = flag.Uint64("memory", 50, "Memory threshold in MB (processes using more than this will be shown)")
		refreshRate     = flag.Duration("refresh", time.Second, "Refresh rate (e.g., 500ms, 2s)")
		showHelp        = flag.Bool("help", false, "Show help information")
		showVersion     = flag.Bool("version", false, "Show version information")
	)
	
	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "brieftop - A focused process monitoring tool showing only the essentials\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nControls:\n")
		fmt.Fprintf(os.Stderr, "  ↑/↓       Navigate through processes\n")
		fmt.Fprintf(os.Stderr, "  Enter     Expand/collapse process details\n")
		fmt.Fprintf(os.Stderr, "  Space     Pause/unpause updates\n")
		fmt.Fprintf(os.Stderr, "  R         Force refresh\n")
		fmt.Fprintf(os.Stderr, "  Q         Quit application\n")
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s --cpu 10 --memory 100 --refresh 2s\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nThis will show processes using >10%% CPU or >100MB memory, refreshing every 2 seconds.\n")
	}
	
	flag.Parse()
	
	// Handle help and version flags
	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}
	
	if *showVersion {
		fmt.Println("brieftop v1.0.0")
		fmt.Println("A focused process monitoring tool showing only the essentials")
		os.Exit(0)
	}
	
	// Create config with command line values
	cfg := config.New()
	cfg.SetCPUThreshold(*cpuThreshold)
	cfg.SetMemoryThreshold(*memoryThreshold * 1024 * 1024) // Convert MB to bytes
	cfg.SetRefreshRate(*refreshRate)
	
	mon := monitor.New(cfg)
	
	display := ui.New(cfg, mon)
	
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		display.Stop()
		os.Exit(0)
	}()
	
	if err := display.Run(); err != nil {
		log.Fatalf("Failed to run display: %v", err)
	}
}