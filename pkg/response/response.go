package response

import (
	"github.com/gofiber/fiber/v2"
)

// APIResponse defines the standard response envelope for all APIs.
type APIResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Status  string      `json:"status"`
}

// Success sends a success response with a standard envelope.
func Success(c *fiber.Ctx, data interface{}, message string, statusCode ...int) error {
	code := fiber.StatusOK
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	return c.Status(code).JSON(APIResponse{
		Data:    data,
		Message: message,
		Status:  "success",
	})
}

// Error sends an error response with a standard envelope.
// Optional data can be provided for additional error context (set to nil otherwise).
func Error(c *fiber.Ctx, message string, code int, data interface{}) error {
	return c.Status(code).JSON(APIResponse{
		Data:    data,
		Message: message,
		Status:  "error",
	})
}
