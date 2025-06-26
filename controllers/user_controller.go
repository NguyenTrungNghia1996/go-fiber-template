package controllers

// This file defines handlers for user-related endpoints.

import (
	"go-fiber-api/models"
	"go-fiber-api/repositories"
	"go-fiber-api/utils"

	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

type UserController struct {
	Repo *repositories.UserRepository // provides DB operations
}

// NewUserController creates a controller with the given repository.
func NewUserController(repo *repositories.UserRepository) *UserController {
	return &UserController{Repo: repo}
}

// CreateUser handles POST /api/users and registers a new account.
func (ctrl *UserController) CreateUser(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Status:  "error",
			Message: "Invalid data",
			Data:    nil,
		})
	}

	// Check username exists
	exists, err := ctrl.Repo.IsUsernameExists(c.Context(), user.Username)
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

	// Create user
	if err := ctrl.Repo.Create(c.Context(), &user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Unable to create user",
			Data:    nil,
		})
	}

	user.Password = ""
	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Created user successfully",
		Data:    user,
	})
}

// GetUsersByRole returns a list of users optionally filtered by role.
func (ctrl *UserController) GetUsersByRole(c *fiber.Ctx) error {
	role := c.Query("role")
	search := c.Query("search")
	page, _ := strconv.ParseInt(c.Query("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.Query("limit", "10"), 10, 64)

	users, total, err := ctrl.Repo.GetByRole(c.Context(), role, search, page, limit)
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
		Data: fiber.Map{
			"items": users,
			"total": total,
		},
	})
}

// PUT /api/users/password
// ChangeUserPassword allows an authenticated user to update their password.
func (ctrl *UserController) ChangeUserPassword(c *fiber.Ctx) error {
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

	user, err := ctrl.Repo.FindByID(c.Context(), id)
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
			Message: "Unable to encrypt password",
			Data:    nil,
		})
	}

	err = ctrl.Repo.UpdatePassword(c.Context(), id, hashed)
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

// GetCurrentUser returns profile information for the authenticated user.
func (ctrl *UserController) GetCurrentUser(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	id, _ := claims["id"].(string)

	user, err := ctrl.Repo.FindByID(c.Context(), id)
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
