package middleware

import "net/http"

// SecureHeaders adds security-related HTTP response headers following OWASP
// recommendations.
func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()

		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("X-XSS-Protection", "0") // Modern browsers use CSP instead.
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Content-Security-Policy", "default-src 'self'")
		h.Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")

		// Only send HSTS over HTTPS.
		if r.TLS != nil {
			h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		}

		next.ServeHTTP(w, r)
	})
}
