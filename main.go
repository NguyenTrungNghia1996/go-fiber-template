package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"go-fiber-api/config"
	"go-fiber-api/controllers"
	"go-fiber-api/pkg/cloudflare"
	"go-fiber-api/pkg/response"
	"go-fiber-api/repositories"
	"go-fiber-api/routes"
	"go-fiber-api/seed"

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

	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	skipDNS := appEnv == "" || appEnv == "dev" || appEnv == "development"
	envLabel := appEnv
	if envLabel == "" {
		envLabel = "development"
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
	saMenuRepo, err := repositories.NewSuperAdminMenuRepository(config.DB)
	if err != nil {
		log.Fatalf("failed to init superadmin menu repository: %v", err)
	}
	saRoleGroupRepo, err := repositories.NewSuperAdminRoleGroupRepository(config.DB)
	if err != nil {
		log.Fatalf("failed to init superadmin role group repository: %v", err)
	}
	saCtrl := controllers.NewSuperAdminController(saRepo, saRoleGroupRepo)
	saMenuCtrl := controllers.NewSuperAdminMenuController(saMenuRepo)
	saRoleGroupCtrl := controllers.NewSuperAdminRoleGroupController(saRoleGroupRepo, saRepo)
	routes.RegisterAuthRoutes(app, saCtrl)
	routes.RegisterSuperAdminRoutes(app, saCtrl)
	routes.RegisterSuperAdminMenuRoutes(app, saMenuCtrl)
	routes.RegisterSuperAdminRoleGroupRoutes(app, saRoleGroupCtrl)

	// UnitServicePackageRegistration feature (init before Units so it can be injected)
	unitServicePackageRegistrationRepo, err := repositories.NewUnitServicePackageRegistrationRepository(config.DB)
	if err != nil {
		log.Fatalf("failed to init unit service package registration repository: %v", err)
	}

	// ServicePackage feature (repo needed by Units controller)
	servicePackageRepo := repositories.NewServicePackageRepository(config.DB)

	// Unit users repo (used by Units + UnitAuth)
	unitUserRepo := repositories.NewUnitUserRepository(config.DB)

	// Cloudflare DNS for unit subdomains (skipped in dev environments)
	var dnsClient *cloudflare.DNSClient
	if skipDNS {
		log.Printf("APP_ENV=%s -> skipping Cloudflare DNS provisioning for units", envLabel)
	} else {
		dnsClient, err = cloudflare.NewDNSClientFromEnv()
		if err != nil {
			log.Printf("cloudflare dns not configured: %v", err)
		}
	}

	// Units feature
	unitRepo, err := repositories.NewUnitRepository(config.DB, unitServicePackageRegistrationRepo)
	if err != nil {
		log.Fatalf("failed to init unit repository: %v", err)
	}
	unitCtrl := controllers.NewUnitController(unitRepo, unitServicePackageRegistrationRepo, servicePackageRepo, unitUserRepo, dnsClient, skipDNS)
	routes.RegisterUnitRoutes(app, unitCtrl)

	// UnitServicePackageRegistration feature routes after unit repo is ready
	unitServicePackageRegistrationCtrl := controllers.NewUnitServicePackageRegistrationController(unitServicePackageRegistrationRepo, unitRepo, servicePackageRepo)
	routes.RegisterUnitServicePackageRegistrationRoutes(app, unitServicePackageRegistrationCtrl)

	// ServicePackage routes
	servicePackageCtrl := controllers.NewServicePackageController(servicePackageRepo)
	routes.RegisterServicePackageRoutes(app, servicePackageCtrl)
	servicePackageMenuCtrl := controllers.NewServicePackageMenuController(servicePackageRepo)
	routes.RegisterServicePackageMenuRoutes(app, servicePackageMenuCtrl)

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

	// Unit menus derived from active service packages (for any authenticated unit user)
	unitMenuCtrl := controllers.NewUnitMenuController(unitServicePackageRegistrationRepo, servicePackageRepo)
	routes.RegisterUnitMenuRoutes(app, unitMenuCtrl)

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
