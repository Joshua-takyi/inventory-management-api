package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/joshua-takyi/stockflow-api/container"
	"github.com/joshua-takyi/stockflow-api/internal/adapters/routes"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		slog.Error("loading env", "error", err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	ctn, err := container.NewContainer(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "error loading container", "error", err)
		os.Exit(1)
	}
	if err := <-run(ctx, *ctn); err != nil {
		slog.ErrorContext(ctx, "server initialization failed", "error", err)
	}
}

func run(ctx context.Context, ctn container.Container) <-chan error {
	errChan := make(chan error, 1)

	server := &http.Server{
		Addr:    ":" + ctn.Config.Port,
		Handler: routes.Router(&ctn),
	}
	go func() {
		defer close(errChan)
		serverErr := make(chan error, 1)
		go func() {
			serverErr <- server.ListenAndServe()
		}()
		select {
		case err := <-serverErr:
			if !errors.Is(err, http.ErrServerClosed) {
				errChan <- err
			}
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := server.Shutdown(shutdownCtx); err != nil {
				errChan <- err
			}
		}
	}()
	return errChan
}
