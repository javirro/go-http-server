package middleware_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/javier/go-http-server/internal/middleware"
)

func nopHandler(status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	})
}

func TestRequestID_Generated(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	middleware.RequestID(nopHandler(http.StatusOK)).ServeHTTP(rec, req)

	if got := rec.Header().Get(middleware.RequestIDHeader); got == "" {
		t.Error("expected X-Request-ID header to be set")
	}
}

func TestRequestID_Propagated(t *testing.T) {
	const id = "test-request-id-123"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(middleware.RequestIDHeader, id)
	rec := httptest.NewRecorder()

	middleware.RequestID(nopHandler(http.StatusOK)).ServeHTTP(rec, req)

	if got := rec.Header().Get(middleware.RequestIDHeader); got != id {
		t.Errorf("X-Request-ID = %q, want %q", got, id)
	}
}

func TestRecovery_CatchesPanic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	panic := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("intentional panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	middleware.Recovery(logger)(panic).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

func TestSecureHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	middleware.SecureHeaders(nopHandler(http.StatusOK)).ServeHTTP(rec, req)

	expected := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}
	for header, want := range expected {
		if got := rec.Header().Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
}

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := middleware.NewRateLimiter(100, 10)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	rl.Middleware(nopHandler(http.StatusOK)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	// Burst of 1, after that every request should be rejected.
	rl := middleware.NewRateLimiter(0.001, 1)
	handler := rl.Middleware(nopHandler(http.StatusOK))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.1:1234"

	// First request should pass (burst = 1).
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Second request should be rate-limited.
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req)

	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", rec2.Code)
	}
}

func TestChain_Order(t *testing.T) {
	var order []int

	mw := func(n int) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, n)
				next.ServeHTTP(w, r)
			})
		}
	}

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, 99)
	})

	chain := middleware.Chain(mw(1), mw(2), mw(3))(base)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	chain.ServeHTTP(httptest.NewRecorder(), req)

	want := []int{1, 2, 3, 99}
	for i, v := range want {
		if order[i] != v {
			t.Errorf("order[%d] = %d, want %d", i, order[i], v)
		}
	}
}
