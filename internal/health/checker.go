package health

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/RevEngine3r/SocksBalance/internal/backend"
	"golang.org/x/net/proxy"
)

// Checker performs periodic health checks on backends
type Checker struct {
	pool              *backend.Pool
	connectTimeout    time.Duration
	testURL           string
	checkInterval     time.Duration
	requestTimeout    time.Duration
	failureThreshold  int
	mu                sync.Mutex
	running           bool
	cancelFunc        context.CancelFunc
}

// New creates a new health checker
func New(pool *backend.Pool, connectTimeout time.Duration, testURL string, checkInterval time.Duration, requestTimeout time.Duration, failureThreshold int) *Checker {
	return &Checker{
		pool:             pool,
		connectTimeout:   connectTimeout,
		testURL:          testURL,
		checkInterval:    checkInterval,
		requestTimeout:   requestTimeout,
		failureThreshold: failureThreshold,
	}
}

// Start begins periodic health checking
func (c *Checker) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return fmt.Errorf("health checker already running")
	}

	checkerCtx, cancel := context.WithCancel(ctx)
	c.cancelFunc = cancel
	c.running = true
	c.mu.Unlock()

	log.Printf("[INFO] Health checker started (interval: %v)", c.checkInterval)

	// Run initial check immediately
	c.checkAll()

	// Start periodic checks
	go c.runPeriodicChecks(checkerCtx)

	return nil
}

// runPeriodicChecks runs health checks at configured interval
func (c *Checker) runPeriodicChecks(ctx context.Context) {
	ticker := time.NewTicker(c.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.checkAll()
		}
	}
}

// checkAll checks health of all backends concurrently
func (c *Checker) checkAll() {
	backends := c.pool.GetAll()
	if len(backends) == 0 {
		return
	}

	log.Printf("[INFO] Running health checks on %d backend(s)", len(backends))

	var wg sync.WaitGroup
	for _, b := range backends {
		wg.Add(1)
		go func(backend *backend.Backend) {
			defer wg.Done()
			c.checkBackend(backend)
		}(b)
	}

	wg.Wait()

	healthyCount := c.pool.CountHealthy()
	log.Printf("[INFO] Health check complete: %d/%d backends healthy", healthyCount, len(backends))
}

// checkBackend checks a single backend's health
func (c *Checker) checkBackend(b *backend.Backend) {
	address := b.Address()

	// Test 1: Connection test
	if !c.testConnection(address) {
		log.Printf("[WARN] Connection test failed for %s", address)
		b.MarkFailure(c.failureThreshold)
		return
	}

	// Test 2: Latency measurement via URL test
	if c.testURL != "" {
		latency, err := c.measureLatency(address)
		if err != nil {
			log.Printf("[WARN] Latency test failed for %s: %v", address, err)
			b.MarkFailure(c.failureThreshold)
			return
		}

		b.MarkSuccess(latency)
		log.Printf("[INFO] Backend %s healthy (latency: %v)", address, latency)
	} else {
		// No URL test configured, just mark as healthy with 0 latency
		b.MarkSuccess(0)
		log.Printf("[INFO] Backend %s healthy (connection test only)", address)
	}
}

// testConnection tests if backend accepts connections
func (c *Checker) testConnection(address string) bool {
	conn, err := net.DialTimeout("tcp", address, c.connectTimeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// measureLatency measures latency by fetching URL through backend proxy
func (c *Checker) measureLatency(proxyAddr string) (time.Duration, error) {
	// Create SOCKS5 dialer through backend
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		return 0, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
	}

	// Create HTTP client with SOCKS5 proxy
	transport := &http.Transport{
		Dial: dialer.Dial,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   c.requestTimeout,
	}

	// Measure request time
	start := time.Now()
	resp, err := client.Get(c.testURL)
	latency := time.Since(start)

	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return 0, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	return latency, nil
}

// Stop stops the health checker
func (c *Checker) Stop() error {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return fmt.Errorf("health checker not running")
	}
	c.running = false
	c.mu.Unlock()

	log.Printf("[INFO] Stopping health checker...")

	if c.cancelFunc != nil {
		c.cancelFunc()
	}

	log.Printf("[INFO] Health checker stopped")
	return nil
}

// IsRunning returns whether the checker is running
func (c *Checker) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.running
}
