package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/carlosfpatino/sigo-dental/platform/config"
	"github.com/carlosfpatino/sigo-dental/platform/logging"
	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env from project root during development.
	// In production, environment variables are set by the container runtime.
	if err := godotenv.Load("../deploy/.env"); err != nil {
		slog.Warn("failed to load .env (this is expected in production)", "error", err)
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	logger := logging.New(os.Stdout, cfg, "sigo-api")
	logger.Info("starting sigo-dental monolith", "env", cfg.Env)

	// Initialize database connection
	ctx := context.Background()
	db, err := config.InitDB(ctx, cfg)
	if err != nil {
		logger.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connected", "database", cfg.Postgres.Database)

	// Initialize Redis (cache). If it fails, the API continues without cache.
	redisClient := config.InitRedis(ctx, cfg.Redis, logger)
	defer redisClient.Close()

	// TODO: Initialize repositories, services, and handlers here.
	// repo := repository.NewPostgresRepository(db)
	// cacheRepo := cache.NewRedisCache(redisClient)
	// ctrls := controllers.NewControllers(repo, cacheRepo, db, redisClient)

	app := newApp(logger)

	// Start server in a goroutine
	go func() {
		addr := ":8080"
		logger.Info("http server listening", "addr", addr)
		if err := app.Listen(addr); err != nil {
			logger.Error("http server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down http server")
	if err := app.Shutdown(); err != nil {
		logger.Error("shutdown error", "error", err)
	}
	logger.Info("shutdown complete")
}

func newApp(logger *slog.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			stableCode := "internal_error"
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				stableCode = "http_error"
			}
			logger.Error("request error", "error", err, "path", c.Path(), "status", code)
			return c.Status(code).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    stableCode,
					"message": http.StatusText(code),
					"details": nil,
				},
			})
		},
	})

	// TODO: Add CORS middleware here if needed
	// corsOrigins := os.Getenv("CORS_ORIGINS")
	// if corsOrigins == "" {
	// 	corsOrigins = "*"
	// }
	// app.Use(cors.New(cors.Config{
	// 	AllowOrigins: corsOrigins,
	// 	AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	// }))

	// Health check endpoint
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// TODO: Register routes with controllers here
	// routes.SetupRoutes(app, ctrls)

	return app
}
