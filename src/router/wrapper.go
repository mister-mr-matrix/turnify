package router

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/mister-mr-matrix/turnify/src/config"
)

type RouterWrapper struct {
	mux  *chi.Mux
	port int
}

func New(cfg config.Config) RouterWrapper {
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	mux.Use(middleware.StripSlashes)

	mux.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(cfg, w, r)
	})

	return RouterWrapper{mux: mux, port: cfg.TurnifyPort}
}

func (rw RouterWrapper) Start(ctx context.Context) {
	srv := http.Server{
		Addr:    ":" + strconv.Itoa(rw.port),
		Handler: rw.mux,
	}

	slog.Info("Starting server", "port", rw.port)

	go func() {
		<-ctx.Done()
		slog.Info("Shutting down...")

		timeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(timeout)
		if err != nil {
			slog.Error("Server shut down failed", "err", err)
		} else {
			slog.Info("Server shut down")
		}
	}()

	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		slog.Error("Failed to start", "err", err)
	}
}
