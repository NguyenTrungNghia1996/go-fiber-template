package config

// Package config provides helpers for connecting to Postgres via GORM.

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"go-fiber-api/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// LoadDotEnv searches for a .env file in parent directories and loads it.
func LoadDotEnv() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	for {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			return godotenv.Load(envPath)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return fmt.Errorf(".env not found in any parent directory")
}

// ConnectDB initializes the global Postgres connection using GORM, applies basic
// connection pooling settings, and runs auto migrations.
func ConnectDB() {
	start := time.Now()

	if err := LoadDotEnv(); err != nil {
		log.Println("⚠️ Không tìm thấy .env, sẽ dùng biến môi trường hệ thống nếu có")
	} else {
		log.Println("✅ Đã load file .env thành công")
	}

	dsn := os.Getenv("POSTGRES_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		log.Fatal("❌ POSTGRES_URL (hoặc DATABASE_URL) không được để trống")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("Lỗi kết nối Postgres: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Lỗi lấy sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatalf("Không thể ping Postgres: %v", err)
	}

	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		log.Fatalf("Không thể tạo extension uuid-ossp: %v", err)
	}

	if err := db.AutoMigrate(
		&models.Unit{},
		&models.ServicePackage{},
		&models.UnitServicePackageRegistration{},
		&models.UnitUser{},
		&models.SuperAdmin{},
	); err != nil {
		log.Fatalf("Chạy migration thất bại: %v", err)
	}

	DB = db
	log.Printf("✅ Kết nối Postgres (GORM) thành công sau %v\n", time.Since(start))
}
