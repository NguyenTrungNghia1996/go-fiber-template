package routes

import (
	"go-fiber-api/controllers"
	"go-fiber-api/pkg/auth"

	"github.com/gofiber/fiber/v2"
)

func RegisterUnitRoutes(app *fiber.App, ctrl *controllers.UnitController) {
	g := app.Group("/admin/units")
	g.Use(auth.RequireAdmin())
	useGetCache(g)
	g.Get("/", ctrl.List)
	g.Post("/", ctrl.Create)
	g.Put("/", ctrl.Update)
	g.Delete("/", ctrl.Delete)

	public := app.Group("/units")
	useGetCache(public)
	public.Get("/public_info", ctrl.PublicInfo)
}
