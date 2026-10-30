// Package logging builds the structured logger shared by every service.
//
// The logger writes to stdout and never to a file: in a container the runtime
// collects stdout, and a service that writes its own log file loses that log
// when the container is replaced.
package logging

import (
	"io"
	"log/slog"

	"github.com/carlosfpatino/sigo-dental/platform/config"
)

// New returns a logger tagged with the service name and environment.
//
// Development uses a text handler because a human reads it. Every other
// environment uses JSON because a log shipper reads it.
func New(w io.Writer, cfg config.Config, service string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}

	var handler slog.Handler
	if cfg.Env == config.EnvDevelopment {
		handler = slog.NewTextHandler(w, opts)
	} else {
		handler = slog.NewJSONHandler(w, opts)
	}

	return slog.New(handler).With(
		slog.String("service", service),
		slog.String("env", cfg.Env),
	)
}
