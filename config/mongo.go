package config

// Package config provides helpers for connecting to MongoDB.

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

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

// ConnectDB initializes the global MongoDB connection using environment
// variables. It should be called once at application startup.
func ConnectDB() {
	start := time.Now()

	if err := LoadDotEnv(); err != nil {
		log.Println("⚠️ Không tìm thấy .env, sẽ dùng biến môi trường hệ thống nếu có")
	} else {
		log.Println("✅ Đã load file .env thành công")
	}

	mongoURI := os.Getenv("MONGO_URL")
	if mongoURI == "" {
		log.Fatal("❌ MONGO_URL không được để trống")
	}

	mongoName := os.Getenv("MONGO_NAME")
	if mongoName == "" {
		log.Fatal("❌ MONGO_NAME không được để trống")
	}

	clientOpts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.NewClient(clientOpts)
	if err != nil {
		log.Fatalf("Lỗi tạo MongoDB client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Lỗi kết nối MongoDB: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Không thể ping MongoDB: %v", err)
	}

	DB = client.Database(mongoName)
	log.Printf("✅ Kết nối MongoDB thành công sau %v\n", time.Since(start))
}
