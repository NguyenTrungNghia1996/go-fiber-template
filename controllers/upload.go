package controllers

import (
	"context"
	"os"
	"strconv"
	"time"

	"go-fiber-api/models"
	"fmt"
	"strings"
	"github.com/gofiber/fiber/v2"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func GetUploadUrl(c *fiber.Ctx) error {
	var input models.PutObjectUpload
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"data":    nil,
		})
	}

	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucket := os.Getenv("MINIO_BUCKET")
	ssl, _ := strconv.ParseBool(os.Getenv("MINIO_SSL"))

	// Khởi tạo client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: ssl,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to init MinIO client",
			"data":    err.Error(),
		})
	}

	// Tạo URL tạm thời để upload file
	presignedURL, err := minioClient.PresignedPutObject(
		context.Background(),
		bucket,
		input.Key,
		15*time.Minute,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create upload URL",
			"data":    err.Error(),
		})
	}
	// directURL := "https://" + endpoint + "/" + bucket + "/" + input.Key
	publicURL := os.Getenv("MINIO_PUBLIC_URL")
objectKey := input.Key
directURL := fmt.Sprintf("%s/%s", strings.TrimRight(publicURL, "/"), objectKey)
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Upload URL generated successfully",
		"data": fiber.Map{
			"upload_url": presignedURL.String(),
			"direct_url": directURL,
		},
	})
}
