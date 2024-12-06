package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/olawolu/zk-pass/logger"
	"github.com/olawolu/zk-pass/server"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Getenv, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	getenv func(string) string,
	args []string,
) error {
	_, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	appEnv := args[1]
	setEnv(appEnv)

	host := getenv("host")
	port := getenv("port")
	// setup databse
	logger := logger.NewLogger()
	config := server.ServerConfig(host, port)
	serverInstance := server.NewServer(nil, logger, nil, nil)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(config.Host, config.Port),
		Handler: serverInstance,
	}

	go func() {
		slog.Info(fmt.Sprintf("listening on %s\n", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()

	return nil
}

func setEnv(appEnv string) error {
	switch appEnv {
	case "local":
		if err := loadBaseEnv(); err != nil {
			return err
		}
	case "test":
		if err := loadBaseEnv(); err != nil {
			return err
		}
		// overloadSpecificEnv(".env.test.local")
	case "production":
		// env should already be set
	default:
		slog.Warn("Application environment was not specified with a recognised APP_ENV")
	}
	return nil
}

func loadBaseEnv() error {
	var err error
	if err = godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %v", err)
	}
	return nil
}

/*
	getenv := func(key string) string {
		switch key {
		case "MYAPP_FORMAT":
			return "markdown"
		case "MYAPP_TIMEOUT":
			return "5s"
		default:
			return os.Getenv(key)
		}
	}
*/
