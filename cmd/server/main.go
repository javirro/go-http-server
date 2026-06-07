package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/javier/go-http-server/internal/platform/config"
	"github.com/javier/go-http-server/internal/platform/database"
	"github.com/javier/go-http-server/internal/platform/server/httpserver"
	"github.com/javier/go-http-server/internal/platform/server/middleware"
	"github.com/javier/go-http-server/internal/platform/server/routes"
	"github.com/javier/go-http-server/internal/products"
)

// In Go, os.Exit terminates the process immediately — it does not run any defer statements.
// If you put logic directly in main() and called os.Exit,
// any defer calls you had (to close files, flush logs, etc.) would be silently skipped.
func main() {
	os.Exit(run())
}

// run is separated from main so deferred calls execute before os.Exit.
// This is a common pattern in Go to ensure that deferred calls are executed
// before the process exits. Run returns 0 if successful, 1 if there was an error.
func run() int {
	// Load configuration from environment variables.
	cfg, err := config.Load()
	if err != nil {
		// Logger isn't set up yet — use stdlib default.
		slog.Error("failed to load configuration", slog.Any("error", err))
		return 1
	}

	// Build the logger.
	logger := buildLogger(cfg)

	// Connect to the database and verify connectivity before serving traffic.
	pool, err := database.NewPool(context.Background(), cfg)
	if err != nil {
		logger.Error("failed to connect to database", slog.Any("error", err))
		return 1
	}
	defer pool.Close()
	logger.Info("connected to database")

	// Apply database migrations.
	if err := database.Migrate(context.Background(), pool); err != nil {
		logger.Error("failed to run migrations", slog.Any("error", err))
		return 1
	}

	// Build the product repository (PostgreSQL) and seed it on first boot.
	productRepo := products.NewPostgresRepository(pool)
	if err := productRepo.Seed(context.Background()); err != nil {
		logger.Error("failed to seed products", slog.Any("error", err))
		return 1
	}

	// Build the middleware chain (outermost first). This is a chain of middleware functions
	// that will be applied to the request in the order they are defined.
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)

	chain := middleware.Chain(
		middleware.RequestID,
		middleware.Logger(logger),
		middleware.Recovery(logger),
		middleware.SecureHeaders,
		middleware.CORS(cfg),
		rateLimiter.Middleware,
	)

	// Build the router.
	mux := routes.NewRouter(productRepo)

	// Build the server.
	srv := httpserver.New(chain(mux), cfg, logger)

	// Listen for OS termination signals.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start the server.
	if err := srv.Start(ctx); err != nil {
		logger.Error("server error", slog.Any("error", err))
		return 1
	}

	return 0
}

func buildLogger(cfg *config.Config) *slog.Logger {
	var level slog.Level
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.IsDevelopment(),
	}

	var handler slog.Handler
	if cfg.LogFormat == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
