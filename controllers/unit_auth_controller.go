package controllers

import (
	"strings"
	"time"

	"go-fiber-api/pkg/auth"
	"go-fiber-api/pkg/response"
	"go-fiber-api/repositories"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// UnitAuthController handles authentication for unit users.
type UnitAuthController struct {
	unitRepo *repositories.UnitRepository
	userRepo *repositories.UnitUserRepository
}

func NewUnitAuthController(unitRepo *repositories.UnitRepository, userRepo *repositories.UnitUserRepository) *UnitAuthController {
	return &UnitAuthController{unitRepo: unitRepo, userRepo: userRepo}
}

// Login authenticates a unit user with subdomain, username, password.
func (h *UnitAuthController) Login(c *fiber.Ctx) error {
	var in struct {
		Subdomain string `json:"subdomain"`
		Username  string `json:"username"`
		Password  string `json:"password"`
	}
	if err := c.BodyParser(&in); err != nil {
		return response.Error(c, "invalid body", fiber.StatusBadRequest, nil)
	}
	in.Subdomain = strings.TrimSpace(in.Subdomain)
	in.Username = strings.TrimSpace(in.Username)
	in.Password = strings.TrimSpace(in.Password)
	if in.Subdomain == "" || in.Username == "" || in.Password == "" {
		return response.Error(c, "subdomain, username and password are required", fiber.StatusBadRequest, nil)
	}

	unit, err := h.unitRepo.FindBySubdomain(c.Context(), in.Subdomain)
	if err != nil {
		return response.Error(c, "authentication failed", fiber.StatusUnauthorized, nil)
	}
	if unit == nil {
		return response.Error(c, "authentication failed", fiber.StatusUnauthorized, nil)
	}

	user, err := h.userRepo.FindByUsernameAndUnitID(c.Context(), in.Username, unit.ID)
	if err != nil {
		return response.Error(c, "authentication failed", fiber.StatusUnauthorized, nil)
	}
	if user == nil {
		return response.Error(c, "authentication failed", fiber.StatusUnauthorized, nil)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		return response.Error(c, "authentication failed", fiber.StatusUnauthorized, nil)
	}

	token, exp, err := auth.GenerateUserToken(user.ID.Hex(), unit.ID.Hex(), unit.Subdomain, user.IsAdmin, 24*time.Hour)
	if err != nil {
		return response.Error(c, "failed to issue token", fiber.StatusInternalServerError, nil)
	}

	userPayload := fiber.Map{
		"id":             user.ID.Hex(),
		"username":       user.Username,
		"name":           user.Name,
		"email":          user.Email,
		"image_url":      user.ImageURL,
		"unit_id":        unit.ID.Hex(),
		"subdomain":      unit.Subdomain,
		"is_admin":       user.IsAdmin,
		"role_group_ids": user.RoleGroupIDs,
	}

	return response.Success(c, fiber.Map{
		"token":      token,
		"expires_at": exp.Format(time.RFC3339),
		"user":       userPayload,
	}, "logged in")
}
