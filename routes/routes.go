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
	roleGroupRepo := repositories.NewRoleGroupRepository(db)
	userCtrl := controllers.NewUserController(userRepo, roleGroupRepo)
	authCtrl := controllers.NewAuthController(userRepo)
	menuRepo := repositories.NewMenuRepository(db)
	menuCtrl := controllers.NewMenuController(menuRepo)
	roleGroupCtrl := controllers.NewRoleGroupController(roleGroupRepo)

	// Public routes do not require authentication
	app.Post("/login", authCtrl.Login)
	app.Get("/test", controllers.Hello)

	// Protected API group requires JWT authentication
	api := app.Group("/api", middleware.Protected())
	api.Get("/test2", controllers.Hello)

	// Endpoints accessible to any authenticated user
	api.Get("/me", userCtrl.GetCurrentUser)
	api.Get("/permissions", userCtrl.GetUserPermissions)
	api.Put("/users/password", userCtrl.ChangeUserPassword)
	api.Put("/presigned_url", controllers.GetUploadUrl)

	// Admin-only routes are nested under /api/users
	admin := api.Group("/users")
	admin.Post("/", userCtrl.CreateUser)
	admin.Get("/", userCtrl.GetUsers)
	admin.Put("/", userCtrl.UpdateUser)

	menuAdmin := api.Group("/menus")
	menuAdmin.Post("/", menuCtrl.CreateMenu)
	menuAdmin.Put("/", menuCtrl.UpdateMenu)
	menuAdmin.Get("/", menuCtrl.GetMenus)
	menuAdmin.Delete("/", menuCtrl.DeleteMenu)

	roleGroupAdmin := api.Group("/role-groups")
	roleGroupAdmin.Post("/", roleGroupCtrl.CreateRoleGroup)
	roleGroupAdmin.Put("/", roleGroupCtrl.UpdateRoleGroup)
	roleGroupAdmin.Get("/detail", roleGroupCtrl.GetRoleGroupDetail)
	roleGroupAdmin.Get("/", roleGroupCtrl.GetRoleGroups)
	roleGroupAdmin.Delete("/", roleGroupCtrl.DeleteRoleGroup)
}
