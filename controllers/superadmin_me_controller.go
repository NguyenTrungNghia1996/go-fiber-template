package controllers

import (
	"strings"

	"go-fiber-api/pkg/response"
	"go-fiber-api/repositories"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

// SuperAdminMeController lets a super admin view and update their own profile.
type SuperAdminMeController struct {
	repo *repositories.SuperAdminRepository
}

func NewSuperAdminMeController(repo *repositories.SuperAdminRepository) *SuperAdminMeController {
	return &SuperAdminMeController{repo: repo}
}

// Get returns the current super admin info based on the token.
func (h *SuperAdminMeController) Get(c *fiber.Ctx) error {
	adminID, _ := c.Locals("admin_id").(string)
	adminID = strings.TrimSpace(adminID)
	if adminID == "" {
		return response.Error(c, "unauthorized", fiber.StatusUnauthorized, nil)
	}
	sa, err := h.repo.FindByID(c.Context(), adminID)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if sa == nil {
		return response.Error(c, "not found", fiber.StatusNotFound, nil)
	}
	return response.Success(c, sa, "ok")
}

// Update lets the current super admin update profile fields.
func (h *SuperAdminMeController) Update(c *fiber.Ctx) error {
	adminID, _ := c.Locals("admin_id").(string)
	adminID = strings.TrimSpace(adminID)
	if adminID == "" {
		return response.Error(c, "unauthorized", fiber.StatusUnauthorized, nil)
	}
	var in struct {
		Name     *string `json:"name,omitempty"`
		Email    *string `json:"email,omitempty"`
		ImageURL *string `json:"image_url,omitempty"`
		Password *string `json:"password,omitempty"`
	}
	if err := c.BodyParser(&in); err != nil {
		return response.Error(c, "invalid body", fiber.StatusBadRequest, nil)
	}
	updates := bson.D{}
	if in.Name != nil {
		v := strings.TrimSpace(*in.Name)
		if v == "" {
			return response.Error(c, "name cannot be empty", fiber.StatusBadRequest, nil)
		}
		updates = append(updates, bson.E{Key: "name", Value: v})
	}
	if in.Email != nil {
		updates = append(updates, bson.E{Key: "email", Value: strings.TrimSpace(*in.Email)})
	}
	if in.ImageURL != nil {
		updates = append(updates, bson.E{Key: "image_url", Value: strings.TrimSpace(*in.ImageURL)})
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
		updates = append(updates, bson.E{Key: "password_hash", Value: string(hash)})
	}
	if len(updates) == 0 {
		return response.Error(c, "no fields to update", fiber.StatusBadRequest, nil)
	}
	updated, err := h.repo.UpdateByID(c.Context(), adminID, updates)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if updated == nil {
		return response.Error(c, "not found", fiber.StatusNotFound, nil)
	}
	return response.Success(c, updated, "updated")
}
