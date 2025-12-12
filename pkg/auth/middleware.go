package auth

import (
	"strings"

	"go-fiber-api/pkg/response"

	"github.com/gofiber/fiber/v2"
)

// RequireAdmin verifies Authorization: Bearer <token> using JWT_SECRET_ADMIN.
func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Error(c, "missing Authorization header", fiber.StatusUnauthorized, nil)
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return response.Error(c, "invalid Authorization header", fiber.StatusUnauthorized, nil)
		}
		tokenStr := strings.TrimSpace(parts[1])
		claims, err := VerifyToken(tokenStr)
		if err != nil {
			return response.Error(c, "unauthorized", fiber.StatusUnauthorized, nil)
		}
		if sub, ok := claims["sub"].(string); ok && sub != "" {
			c.Locals("admin_id", sub)
		}
		if ia, ok := claims["is_admin"].(bool); ok {
			c.Locals("is_admin", ia)
		}
		return c.Next()
	}
}

// RequireUser verifies Authorization: Bearer <token> using JWT_SECRET for unit users.
func RequireUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Error(c, "missing Authorization header", fiber.StatusUnauthorized, nil)
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return response.Error(c, "invalid Authorization header", fiber.StatusUnauthorized, nil)
		}
		tokenStr := strings.TrimSpace(parts[1])
		claims, err := VerifyUserToken(tokenStr)
		if err != nil {
			return response.Error(c, "unauthorized", fiber.StatusUnauthorized, nil)
		}
		if sub, ok := claims["sub"].(string); ok && sub != "" {
			c.Locals("user_id", sub)
		}
		if unitID, ok := claims["unit_id"].(string); ok && unitID != "" {
			c.Locals("unit_id", unitID)
		}
		if sd, ok := claims["subdomain"].(string); ok && sd != "" {
			c.Locals("unit_subdomain", sd)
		}
		if ia, ok := claims["is_admin"].(bool); ok {
			c.Locals("is_admin", ia)
		}
		return c.Next()
	}
}
