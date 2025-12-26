package routes

import (
	"go-fiber-api/controllers"
	"go-fiber-api/pkg/auth"

	"github.com/gofiber/fiber/v2"
)

func RegisterSuperAdminMenuRoutes(app *fiber.App, ctrl *controllers.SuperAdminMenuController) {
	g := app.Group("/admin/superadmin_menus")
	g.Use(auth.RequireAdmin())
	useGetCache(g)
	g.Get("/", ctrl.List)
	g.Post("/", ctrl.Create)
	g.Put("/", ctrl.Update)
	g.Delete("/", ctrl.Delete)
}
