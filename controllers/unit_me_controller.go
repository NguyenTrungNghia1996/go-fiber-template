package controllers

import (
	"strings"

	// "go-fiber-api/models"
	"go-fiber-api/pkg/response"
	"go-fiber-api/repositories"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// UnitMeController allows a unit user to view and update their own profile.
type UnitMeController struct {
	repo *repositories.UnitUserRepository
}

func NewUnitMeController(repo *repositories.UnitUserRepository) *UnitMeController {
	return &UnitMeController{repo: repo}
}

func (h *UnitMeController) Get(c *fiber.Ctx) error {
	unitIDStr, _ := c.Locals("unit_id").(string)
	userIDStr, _ := c.Locals("user_id").(string)
	if unitIDStr == "" || userIDStr == "" {
		return response.Error(c, "unauthorized", fiber.StatusUnauthorized, nil)
	}
	u, err := h.repo.FindByIDWithinUnit(c.Context(), userIDStr, unitIDStr)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if u == nil {
		return response.Error(c, "not found", fiber.StatusNotFound, nil)
	}
	return response.Success(c, u, "ok")
}

// Update allows the current user to update their name, email, and/or password.
func (h *UnitMeController) Update(c *fiber.Ctx) error {
	unitIDStr, _ := c.Locals("unit_id").(string)
	userIDStr, _ := c.Locals("user_id").(string)
	if unitIDStr == "" || userIDStr == "" {
		return response.Error(c, "unauthorized", fiber.StatusUnauthorized, nil)
	}
	var in struct {
		Name     *string `json:"name,omitempty"`
		Email    *string `json:"email,omitempty"`
		Password *string `json:"password,omitempty"`
	}
	if err := c.BodyParser(&in); err != nil {
		return response.Error(c, "invalid body", fiber.StatusBadRequest, nil)
	}
	updates := map[string]interface{}{}
	if in.Name != nil {
		v := strings.TrimSpace(*in.Name)
		if v == "" {
			return response.Error(c, "name cannot be empty", fiber.StatusBadRequest, nil)
		}
		updates["name"] = v
	}
	if in.Email != nil {
		updates["email"] = strings.TrimSpace(*in.Email)
	}
	if in.Password != nil {
		v := strings.TrimSpace(*in.Password)
		if v == "" {
			return response.Error(c, "password cannot be empty", fiber.StatusBadRequest, nil)
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(v), bcrypt.DefaultCost)
		if err != nil {
			return response.Error(c, "failed to hash password", fiber.StatusInternalServerError, nil)
		}
		updates["password_hash"] = string(hash)
	}
	if len(updates) == 0 {
		return response.Error(c, "no fields to update", fiber.StatusBadRequest, nil)
	}
	u, err := h.repo.UpdateByIDWithinUnit(c.Context(), userIDStr, unitIDStr, updates)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if u == nil {
		return response.Error(c, "not found", fiber.StatusNotFound, nil)
	}
	return response.Success(c, u, "updated")
}
