package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/go-kit/log/level"
	"golang.org/x/sync/errgroup"

	"github.com/kakkoyun/go-tool-dependency-demo/pkg/logger"
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

type Flags struct {
	LogLevel string        `default:"info" enum:"error,warn,info,debug" help:"log level."`
	Address  string        `default:":8080" help:"Address string for internal server"`
	Interval time.Duration `default:"1s" help:"Interval between each shell execution"`
}

func main() {
	flags := &Flags{}
	_ = kong.Parse(flags)

	logger := logger.NewLogger(flags.LogLevel, logger.LogFormatLogfmt, "gotools")

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
			level.Error(logger).Log("msg", "Failed to write response", "error", err)
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

	level.Info(logger).Log("msg", "Starting server", "port", srv.Addr)
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
		level.Error(logger).Log("msg", "Fatal error", "error", err)
	}

	level.Info(logger).Log("msg", "Server stopped")
}
