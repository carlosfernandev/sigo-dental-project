package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/carlosfpatino/sigo-dental/internal/identity"
	"github.com/carlosfpatino/sigo-dental/platform/config"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("../deploy/.env"); err != nil {
		slog.Warn("failed to load .env (this is expected in production)", "error", err)
	}

	email := strings.TrimSpace(os.Getenv("SIGO_ADMIN_EMAIL"))
	password := os.Getenv("SIGO_ADMIN_PASSWORD")
	if email == "" || password == "" {
		slog.Error("SIGO_ADMIN_EMAIL and SIGO_ADMIN_PASSWORD are required")
		os.Exit(2)
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	ctx := context.Background()
	db, err := config.InitDB(ctx, cfg)
	if err != nil {
		slog.Error("initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := identity.SeedInitialAdmin(ctx, identity.NewPostgresRepository(db), email, password); err != nil {
		slog.Error("seed initial administrator", "error", err)
		os.Exit(1)
	}
	slog.Info("initial administrator seed completed; existing accounts were left unchanged")
}
