package handler

import (
	"net/http"
	"runtime"
	"time"
)

var startTime = time.Now()

// HealthResponse is the payload returned by the liveness endpoint.
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Uptime    string `json:"uptime"`
}

// Health handles GET /health — used by load balancers for liveness checks.
// Always returns 200 as long as the process is running.
func Health(w http.ResponseWriter, r *http.Request) {
	JSON(w, r, http.StatusOK, Envelope[HealthResponse]{
		Data: HealthResponse{
			Status:    "ok",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Uptime:    time.Since(startTime).Round(time.Second).String(),
		},
	})
}

// ReadyResponse is the payload returned by the readiness endpoint.
type ReadyResponse struct {
	Status   string         `json:"status"`
	Checks   map[string]any `json:"checks"`
	GoVersion string        `json:"go_version"`
}

// Ready handles GET /ready — used by orchestrators (k8s readiness probe).
// Returns 503 if the service is not ready to handle traffic.
func Ready(w http.ResponseWriter, r *http.Request) {
	checks := map[string]any{
		"goroutines": runtime.NumGoroutine(),
	}

	JSON(w, r, http.StatusOK, Envelope[ReadyResponse]{
		Data: ReadyResponse{
			Status:    "ready",
			Checks:    checks,
			GoVersion: runtime.Version(),
		},
	})
}
