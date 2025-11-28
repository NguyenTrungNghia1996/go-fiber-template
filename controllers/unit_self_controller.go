package controllers

import (
	"strings"

	"go-fiber-api/pkg/response"
	"go-fiber-api/repositories"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// UnitSelfController allows a unit admin to view/update their unit info.
type UnitSelfController struct {
	unitRepo *repositories.UnitRepository
}

func NewUnitSelfController(unitRepo *repositories.UnitRepository) *UnitSelfController {
	return &UnitSelfController{unitRepo: unitRepo}
}

// Get returns current unit detail inferred from token.
func (h *UnitSelfController) Get(c *fiber.Ctx) error {
	unitID, _ := c.Locals("unit_id").(string)
	unitID = strings.TrimSpace(unitID)
	if unitID == "" {
		return response.Error(c, "unauthorized", fiber.StatusUnauthorized, nil)
	}
	u, err := h.unitRepo.FindByID(c.Context(), unitID)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if u == nil {
		return response.Error(c, "not found", fiber.StatusNotFound, nil)
	}
	return response.Success(c, u, "ok")
}

// Update allows unit admin to update name/description/logo_url of their unit.
func (h *UnitSelfController) Update(c *fiber.Ctx) error {
	// Must be authenticated as unit user and admin
	if !isUnitAdmin(c) {
		return response.Error(c, "forbidden", fiber.StatusForbidden, nil)
	}
	unitID, _ := c.Locals("unit_id").(string)
	unitID = strings.TrimSpace(unitID)
	if unitID == "" {
		return response.Error(c, "unauthorized", fiber.StatusUnauthorized, nil)
	}
	var in struct {
		Name        *string `json:"name,omitempty"`
		Description *string `json:"description,omitempty"`
		LogoURL     *string `json:"logo_url,omitempty"`
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
	if in.Description != nil {
		updates = append(updates, bson.E{Key: "description", Value: strings.TrimSpace(*in.Description)})
	}
	if in.LogoURL != nil {
		updates = append(updates, bson.E{Key: "logo_url", Value: strings.TrimSpace(*in.LogoURL)})
	}
	if len(updates) == 0 {
		return response.Error(c, "no fields to update", fiber.StatusBadRequest, nil)
	}
	u, err := h.unitRepo.UpdateByID(c.Context(), unitID, updates, nil)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if u == nil {
		return response.Error(c, "not found", fiber.StatusNotFound, nil)
	}
	return response.Success(c, u, "updated")
}
