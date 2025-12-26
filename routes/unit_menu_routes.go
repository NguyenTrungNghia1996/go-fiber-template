package routes

import (
	"go-fiber-api/controllers"
	"go-fiber-api/pkg/auth"

	"github.com/gofiber/fiber/v2"
)

// RegisterUnitMenuRoutes exposes GET for unit menus derived from active service packages.
func RegisterUnitMenuRoutes(app *fiber.App, ctrl *controllers.UnitMenuController) {
	g := app.Group("/unit/unit_menus")
	g.Use(auth.RequireUser())
	useGetCache(g)
	g.Get("/", ctrl.List)
}
