package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-buku-kas/app/repository"
	"api-buku-kas/app/service"
	"api-buku-kas/config"
	"api-buku-kas/database"
	"api-buku-kas/helper"
	"api-buku-kas/route"
)

const minSecretLength = 32

func main() {
	config.LoadEnv()
	logger := config.NewLogger()

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < minSecretLength {
		logger.Error(
			"JWT_SECRET tidak diisi atau terlalu pendek",
			slog.Int("minimal_karakter", minSecretLength),
		)
		os.Exit(1)
	}

	accessMinutes := config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15)
	refreshDays := config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7)

	if accessMinutes <= 0 || refreshDays <= 0 {
		logger.Error("masa berlaku token harus lebih dari nol")
		os.Exit(1)
	}

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

	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "api-buku-kas"),
		time.Duration(accessMinutes)*time.Minute,
	)

	userRepository := repository.NewUserRepository(pool)
	tokenRepository := repository.NewTokenRepository(pool)

	authService := service.NewAuthService(
		userRepository,
		tokenRepository,
		jwtManager,
		time.Duration(refreshDays)*24*time.Hour,
	)

	app := config.NewApp(logger, route.Dependencies{
		Pool:        pool,
		JWT:         jwtManager,
		AuthService: authService,
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