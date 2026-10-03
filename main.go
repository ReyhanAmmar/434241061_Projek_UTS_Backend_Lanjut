package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-buku-kas/config"
	"api-buku-kas/database"
	"api-buku-kas/route"
)

func main() {
	config.LoadEnv()
	logger := config.NewLogger()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error(
			"gagal terhubung ke database",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer pool.Close()

	logger.Info("database terhubung")

	app := config.NewApp(logger, route.Dependencies{
		Pool: pool,
	})

	port := config.GetEnv("APP_PORT", "3000")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	listenErr := make(chan error, 1)

	go func() {
		listenErr <- app.Listen(":" + port)
	}()

	select {
	case err := <-listenErr:
		if err != nil {
			logger.Error(
				"server gagal berjalan",
				slog.String("error", err.Error()),
			)
			pool.Close()
			os.Exit(1)
		}
		return

	case <-quit:
		logger.Info("sinyal berhenti diterima, menutup server")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(), 10*time.Second,
	)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error(
			"gagal menutup server dengan rapi",
			slog.String("error", err.Error()),
		)
		pool.Close()
		os.Exit(1)
	}

	logger.Info("server berhenti dengan rapi")
}
