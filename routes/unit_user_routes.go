package routes

import (
    "go-fiber-api/controllers"
    "go-fiber-api/pkg/auth"

    "github.com/gofiber/fiber/v2"
)

// RegisterUnitUserRoutes registers CRUD endpoints for unit users, protected by unit user auth.
func RegisterUnitUserRoutes(app *fiber.App, ctrl *controllers.UnitUserController) {
    g := app.Group("/unit_users")
    g.Use(auth.RequireUser())
    g.Get("/", ctrl.List)
    g.Post("/", ctrl.Create)
    g.Put("/", ctrl.Update)
    g.Delete("/", ctrl.Delete)
}

