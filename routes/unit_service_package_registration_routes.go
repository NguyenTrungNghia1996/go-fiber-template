package routes

import (
	"go-fiber-api/controllers"
	"go-fiber-api/pkg/auth"

	"github.com/gofiber/fiber/v2"
)

func RegisterUnitServicePackageRegistrationRoutes(app *fiber.App, ctrl *controllers.UnitServicePackageRegistrationController) {
	g := app.Group("/admin/unit_service_package_registrations")
	g.Use(auth.RequireAdmin())
	useGetCache(g)
	g.Get("/", ctrl.List)
	g.Post("/", ctrl.Create)
	g.Put("/", ctrl.Update)
	g.Delete("/", ctrl.Delete)
}
