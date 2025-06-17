package routes

import (
	"go-fiber-api/controllers"
	// "go-fiber-api/repositories"
	"github.com/gofiber/fiber/v2"
	"go-fiber-api/middleware"
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
	api.Get("/me", controllers.GetCurrentUser)
	api.Put("/users/password", controllers.ChangeUserPassword)
	// Upload URL
	api.Put("/presigned_url", controllers.GetUploadUrl)

	admin := api.Group("/users", middleware.AdminOnly())
	admin.Post("/", controllers.CreateUser)
	admin.Get("/", controllers.GetUsersByRole)
}
