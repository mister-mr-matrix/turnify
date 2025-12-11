package main

import (
	"context"
	"log/slog"
	"os"
	_ "time/tzdata"

	"github.com/mister-mr-matrix/turnify/src/config"
	"github.com/mister-mr-matrix/turnify/src/router"
)

func main() {
	cfg := config.New()

	if cfg.Debug {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	} else {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	}
	slog.Debug("DEBUG logging enabled!")

	rtr := router.New(cfg)
	ctx := context.Background()
	rtr.Start(ctx)
}
