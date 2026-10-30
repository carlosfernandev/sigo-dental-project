package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/carlosfpatino/sigo-dental/platform/config"
	"github.com/carlosfpatino/sigo-dental/platform/migrate"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("../deploy/.env"); err != nil {
		slog.Warn("failed to load .env (this is expected in production)", "error", err)
	}
	if len(os.Args) != 2 || os.Args[1] != "up" {
		slog.Error("usage: go run ./cmd/migrate up")
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

	if err := migrate.Up(ctx, db, os.DirFS("migrations")); err != nil {
		slog.Error("apply migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations applied")
}
