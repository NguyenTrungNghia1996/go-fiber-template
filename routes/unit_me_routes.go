package routes

import (
	"go-fiber-api/controllers"
	"go-fiber-api/pkg/auth"

	"github.com/gofiber/fiber/v2"
)

// RegisterUnitMeRoutes exposes GET/PUT for a unit user to manage own profile.
func RegisterUnitMeRoutes(app *fiber.App, ctrl *controllers.UnitMeController) {
	g := app.Group("/unit/unit_me")
	g.Use(auth.RequireUser())
	useGetCache(g)
	g.Get("/", ctrl.Get)
	g.Put("/", ctrl.Update)
}
