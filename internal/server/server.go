package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/javier/go-http-server/internal/config"
)

// Server wraps net/http.Server with lifecycle management.
type Server struct {
	httpServer *http.Server
	cfg        *config.Config
	logger     *slog.Logger
}

// New creates a configured Server ready to be started.
func New(handler http.Handler, cfg *config.Config, logger *slog.Logger) *Server {
	httpServer := &http.Server{
		Addr:    cfg.Addr(),
		Handler: handler,

		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,

		// Limit request header size (default 1 MB is too generous).
		MaxHeaderBytes: 1 << 20,
	}

	if cfg.TLSEnabled() {
		httpServer.TLSConfig = defaultTLSConfig()
	}

	return &Server{
		httpServer: httpServer,
		cfg:        cfg,
		logger:     logger,
	}
}

// Start begins listening and blocks until the context is cancelled, at which
// point it performs a graceful shutdown.
func (s *Server) Start(ctx context.Context) error {
	shutdownErr := make(chan error, 1)

	go func() {
		<-ctx.Done()
		s.logger.Info("shutdown signal received, draining connections",
			slog.Duration("timeout", s.cfg.ShutdownTimeout),
		)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()

		shutdownErr <- s.httpServer.Shutdown(shutdownCtx)
	}()

	s.logger.Info("server starting",
		slog.String("addr", s.cfg.Addr()),
		slog.Bool("tls", s.cfg.TLSEnabled()),
		slog.String("env", s.cfg.Env),
	)

	var listenErr error
	if s.cfg.TLSEnabled() {
		listenErr = s.httpServer.ListenAndServeTLS(s.cfg.TLSCertFile, s.cfg.TLSKeyFile)
	} else {
		listenErr = s.httpServer.ListenAndServe()
	}

	if !errors.Is(listenErr, http.ErrServerClosed) {
		return fmt.Errorf("listen: %w", listenErr)
	}

	if err := <-shutdownErr; err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	s.logger.Info("server stopped gracefully")
	return nil
}

// ListenAddr returns the actual listening address, useful when using port 0
// in tests. It must be called after the server has started.
func ListenAddr(handler http.Handler, cfg *config.Config) (net.Listener, *http.Server, error) {
	ln, err := net.Listen("tcp", cfg.Addr())
	if err != nil {
		return nil, nil, fmt.Errorf("listen: %w", err)
	}

	srv := &http.Server{
		Handler:           handler,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    1 << 20,
	}

	return ln, srv, nil
}

// defaultTLSConfig returns a TLS configuration following modern security
// recommendations (TLS 1.2+, strong cipher suites).
func defaultTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}
}
