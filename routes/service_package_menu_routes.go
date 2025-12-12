package routes

import (
	"go-fiber-api/controllers"
	"go-fiber-api/pkg/auth"

	"github.com/gofiber/fiber/v2"
)

func RegisterServicePackageMenuRoutes(app *fiber.App, ctrl *controllers.ServicePackageMenuController) {
	g := app.Group("/service_package_menus")
	g.Use(auth.RequireAdmin())
	g.Get("/", ctrl.List)
	g.Post("/", ctrl.Create)
	g.Put("/", ctrl.Update)
	g.Delete("/", ctrl.Delete)
}
