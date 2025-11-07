package routes

import (
    "go-fiber-api/controllers"
    "go-fiber-api/pkg/auth"

    "github.com/gofiber/fiber/v2"
)

// RegisterUploadRoutes registers S3/MinIO upload related endpoints.
func RegisterUploadRoutes(app *fiber.App, ctrl *controllers.UploadController) {
    g := app.Group("/api")
    g.Use(auth.RequireAdmin())
    g.Put("/presigned_url", ctrl.PresignedURL)
}

