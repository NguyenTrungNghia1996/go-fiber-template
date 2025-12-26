package routes

import (
	"go-fiber-api/controllers"
	"go-fiber-api/pkg/auth"

	"github.com/gofiber/fiber/v2"
)

func RegisterUnitRoleGroupRoutes(app *fiber.App, ctrl *controllers.UnitRoleGroupController) {
	g := app.Group("/unit/unit_role_groups")
	g.Use(auth.RequireUser())
	useGetCache(g)
	g.Get("/", ctrl.List)
	g.Post("/", ctrl.Create)
	g.Put("/", ctrl.Update)
	g.Delete("/", ctrl.Delete)
}
