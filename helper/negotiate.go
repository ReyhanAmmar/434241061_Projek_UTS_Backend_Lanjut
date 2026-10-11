	package helper

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

const MIMETextCSV = "text/csv"

func NegotiateFormat(c *fiber.Ctx) (string, error) {
	if format := strings.ToLower(c.Query("format")); format != "" {
		switch format {
		case "json", "csv":
			return format, nil
		default:
			return "", BadRequest("format harus json atau csv")
		}
	}

	switch c.Accepts(fiber.MIMEApplicationJSON, MIMETextCSV) {
	case fiber.MIMEApplicationJSON:
		return "json", nil
	case MIMETextCSV:
		return "csv", nil
	default:
		return "", NotAcceptable(
			"format tidak didukung, gunakan application/json atau text/csv",
		)
	}
}