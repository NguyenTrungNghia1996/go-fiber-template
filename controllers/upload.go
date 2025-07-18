package controllers

import (
	"context"
	"os"
	"strconv"
	"time"

	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go-fiber-api/models"
	"strings"
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

func DeleteImage(c *fiber.Ctx) error {
	key := c.Query("key")
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing key",
			"data":    nil,
		})
	}

	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucket := os.Getenv("MINIO_BUCKET")
	ssl, _ := strconv.ParseBool(os.Getenv("MINIO_SSL"))

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

	err = minioClient.RemoveObject(context.Background(), bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete image",
			"data":    err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Deleted image successfully",
		"data":    nil,
	})
}
