package routes

import (
    "go-fiber-api/controllers"

    "github.com/gofiber/fiber/v2"
)

func RegisterSuperAdminRoutes(app *fiber.App, ctrl *controllers.SuperAdminController) {
    g := app.Group("/superadmins")
    // GET: list or get by query ?id=...
    g.Get("/", ctrl.List)
    // POST: create
    g.Post("/", ctrl.Create)
    // PUT: update with id in body
    g.Put("/", ctrl.Update)
    // DELETE: delete with ?id=...
    g.Delete("/", ctrl.Delete)
}
