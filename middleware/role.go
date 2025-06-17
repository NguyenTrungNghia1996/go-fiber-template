package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// AdminOnly ensures that the requester has admin role
func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userToken := c.Locals("user").(*jwt.Token)
		claims := userToken.Claims.(jwt.MapClaims)
		role, _ := claims["role"].(string)
		if role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "error",
				"message": "Admin access only",
				"data":    nil,
			})
		}
		return c.Next()
	}
}
