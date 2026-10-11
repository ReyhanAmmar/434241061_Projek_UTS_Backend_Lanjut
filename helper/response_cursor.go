package helper

import (
	"github.com/gofiber/fiber/v2"

	"api-buku-kas/app/model"
)

func SuccessCursor(
	c *fiber.Ctx, message string, data any, meta model.CursorMeta,
) error {
	return c.Status(fiber.StatusOK).JSON(model.CursorResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}