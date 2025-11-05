package main

import (
    "context"
    "log"
    "os"
    "time"

    "go-fiber-api/config"

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

    // Endpoint kiểm tra sức khỏe hệ thống và kết nối DB
    app.Get("/health", func(c *fiber.Ctx) error {
        ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
        defer cancel()
        // Ping DB thông qua lệnh ping
        err := config.DB.RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Err()
        dbStatus := "up"
        if err != nil {
            dbStatus = "down"
        }
        return c.JSON(fiber.Map{
            "status": "ok",
            "db":     dbStatus,
        })
    })

    port := os.Getenv("PORT")
    if port == "" {
        port = "4000"
    }
    log.Fatal(app.Listen(":" + port))
}
