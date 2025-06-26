package routes

import (
	"go-fiber-api/controllers"
	"go-fiber-api/middleware"
	"go-fiber-api/repositories"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func Setup(app *fiber.App, db *mongo.Database) {
	// Initialize repositories and controllers once so they can be reused
	userRepo := repositories.NewUserRepository(db)
	userCtrl := controllers.NewUserController(userRepo)
	authCtrl := controllers.NewAuthController(userRepo)
	menuRepo := repositories.NewMenuRepository(db)
	menuCtrl := controllers.NewMenuController(menuRepo)
	roleGroupRepo := repositories.NewRoleGroupRepository(db)
	roleGroupCtrl := controllers.NewRoleGroupController(roleGroupRepo)

	// Public routes do not require authentication
	app.Post("/login", authCtrl.Login)
	app.Get("/test", controllers.Hello)

	// Protected API group requires JWT authentication
	api := app.Group("/api", middleware.Protected())
	api.Get("/test2", controllers.Hello)

	// Endpoints accessible to any authenticated user
	api.Get("/me", userCtrl.GetCurrentUser)
	api.Put("/users/password", userCtrl.ChangeUserPassword)
	api.Put("/presigned_url", controllers.GetUploadUrl)

	// Admin-only routes are nested under /api/users
	admin := api.Group("/users", middleware.AdminOnly())
	admin.Post("/", userCtrl.CreateUser)
	admin.Get("/", userCtrl.GetUsersByRole)

	menuAdmin := api.Group("/menus", middleware.AdminOnly())
	menuAdmin.Post("/", menuCtrl.CreateMenu)
	menuAdmin.Put("/", menuCtrl.UpdateMenu)
	menuAdmin.Get("/", menuCtrl.GetMenus)
	menuAdmin.Delete("/", menuCtrl.DeleteMenu)

	roleGroupAdmin := api.Group("/role-groups", middleware.AdminOnly())
	roleGroupAdmin.Post("/", roleGroupCtrl.CreateRoleGroup)
	roleGroupAdmin.Put("/", roleGroupCtrl.UpdateRoleGroup)
	roleGroupAdmin.Get("/", roleGroupCtrl.GetRoleGroups)
	roleGroupAdmin.Delete("/", roleGroupCtrl.DeleteRoleGroup)
}
