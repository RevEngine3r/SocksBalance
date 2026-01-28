package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/RevEngine3r/SocksBalance/internal/backend"
	"github.com/RevEngine3r/SocksBalance/internal/balancer"
	"github.com/RevEngine3r/SocksBalance/internal/config"
	"github.com/RevEngine3r/SocksBalance/internal/health"
	"github.com/RevEngine3r/SocksBalance/internal/proxy"
)

const version = "0.5.0"

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	listenAddr := flag.String("listen", "", "Override listen address (e.g., 0.0.0.0:1080)")
	mode := flag.String("mode", "", "Proxy mode: transparent (default) or socks5")
	flag.Parse()

	if *showVersion {
		fmt.Printf("SocksBalance v%s\n", version)
		os.Exit(0)
	}

	fmt.Printf("SocksBalance v%s\n", version)

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Override with command-line flags
	if *listenAddr != "" {
		cfg.Listen = *listenAddr
		fmt.Printf("[INFO] Listen address overridden: %s\n", cfg.Listen)
	}
	if *mode != "" {
		cfg.Mode = *mode
		fmt.Printf("[INFO] Mode overridden: %s\n", cfg.Mode)
	}

	// Expand port ranges in backends
	expandedBackends := cfg.ExpandBackends()

	// Display configuration
	fmt.Printf("[INFO] Configuration loaded successfully\n")
	fmt.Printf("  Listen: %s\n", cfg.Listen)
	fmt.Printf("  Mode: %s\n", cfg.Mode)
	fmt.Printf("  Backends (configured): %d\n", len(cfg.Backends))

	// Show original backend configs with range expansion info
	for i, b := range cfg.Backends {
		addrs, _ := config.ParseAddress(b.Address)
		if len(addrs) > 1 {
			// Port range
			if b.Name != "" {
				fmt.Printf("    [%d] %s (%s) → expands to %d backends\n", i+1, b.Name, b.Address, len(addrs))
			} else {
				fmt.Printf("    [%d] %s → expands to %d backends\n", i+1, b.Address, len(addrs))
			}
		} else {
			// Single backend
			if b.Name != "" {
				fmt.Printf("    [%d] %s (%s)\n", i+1, b.Name, b.Address)
			} else {
				fmt.Printf("    [%d] %s\n", i+1, b.Address)
			}
		}
	}

	fmt.Printf("  Backends (total after expansion): %d\n", len(expandedBackends))
	fmt.Printf("  Health Check Interval: %v\n", cfg.Health.CheckInterval)
	if cfg.Health.TestURL != "" {
		fmt.Printf("  Test URL: %s\n", cfg.Health.TestURL)
	}
	fmt.Printf("  Load Balancer: %s\n", cfg.Balancer.Algorithm)
	if cfg.Balancer.MaxLatency > 0 {
		fmt.Printf("  Max Latency Filter: %v (only use backends faster than this)\n", cfg.Balancer.MaxLatency)
	} else {
		fmt.Printf("  Max Latency Filter: disabled (use all healthy backends)\n")
	}
	if cfg.Balancer.StickySessionTTL > 0 {
		fmt.Printf("  Sticky Sessions: %v (same client → same backend)\n", cfg.Balancer.StickySessionTTL)
	} else {
		fmt.Printf("  Sticky Sessions: disabled\n")
	}
	if cfg.Balancer.MaxActiveBackends > 0 {
		fmt.Printf("  Max Active Backends: %d (only use top %d fastest backends)\n", cfg.Balancer.MaxActiveBackends, cfg.Balancer.MaxActiveBackends)
	} else {
		fmt.Printf("  Max Active Backends: unlimited (use all available backends)\n")
	}
	fmt.Printf("  Log Level: %s\n", cfg.Log.Level)

	// Initialize backend pool with expanded backends
	fmt.Println("\n[INFO] Initializing backend pool...")
	pool := backend.NewPool()
	for i, b := range expandedBackends {
		backend := backend.New(b.Address, b.Name)
		pool.Add(backend)
		if i < 5 || cfg.Log.Level == "debug" {
			// Show first 5 or all if debug
			if b.Name != "" {
				fmt.Printf("[INFO] Added backend: %s (%s)\n", b.Address, b.Name)
			} else {
				fmt.Printf("[INFO] Added backend: %s\n", b.Address)
			}
		}
	}
	if len(expandedBackends) > 5 && cfg.Log.Level != "debug" {
		fmt.Printf("[INFO] ... and %d more backends\n", len(expandedBackends)-5)
	}

	// Initialize load balancer with latency filtering, sticky sessions, and max active backends
	fmt.Println("[INFO] Initializing load balancer...")
	bal := balancer.New(pool, cfg.Balancer.MaxLatency, cfg.Balancer.StickySessionTTL, cfg.Balancer.MaxActiveBackends)
	fmt.Printf("[INFO] Load balancer initialized with algorithm: %s\n", cfg.Balancer.Algorithm)
	if cfg.Balancer.MaxLatency > 0 {
		fmt.Printf("[INFO] Only backends with latency ≤ %v will be used\n", cfg.Balancer.MaxLatency)
	}
	if cfg.Balancer.StickySessionTTL > 0 {
		fmt.Printf("[INFO] Sticky sessions enabled: clients stick to same backend for %v\n", cfg.Balancer.StickySessionTTL)
	}
	if cfg.Balancer.MaxActiveBackends > 0 {
		fmt.Printf("[INFO] Anti-detection mode: Only top %d fastest backends will be used concurrently\n", cfg.Balancer.MaxActiveBackends)
	}

	// Start health checker
	fmt.Println("[INFO] Starting health checker...")
	healthChecker := health.New(
		pool,
		cfg.Health.ConnectTimeout,
		cfg.Health.TestURL,
		cfg.Health.CheckInterval,
		cfg.Health.RequestTimeout,
		cfg.Health.FailureThreshold,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := healthChecker.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to start health checker: %v\n", err)
		os.Exit(1)
	}

	// Start proxy server based on mode
	fmt.Printf("[INFO] Starting proxy server on %s...\n", cfg.Listen)

	var serverStopper interface{ Stop() error }

	switch cfg.Mode {
	case "transparent":
		fmt.Println("[INFO] Mode: Transparent TCP forwarding (zero-copy, no SOCKS5 decoding)")
		server := proxy.NewTransparent(cfg.Listen, bal)
		if err := server.Start(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to start server: %v\n", err)
			os.Exit(1)
		}
		serverStopper = server

	case "socks5":
		fmt.Println("[INFO] Mode: SOCKS5 protocol handling (decode/re-encode)")
		server := proxy.New(cfg.Listen, bal)
		if err := server.Start(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to start server: %v\n", err)
			os.Exit(1)
		}
		serverStopper = server

	default:
		fmt.Fprintf(os.Stderr, "[ERROR] Invalid mode: %s (use 'transparent' or 'socks5')\n", cfg.Mode)
		os.Exit(1)
	}

	fmt.Println("[INFO] Server started successfully")
	fmt.Printf("[INFO] Managing %d backends with health monitoring\n", len(expandedBackends))
	if cfg.Balancer.MaxActiveBackends > 0 {
		fmt.Printf("[INFO] GFW Evasion: Rotating through top %d fastest backends only\n", cfg.Balancer.MaxActiveBackends)
	}
	fmt.Println("[INFO] Press Ctrl+C to stop...")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\n[INFO] Shutdown signal received...")

	cancel()

	if err := healthChecker.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] Failed to stop health checker: %v\n", err)
	}

	if err := serverStopper.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to stop server gracefully: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[INFO] Shutdown complete")
}
