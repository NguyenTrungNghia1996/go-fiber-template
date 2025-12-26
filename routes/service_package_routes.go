package routes

import (
	"go-fiber-api/controllers"
	"go-fiber-api/pkg/auth"

	"github.com/gofiber/fiber/v2"
)

func RegisterServicePackageRoutes(app *fiber.App, ctrl *controllers.ServicePackageController) {
	g := app.Group("/admin/service_packages")
	g.Use(auth.RequireAdmin())
	useGetCache(g)
	g.Get("/", ctrl.List)
	g.Post("/", ctrl.Create)
	g.Put("/", ctrl.Update)
	g.Delete("/", ctrl.Delete)
}
