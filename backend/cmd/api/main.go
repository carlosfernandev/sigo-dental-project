package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/carlosfpatino/sigo-dental/internal/billing"
	billingroutes "github.com/carlosfpatino/sigo-dental/internal/billing/routes"
	"github.com/carlosfpatino/sigo-dental/internal/identity"
	"github.com/carlosfpatino/sigo-dental/internal/identity/passwords"
	identityroutes "github.com/carlosfpatino/sigo-dental/internal/identity/routes"
	"github.com/carlosfpatino/sigo-dental/internal/patient"
	patientroutes "github.com/carlosfpatino/sigo-dental/internal/patient/routes"
	"github.com/carlosfpatino/sigo-dental/internal/scheduling"
	schedulingroutes "github.com/carlosfpatino/sigo-dental/internal/scheduling/routes"
	"github.com/carlosfpatino/sigo-dental/middleware"
	"github.com/carlosfpatino/sigo-dental/platform/authn"
	"github.com/carlosfpatino/sigo-dental/platform/config"
	"github.com/carlosfpatino/sigo-dental/platform/httpx"
	"github.com/carlosfpatino/sigo-dental/platform/logging"
	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
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

	dummyHash, err := passwords.NewDummyHash()
	if err != nil {
		logger.Error("failed to initialize authentication timing hash", "error", err)
		os.Exit(1)
	}
	authService := identity.NewAuthService(identity.NewPostgresRepository(db), dummyHash)
	tokens, err := authn.NewTokenManager(cfg.Auth.JWTSecret, authn.AccessTokenTTL)
	if err != nil {
		logger.Error("invalid token configuration", "error", err)
		os.Exit(1)
	}
	activity := authn.NewActivityTracker()

	// Initialize Redis (cache). If it fails, the API continues without cache.
	redisClient := config.InitRedis(ctx, cfg.Redis, logger)
	defer redisClient.Close()

	// TODO: Initialize repositories, services, and handlers here.
	// repo := repository.NewPostgresRepository(db)
	// cacheRepo := cache.NewRedisCache(redisClient)
	// ctrls := controllers.NewControllers(repo, cacheRepo, db, redisClient)

	app := newApp(logger, db, redisHealthPinger{client: redisClient})
	identityroutes.Register(app,
		identity.NewLoginHandler(authService, tokens),
		identity.NewUserAdminHandler(identity.NewUserAdminService(identity.NewPostgresRepository(db))),
		middleware.RequireAuth(tokens, activity),
	)
	patientroutes.Register(app,
		patient.NewHandler(patient.NewService(patient.NewPostgresRepository(db))),
		middleware.RequireAuth(tokens, activity),
	)
	schedulingService := scheduling.NewService(scheduling.NewPostgresRepository(db))
	schedulingroutes.Register(app,
		scheduling.NewHandler(schedulingService),
		middleware.RequireAuth(tokens, activity),
	)
	reminderWorkerCtx, stopReminderWorker := context.WithCancel(context.Background())
	defer stopReminderWorker()
	startReminderWorker(reminderWorkerCtx, logger, schedulingService)
	billingroutes.Register(app,
		billing.NewHandler(billing.NewService(billing.NewPostgresRepository(db), schedulingService)),
		middleware.RequireAuth(tokens, activity),
	)

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
	stopReminderWorker()
	if err := app.Shutdown(); err != nil {
		logger.Error("shutdown error", "error", err)
	}
	logger.Info("shutdown complete")
}

func startReminderWorker(ctx context.Context, logger *slog.Logger, service *scheduling.Service) {
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()
		generate := func() {
			if err := service.GeneratePendingReminders(ctx); err != nil {
				logger.Error("appointment reminder generation failed", "error", err)
			}
		}
		generate()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				generate()
			}
		}
	}()
}

type healthPinger interface {
	Ping(context.Context) error
}

type redisHealthPinger struct {
	client *redis.Client
}

func (p redisHealthPinger) Ping(ctx context.Context) error {
	return p.client.Ping(ctx).Err()
}

func newApp(logger *slog.Logger, postgresHealth, redisHealth healthPinger) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: httpx.FiberErrorHandler(logger),
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

	app.Get("/health", healthHandler(postgresHealth, redisHealth))

	// TODO: Register routes with controllers here
	// routes.SetupRoutes(app, ctrls)

	return app
}

func healthHandler(postgresHealth, redisHealth healthPinger) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()
		if postgresHealth == nil || postgresHealth.Ping(ctx) != nil {
			return httpx.NewError(fiber.StatusServiceUnavailable, "postgres_unavailable", "Base de datos no disponible", nil)
		}

		redisStatus := "ok"
		if redisHealth == nil || redisHealth.Ping(ctx) != nil {
			redisStatus = "degraded"
		}
		status := "ok"
		if redisStatus == "degraded" {
			status = "degraded"
		}
		return c.JSON(fiber.Map{"data": fiber.Map{
			"status":       status,
			"dependencies": fiber.Map{"postgres": "ok", "redis": redisStatus},
		}})
	}
}
