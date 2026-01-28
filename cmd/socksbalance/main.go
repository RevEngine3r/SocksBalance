package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/RevEngine3r/SocksBalance/internal/backend"
	"github.com/RevEngine3r/SocksBalance/internal/config"
	"github.com/RevEngine3r/SocksBalance/internal/proxy"
)

const version = "0.1.0"

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	listenAddr := flag.String("listen", "", "Override listen address (e.g., 0.0.0.0:1080)")
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

	if *listenAddr != "" {
		cfg.Listen = *listenAddr
		fmt.Printf("[INFO] Listen address overridden: %s\n", cfg.Listen)
	}

	fmt.Printf("[INFO] Configuration loaded successfully\n")
	fmt.Printf("  Listen: %s\n", cfg.Listen)
	fmt.Printf("  Backends: %d\n", len(cfg.Backends))
	for i, b := range cfg.Backends {
		if b.Name != "" {
			fmt.Printf("    [%d] %s (%s)\n", i+1, b.Name, b.Address)
		} else {
			fmt.Printf("    [%d] %s\n", i+1, b.Address)
		}
	}
	fmt.Printf("  Health Check Interval: %v\n", cfg.Health.CheckInterval)
	fmt.Printf("  Load Balancer: %s\n", cfg.Balancer.Algorithm)
	fmt.Printf("  Log Level: %s\n", cfg.Log.Level)

	fmt.Println("\n[INFO] Initializing backend pool...")
	pool := backend.NewPool()
	for _, b := range cfg.Backends {
		backend := backend.New(b.Address, b.Name)
		pool.Add(backend)
		fmt.Printf("[INFO] Added backend: %s\n", b.Address)
	}

	fmt.Println("[TODO] Start health checker")

	fmt.Printf("[INFO] Starting TCP proxy server on %s...\n", cfg.Listen)
	server := proxy.New(cfg.Listen, pool)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := server.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to start server: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[INFO] Server started successfully")
	fmt.Println("[INFO] Press Ctrl+C to stop...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\n[INFO] Shutdown signal received...")

	cancel()

	if err := server.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to stop server gracefully: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[INFO] Shutdown complete")
}
