package controllers

// This file contains authentication handlers used by the API.

import (
	"go-fiber-api/models"
	"go-fiber-api/repositories"
	"go-fiber-api/utils"

	"github.com/gofiber/fiber/v2"
)

// AuthController handles authentication requests.
type AuthController struct {
	Repo *repositories.UserRepository
}

// NewAuthController creates a controller with the provided user repository.
func NewAuthController(repo *repositories.UserRepository) *AuthController {
	return &AuthController{Repo: repo}
}

// Login authenticates a user and returns a signed JWT on success.
func (ctrl *AuthController) Login(c *fiber.Ctx) error {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Status:  "error",
			Message: "Invalid input",
			Data:    nil,
		})
	}

	user, err := ctrl.Repo.FindByUsername(c.Context(), input.Username)
	if err != nil || !utils.CheckPasswordHash(input.Password, user.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIResponse{
			Status:  "error",
			Message: "Invalid credentials",
			Data:    nil,
		})
	}

	token, _ := utils.GenerateJWT(user.ID.Hex())
	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Login successful",
		Data: fiber.Map{
			"id":    user.ID.Hex(),
			"token": token,
		},
	})
}
