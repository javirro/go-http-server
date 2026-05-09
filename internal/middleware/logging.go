package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture the status code and
// number of bytes written for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	bytesWritten int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

func (rw *responseWriter) statusCode() int {
	if rw.status == 0 {
		return http.StatusOK
	}
	return rw.status
}

// Logger logs each incoming HTTP request using structured logging (slog).
// It records method, path, status code, latency, bytes written, remote addr,
// and the request ID if one is present in the context.
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := wrapResponseWriter(w)

			next.ServeHTTP(wrapped, r)

			status := wrapped.statusCode()
			level := slog.LevelInfo
			if status >= 500 {
				level = slog.LevelError
			} else if status >= 400 {
				level = slog.LevelWarn
			}

			logger.LogAttrs(r.Context(), level, "request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("query", r.URL.RawQuery),
				slog.Int("status", status),
				slog.Duration("latency", time.Since(start)),
				slog.Int("bytes", wrapped.bytesWritten),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", GetRequestID(r.Context())),
			)
		})
	}
}
