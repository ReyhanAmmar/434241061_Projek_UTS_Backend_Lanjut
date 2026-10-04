package middleware

import (
	"github.com/gofiber/fiber/v2"

	"api-buku-kas/helper"
)

func RequirePermission(
	perms *helper.PermissionSet,
	permission string,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		current, ok := helper.CurrentUser(c)
		if !ok {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			return helper.Unauthorized("belum terautentikasi")
		}

		if !perms.Can(current.Role, permission) {
			return helper.Forbidden(
				"role " + current.Role +
					" tidak memiliki hak " + permission,
			)
		}

		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))

	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		current, ok := helper.CurrentUser(c)
		if !ok {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			return helper.Unauthorized("belum terautentikasi")
		}

		if _, granted := allowed[current.Role]; !granted {
			return helper.Forbidden(
				"role Anda tidak berhak mengakses endpoint ini",
			)
		}

		return c.Next()
	}
}