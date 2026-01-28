package web

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"
)

// BackendStats represents statistics for a single backend
type BackendStats struct {
	Address     string `json:"address"`
	Name        string `json:"name"`
	Healthy     bool   `json:"healthy"`
	LatencyMs   int64  `json:"latency_ms"`
	LastChecked string `json:"last_checked"`
}

// StatsResponse represents the complete statistics response
type StatsResponse struct {
	Timestamp       string         `json:"timestamp"`
	TotalBackends   int            `json:"total_backends"`
	HealthyBackends int            `json:"healthy_backends"`
	Backends        []BackendStats `json:"backends"`
}

// handleStatsReal implements the actual stats endpoint logic
func (s *Server) handleStatsReal(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers for development
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get all backends from pool
	allBackends := s.pool.GetAll()

	// Convert to stats format
	stats := make([]BackendStats, 0, len(allBackends))
	healthyCount := 0

	for _, b := range allBackends {
		latencyMs := b.Latency().Milliseconds()
		lastChecked := b.LastChecked()
		lastCheckedStr := ""
		if !lastChecked.IsZero() {
			lastCheckedStr = lastChecked.Format(time.RFC3339)
		}

		isHealthy := b.IsHealthy()
		if isHealthy {
			healthyCount++
		}

		stats = append(stats, BackendStats{
			Address:     b.Address(),
			Name:        b.Name(),
			Healthy:     isHealthy,
			LatencyMs:   latencyMs,
			LastChecked: lastCheckedStr,
		})
	}

	// Sort by latency (ascending - fastest first)
	// Unhealthy backends go to the end
	sort.Slice(stats, func(i, j int) bool {
		// Unhealthy backends always go last
		if stats[i].Healthy != stats[j].Healthy {
			return stats[i].Healthy
		}
		// Among healthy or unhealthy, sort by latency
		return stats[i].LatencyMs < stats[j].LatencyMs
	})

	// Build response
	response := StatsResponse{
		Timestamp:       time.Now().Format(time.RFC3339),
		TotalBackends:   len(allBackends),
		HealthyBackends: healthyCount,
		Backends:        stats,
	}

	// Send JSON response
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
