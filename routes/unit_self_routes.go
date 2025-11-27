package routes

import (
    "go-fiber-api/controllers"
    "go-fiber-api/pkg/auth"

    "github.com/gofiber/fiber/v2"
)

// RegisterUnitSelfRoutes exposes GET/PUT for unit admins to manage their unit.
func RegisterUnitSelfRoutes(app *fiber.App, ctrl *controllers.UnitSelfController) {
    g := app.Group("/unit_self")
    g.Use(auth.RequireUser())
    g.Get("/", ctrl.Get)
    g.Put("/", ctrl.Update)
}

