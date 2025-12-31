package routes

import (
	"go-fiber-api/controllers"
	"go-fiber-api/pkg/auth"

	"github.com/gofiber/fiber/v2"
)

// RegisterSuperAdminMeRoutes exposes GET/PUT for a super admin to manage own profile.
func RegisterSuperAdminMeRoutes(app *fiber.App, ctrl *controllers.SuperAdminMeController) {
	g := app.Group("/admin/superadmin_me")
	g.Use(auth.RequireAdmin())
	useGetCache(g)
	g.Get("/", ctrl.Get)
	g.Put("/", ctrl.Update)
}
