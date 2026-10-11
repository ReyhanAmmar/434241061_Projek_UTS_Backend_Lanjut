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
	Pool            *pgxpool.Pool
	JWT             *helper.JWTManager
	Permissions     *helper.PermissionSet
	UserService     *service.UserService
	AuthService     *service.AuthService
	CategoryService *service.CategoryService

	TransactionService *service.TransactionService
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

	users := api.Group(
		"/users",
		middleware.RequireAuth(deps.JWT),
		middleware.RequireJSON,
	)

	perms := deps.Permissions

	users.Get(
		"/",
		middleware.RequirePermission(perms, "user:list"),
		deps.UserService.List,
	)

	users.Patch(
		"/:id/role",
		middleware.RequirePermission(perms, "role:assign"),
		deps.UserService.AssignRole,
	)

	users.Get("/:id", deps.UserService.Get)

	categories := api.Group(
		"/categories",
		middleware.RequireAuth(deps.JWT),
		middleware.RequireJSON,
	)

	categories.Get(
		"/",
		middleware.RequirePermission(perms, "category:list"),
		deps.CategoryService.List,
	)

	categories.Post(
		"/",
		middleware.RequirePermission(perms, "category:create"),
		deps.CategoryService.Create,
	)

	categories.Delete(
		"/:id",
		middleware.RequirePermission(perms, "category:delete"),
		deps.CategoryService.Delete,
	)

	categories.Get("/:id", deps.CategoryService.Get)
	categories.Put("/:id", deps.CategoryService.Replace)
	categories.Patch("/:id", deps.CategoryService.Patch)

	transactions := api.Group(
		"/transactions",
		middleware.RequireAuth(deps.JWT),
		middleware.RequireJSON,
	)

	transactions.Get(
		"/",
		middleware.RequirePermission(perms, "transaction:list"),
		deps.TransactionService.List,
	)

	transactions.Get(
		"/feed",
		middleware.RequirePermission(perms, "transaction:list"),
		deps.TransactionService.Feed,
	)

	transactions.Get(
		"/summary",
		middleware.RequirePermission(perms, "transaction:list"),
		deps.TransactionService.Summary,
	)

	transactions.Get(
		"/export",
		middleware.RequirePermission(perms, "transaction:list"),
		deps.TransactionService.Export,
	)

	transactions.Post(
		"/",
		middleware.RequirePermission(perms, "transaction:create"),
		deps.TransactionService.Create,
	)

	transactions.Delete(
		"/:id",
		middleware.RequirePermission(perms, "transaction:delete"),
		deps.TransactionService.Delete,
	)

	transactions.Get("/:id", deps.TransactionService.Get)
	transactions.Put("/:id", deps.TransactionService.Replace)
	transactions.Patch("/:id", deps.TransactionService.Patch)
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