package routes

import (
	"go-fiber-api/controllers"
	"go-fiber-api/pkg/auth"

	"github.com/gofiber/fiber/v2"
)

func RegisterSuperAdminRoutes(app *fiber.App, ctrl *controllers.SuperAdminController) {
	g := app.Group("/superadmins")
	g.Use(auth.RequireAdmin())
	// GET: list or get by query ?id=...
	g.Get("/", ctrl.List)
	// GET: aggregate permissions of a super admin (default to current token)
	g.Get("/permissions", ctrl.Permissions)
	// POST: create
	g.Post("/", ctrl.Create)
	// PUT: update with id in body
	g.Put("/", ctrl.Update)
	// DELETE: delete with ?id=...
	g.Delete("/", ctrl.Delete)
}

// RegisterAuthRoutes registers auth endpoints.
func RegisterAuthRoutes(app *fiber.App, ctrl *controllers.SuperAdminController) {
	app.Post("/auth/admin", ctrl.Login)
}
