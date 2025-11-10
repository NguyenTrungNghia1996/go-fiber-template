package main

import (
    "context"
    "log"
    "os"
    "time"

    "go-fiber-api/config"
    "go-fiber-api/controllers"
    "go-fiber-api/repositories"
    "go-fiber-api/routes"
    "go-fiber-api/seed"
    "go-fiber-api/pkg/response"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/joho/godotenv"
    "go.mongodb.org/mongo-driver/bson"
)

func main() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Println("Error loading .env file")
		} else {
			log.Println("Loaded .env file")
		}
	}

    // Kết nối MongoDB một lần duy nhất
    config.ConnectDB()

    // Khởi tạo Fiber server tối giản
    app := fiber.New()
    app.Use(cors.New())

    // Endpoint kiểm tra sức khỏe hệ thống và kết nối DB (chuẩn hóa response)
    app.Get("/health", func(c *fiber.Ctx) error {
        ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
        defer cancel()
        // Ping DB thông qua lệnh ping
        err := config.DB.RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Err()
        dbStatus := "up"
        if err != nil {
            dbStatus = "down"
        }
        return response.Success(c, fiber.Map{
            "status": "ok",
            "db":     dbStatus,
        }, "ok")
    })

    // DI wiring for SuperAdmin feature
    saRepo, err := repositories.NewSuperAdminRepository(config.DB)
    if err != nil {
        log.Fatalf("failed to init superadmin repository: %v", err)
    }
    saCtrl := controllers.NewSuperAdminController(saRepo)
    routes.RegisterAuthRoutes(app, saCtrl)
    routes.RegisterSuperAdminRoutes(app, saCtrl)

    // UnitServicePackageRegistration feature (init before Units so it can be injected)
    unitServicePackageRegistrationRepo, err := repositories.NewUnitServicePackageRegistrationRepository(config.DB)
    if err != nil {
        log.Fatalf("failed to init unit service package registration repository: %v", err)
    }
    unitServicePackageRegistrationCtrl := controllers.NewUnitServicePackageRegistrationController(unitServicePackageRegistrationRepo)
    routes.RegisterUnitServicePackageRegistrationRoutes(app, unitServicePackageRegistrationCtrl)

    // ServicePackage feature (repo needed by Units controller)
    servicePackageRepo := repositories.NewServicePackageRepository(config.DB)
    servicePackageCtrl := controllers.NewServicePackageController(servicePackageRepo)
    routes.RegisterServicePackageRoutes(app, servicePackageCtrl)

    // Unit users repo (used by Units + UnitAuth)
    unitUserRepo := repositories.NewUnitUserRepository(config.DB)

    // Units feature
    unitRepo, err := repositories.NewUnitRepository(config.DB, unitServicePackageRegistrationRepo)
    if err != nil {
        log.Fatalf("failed to init unit repository: %v", err)
    }
    unitCtrl := controllers.NewUnitController(unitRepo, unitServicePackageRegistrationRepo, servicePackageRepo, unitUserRepo)
    routes.RegisterUnitRoutes(app, unitCtrl)

    // Unit user auth (login with subdomain)
    unitAuthCtrl := controllers.NewUnitAuthController(unitRepo, unitUserRepo)
    routes.RegisterUnitAuthRoutes(app, unitAuthCtrl)

    // Unit users management (requires unit user token; admin-only enforced in handlers)
    unitUserCtrl := controllers.NewUnitUserController(unitUserRepo)
    routes.RegisterUnitUserRoutes(app, unitUserCtrl)

    // Unit self-management (unit admin can update their unit)
    unitSelfCtrl := controllers.NewUnitSelfController(unitRepo)
    routes.RegisterUnitSelfRoutes(app, unitSelfCtrl)

    // Unit user self profile (any unit user can update own info)
    unitMeCtrl := controllers.NewUnitMeController(unitUserRepo)
    routes.RegisterUnitMeRoutes(app, unitMeCtrl)

    // (already initialized above)

    // Upload (S3/MinIO) feature
    uploadCtrl := controllers.NewUploadController()
    routes.RegisterUploadRoutes(app, uploadCtrl)

    // Seed default super admin account
    if err := seed.SeedSuperAdmin(saRepo); err != nil {
        log.Printf("seed error: %v", err)
    }
    // Seed default unit admin accounts for existing units (admin/admin)
    if err := seed.SeedUnitAdmins(unitRepo, unitUserRepo); err != nil {
        log.Printf("seed error: %v", err)
    }

    port := os.Getenv("PORT")
    if port == "" {
        port = "4000"
    }
    log.Fatal(app.Listen(":" + port))
}
