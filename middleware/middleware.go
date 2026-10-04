package middleware

import (
	"errors"
	"log/slog"
	"mime"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"api-buku-kas/helper"
)

func Register(
	app *fiber.App, logger *slog.Logger, allowedOrigins string,
) {
	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(corsPolicy(allowedOrigins))
	app.Use(RequestLogger(logger))
}

func corsPolicy(allowedOrigins string) fiber.Handler {
	if strings.TrimSpace(allowedOrigins) == "" {
		allowedOrigins = "http://localhost:5173"
	}

	return cors.New(cors.Config{
		AllowOrigins:  allowedOrigins,
		AllowMethods:  "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:  "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		ExposeHeaders: "X-Request-ID,Location",
	})
}

func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		status := c.Response().StatusCode()

		if err != nil {
			var appErr *helper.AppError
			var fiberErr *fiber.Error

			switch {
			case errors.As(err, &appErr):
				status = appErr.Status
			case errors.As(err, &fiberErr):
				status = fiberErr.Code
			default:
				status = fiber.StatusInternalServerError
			}
		}

		attrs := []any{
			slog.String("request_id", helper.RequestID(c)),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.Duration("duration", time.Since(start)),
			slog.String("ip", c.IP()),
		}

		if current, ok := helper.CurrentUser(c); ok {
			attrs = append(
				attrs,
				slog.Int("user_id", current.UserID),
				slog.String("role", current.Role),
			)
		}

		logger.Info("http_request", attrs...)

		return err
	}
}

var methodsWithBody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

func RequireJSON(c *fiber.Ctx) error {
	if methodsWithBody[c.Method()] {
		mediaType, _, err := mime.ParseMediaType(
			c.Get(fiber.HeaderContentType),
		)

		if err != nil || mediaType != fiber.MIMEApplicationJSON {
			return helper.UnsupportedMediaType(
				"Content-Type harus application/json",
			)
		}
	}

	return c.Next()
}
