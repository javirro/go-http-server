package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter holds per-IP rate limiters and cleans them up periodically.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rps      float64
	burst    int
}

// NewRateLimiter creates a RateLimiter and starts a background goroutine to
// evict stale entries every minute.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rps:      rps,
		burst:    burst,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{limiter: rate.NewLimiter(rate.Limit(rl.rps), rl.burst)}
		rl.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter
}

// cleanup removes visitors that haven't been seen in the last 3 minutes.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware returns an http.Handler middleware that enforces per-IP rate
// limiting. Clients exceeding the limit receive 429 Too Many Requests.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			w.Header().Set("Retry-After", "1")
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// clientIP extracts the real client IP, honouring common proxy headers.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}

	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		// Take only the first (leftmost) address.
		if comma := strings.IndexByte(fwd, ','); comma >= 0 {
			fwd = fwd[:comma]
		}
		return strings.TrimSpace(fwd)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
