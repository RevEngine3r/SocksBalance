package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/RevEngine3r/SocksBalance/internal/config"
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

	fmt.Println("\n[INFO] Starting initialization...")
	fmt.Println("[TODO] Initialize backend pool")
	fmt.Println("[TODO] Start health checker")
	fmt.Println("[TODO] Start TCP proxy server")
	fmt.Println("\n[WARN] Implementation pending. See PROGRESS.md for roadmap.")
}
