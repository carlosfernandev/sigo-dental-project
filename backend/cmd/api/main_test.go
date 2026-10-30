package main

import (
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestHealth(t *testing.T) {
	app := newApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	if got, want := string(body), `{"status":"ok"}`; got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}
