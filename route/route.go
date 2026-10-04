package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-buku-kas/app/service"
	"api-buku-kas/helper"
	"api-buku-kas/middleware"
)

type Dependencies struct {
	Pool        *pgxpool.Pool
	JWT         *helper.JWTManager
	AuthService *service.AuthService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(deps.Pool))

	auth := api.Group("/auth", middleware.RequireJSON)

	auth.Post("/register", deps.AuthService.Register)
	auth.Post(
		"/login",
		middleware.LoginRateLimiter(),
		deps.AuthService.Login,
	)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get(
		"/me",
		middleware.RequireAuth(deps.JWT),
		deps.AuthService.Me,
	)
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(
			c.UserContext(), 2*time.Second,
		)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable(
				"database tidak dapat dihubungi", err,
			)
		}

		return helper.Success(
			c,
			fiber.StatusOK,
			"server dan database berjalan",
			nil,
		)
	}
}