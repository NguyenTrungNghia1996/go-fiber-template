package routes

import (
	"go-fiber-api/controllers"
	"go-fiber-api/middleware"
	"go-fiber-api/repositories"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func Setup(app *fiber.App, db *mongo.Database) {
	userRepo := repositories.NewUserRepository(db)
	userCtrl := controllers.NewUserController(userRepo)

	// Public routes
	app.Post("/login", controllers.Login)
	app.Get("/test", controllers.Hello)

	// Protected API group
	api := app.Group("/api", middleware.Protected())
	api.Get("/test2", controllers.Hello)

	api.Get("/me", userCtrl.GetCurrentUser)
	api.Put("/users/password", userCtrl.ChangeUserPassword)
	api.Put("/presigned_url", controllers.GetUploadUrl)

	// Admin-only routes
	admin := api.Group("/users", middleware.AdminOnly())
	admin.Post("/", userCtrl.CreateUser)
	admin.Get("/", userCtrl.GetUsersByRole)
}
