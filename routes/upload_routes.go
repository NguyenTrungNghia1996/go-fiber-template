package routes

import (
	"go-fiber-api/controllers"
	"go-fiber-api/pkg/auth"

	"github.com/gofiber/fiber/v2"
)

// RegisterUploadRoutes registers S3/MinIO upload related endpoints.
func RegisterUploadRoutes(app *fiber.App, ctrl *controllers.UploadController) {
	// Admin uploads (global, managed by super admin)
	admin := app.Group("/uploads")
	admin.Use(auth.RequireAdmin())
	admin.Put("/presigned_url", ctrl.PresignedURL)
	admin.Delete("/file", ctrl.Delete)

	// Unit uploads (per-tenant, used by unit users)
	unit := app.Group("/unit_uploads")
	unit.Use(auth.RequireUser())
	useGetCache(unit)
	unit.Get("/", ctrl.ListUnitFiles)
	unit.Put("/presigned_url", ctrl.PresignedURL)
	unit.Delete("/file", ctrl.Delete)
}
