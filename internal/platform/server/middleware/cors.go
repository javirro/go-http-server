package middleware

import (
	"net/http"
	"strings"

	"github.com/javier/go-http-server/internal/platform/config"
)

// CORS adds Cross-Origin Resource Sharing headers based on the provided config.
// Preflight OPTIONS requests are handled and short-circuited here.
func CORS(cfg *config.Config) func(http.Handler) http.Handler {
	allowedOrigins := cfg.CORSAllowedOrigins
	allowedMethods := strings.Join(cfg.CORSAllowedMethods, ", ")
	allowedHeaders := strings.Join(cfg.CORSAllowedHeaders, ", ")

	isOriginAllowed := func(origin string) bool {
		for _, allowed := range allowedOrigins {
			if allowed == "*" || strings.EqualFold(allowed, origin) {
				return true
			}
		}
		return false
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" && isOriginAllowed(origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			} else if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
