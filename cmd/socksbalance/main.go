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

const version = "0.2.0"

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

	// Display configuration
	fmt.Printf("[INFO] Configuration loaded successfully\n")
	fmt.Printf("  Listen: %s\n", cfg.Listen)
	fmt.Printf("  Mode: %s\n", cfg.Mode)
	fmt.Printf("  Backends: %d\n", len(cfg.Backends))
	for i, b := range cfg.Backends {
		if b.Name != "" {
			fmt.Printf("    [%d] %s (%s)\n", i+1, b.Name, b.Address)
		} else {
			fmt.Printf("    [%d] %s\n", i+1, b.Address)
		}
	}
	fmt.Printf("  Health Check Interval: %v\n", cfg.Health.CheckInterval)
	if cfg.Health.TestURL != "" {
		fmt.Printf("  Test URL: %s\n", cfg.Health.TestURL)
	}
	fmt.Printf("  Load Balancer: %s\n", cfg.Balancer.Algorithm)
	fmt.Printf("  Log Level: %s\n", cfg.Log.Level)

	// Initialize backend pool
	fmt.Println("\n[INFO] Initializing backend pool...")
	pool := backend.NewPool()
	for _, b := range cfg.Backends {
		backend := backend.New(b.Address, b.Name)
		pool.Add(backend)
		fmt.Printf("[INFO] Added backend: %s\n", b.Address)
	}

	// Initialize load balancer
	fmt.Println("[INFO] Initializing load balancer...")
	bal := balancer.New(pool)
	fmt.Printf("[INFO] Load balancer initialized with algorithm: %s\n", cfg.Balancer.Algorithm)

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
