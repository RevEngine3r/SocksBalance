package main

import (
	"flag"
	"fmt"
	"os"
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
	fmt.Printf("Config: %s\n", *configPath)
	if *listenAddr != "" {
		fmt.Printf("Listen: %s (override)\n", *listenAddr)
	}

	fmt.Println("\n[INFO] Starting initialization...")
	fmt.Println("[TODO] Load configuration")
	fmt.Println("[TODO] Initialize backend pool")
	fmt.Println("[TODO] Start health checker")
	fmt.Println("[TODO] Start TCP proxy server")
	fmt.Println("\n[WARN] Implementation pending. See PROGRESS.md for roadmap.")
}
