package helper

import (
	"github.com/gofiber/fiber/v2"

	"api-buku-kas/app/model"
)

func Success(
	c *fiber.Ctx, status int, message string, data any,
) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SuccessList(
	c *fiber.Ctx, message string, data any, meta *model.Meta,
) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func Created(
	c *fiber.Ctx, message string, data any, location string,
) error {
	c.Set(fiber.HeaderLocation, location)

	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}
