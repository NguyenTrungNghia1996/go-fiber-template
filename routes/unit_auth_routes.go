package routes

import (
	"go-fiber-api/controllers"

	"github.com/gofiber/fiber/v2"
)

// RegisterUnitAuthRoutes registers authentication endpoints for unit users.
func RegisterUnitAuthRoutes(app *fiber.App, ctrl *controllers.UnitAuthController) {
	app.Post("/auth/unit", ctrl.Login)
}
