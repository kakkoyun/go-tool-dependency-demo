package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 10 * time.Second
)

type requestCounter struct {
	mu    sync.Mutex
	count int
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	counter := &requestCounter{
		mu:    sync.Mutex{},
		count: 0,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(wrt http.ResponseWriter, req *http.Request) {
		counter.mu.Lock()
		counter.count++
		counter.mu.Unlock()

		_, err := fmt.Fprintf(wrt, "Hello, World! %d", counter.count)
		if err != nil {
			logger.ErrorContext(req.Context(), "Failed to write response", "error", err)
		}
	})

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer stop()

	errgroup, ctx := errgroup.WithContext(ctx)

	logger.InfoContext(ctx, "Starting server", "port", srv.Addr)
	errgroup.Go(func() error {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("failed to listen and serve: %w", err)
		}

		return nil
	})

	errgroup.Go(func() error {
		<-ctx.Done()

		return srv.Shutdown(ctx)
	})

	err := errgroup.Wait()
	if err != nil {
		logger.ErrorContext(ctx, "Fatal error", "error", err)
	}

	logger.InfoContext(ctx, "Server stopped")
}
