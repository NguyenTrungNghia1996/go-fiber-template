package controllers

import (
	"go-fiber-api/models"
	"go-fiber-api/repositories"
	"go-fiber-api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
)

// CreateUser handles the creation of a new user
// POST /api/users
// Body:
//
//	{
//	  "username": "teacher1",
//	  "password": "123456",
//	  "email": "teacher@example.com",
//	  "role": "member",
//	  "person_id": "665abc..."
//	}
func CreateUser(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Status:  "error",
			Message: "Invalid data",
			Data:    nil,
		})
	}

	// Kiểm tra username đã tồn tại chưa
	exists, err := repositories.IsUsernameExists(user.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Error checking username",
			Data:    nil,
		})
	}
	if exists {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Status:  "error",
			Message: "Username already exists",
			Data:    nil,
		})
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Failed to hash password",
			Data:    nil,
		})
	}
	user.Password = hashedPassword

	// Tạo user
	if err := repositories.CreateUser(&user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Unable to create user",
			Data:    nil,
		})
	}

	// Xoá password trước khi trả về frontend
	user.Password = ""

	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Created user successfully",
		Data:    user,
	})
}

// GetUsersByRole retrieves users by their role
// GET /api/users?role=member
func GetUsersByRole(c *fiber.Ctx) error {
	role := c.Query("role")

	users, err := repositories.GetUsersByRole(role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Cannot get user list",
			Data:    nil,
		})
	}

	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Get user list successfully",
		Data:    users,
	})
}

// UpdateUserPersonID updates a user's associated person_id
// PUT /api/users/person
// Body:
//
//	{
//	  "id": "665e1b3fa6ef0c2d7e3e594f",
//	  "person_id": "665e1cdbabc123..."
//	}
func UpdateUserPersonID(c *fiber.Ctx) error {
	var body struct {
		ID       string `json:"id"`
		PersonID string `json:"person_id"`
	}
	if err := c.BodyParser(&body); err != nil || body.ID == "" || body.PersonID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Status:  "error",
			Message: "Invalid data",
			Data:    nil,
		})
	}

	err := repositories.UpdateUserPersonID(body.ID, body.PersonID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Unable to update PersonID",
			Data:    nil,
		})
	}

	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "PersonID has been updated successfully.g",
		Data:    nil,
	})
}

// ChangeUserPassword updates the current user's password using JWT info
// PUT /api/users/password
// Body:
//
//	{
//	  "old_password": "admin123",
//	  "new_password": "123456"
//	}
func ChangeUserPassword(c *fiber.Ctx) error {
	var body struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := c.BodyParser(&body); err != nil || body.OldPassword == "" || body.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Status:  "error",
			Message: "Invalid data",
			Data:    nil,
		})
	}

	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	id, _ := claims["id"].(string)

	user, err := repositories.FindUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.APIResponse{
			Status:  "error",
			Message: "User not found",
			Data:    nil,
		})
	}

	if !utils.CheckPasswordHash(body.OldPassword, user.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIResponse{
			Status:  "error",
			Message: "Old password is incorrect",
			Data:    nil,
		})
	}

	hashed, err := utils.HashPassword(body.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Unable to encrypt passwordu",
			Data:    nil,
		})
	}

	err = repositories.UpdateUserPassword(id, hashed)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Unable to update password",
			Data:    nil,
		})
	}

	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Password changed successfully",
		Data:    nil,
	})
}

// GetCurrentUser returns the authenticated user's information from JWT
// GET /api/me
func GetCurrentUser(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	id, _ := claims["id"].(string)

	user, err := repositories.FindUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.APIResponse{
			Status:  "error",
			Message: "User not found",
			Data:    nil,
		})
	}
	user.Password = ""
	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Get profile successfully",
		Data:    user,
	})
}
