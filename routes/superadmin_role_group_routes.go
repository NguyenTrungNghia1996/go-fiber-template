package routes

import (
	"go-fiber-api/controllers"
	"go-fiber-api/pkg/auth"

	"github.com/gofiber/fiber/v2"
)

func RegisterSuperAdminRoleGroupRoutes(app *fiber.App, ctrl *controllers.SuperAdminRoleGroupController) {
	g := app.Group("/admin/superadmin_role_groups")
	g.Use(auth.RequireAdmin())
	useGetCache(g)
	g.Get("/", ctrl.List)
	g.Post("/", ctrl.Create)
	g.Put("/", ctrl.Update)
	g.Delete("/", ctrl.Delete)
}
