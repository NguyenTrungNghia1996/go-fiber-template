package routes

import (
	"go-fiber-api/controllers"
	// "go-fiber-api/repositories"
	"go-fiber-api/middleware"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func Setup(app *fiber.App, db *mongo.Database) {
	// teacherController := controllers.NewTeacherController(repositories.NewTeacherRepository(db))
	// teachers := api.Group("/teachers")
	// Auth
	app.Post("/login", controllers.Login)
	app.Get("/test", controllers.Hello)
	// Protected API group
	api := app.Group("/api", middleware.Protected())
	api.Get("/test2", controllers.Hello)
	// Upload URL
	api.Put("/presigned_url", controllers.GetUploadUrl)
}
