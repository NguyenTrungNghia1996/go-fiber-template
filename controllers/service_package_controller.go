package controllers

import (
	"strings"

	"go-fiber-api/models"
	"go-fiber-api/pkg/response"
	"go-fiber-api/repositories"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
)

type ServicePackageController struct {
	repo *repositories.ServicePackageRepository
}

func NewServicePackageController(repo *repositories.ServicePackageRepository) *ServicePackageController {
	return &ServicePackageController{repo: repo}
}

// Create handles POST /service_packages
func (h *ServicePackageController) Create(c *fiber.Ctx) error {
	var in models.CreateServicePackageInput
	if err := c.BodyParser(&in); err != nil {
		return response.Error(c, "invalid body", fiber.StatusBadRequest, nil)
	}

	in.Name = strings.TrimSpace(in.Name)
	in.Duration = strings.TrimSpace(in.Duration)

	if in.Name == "" || in.Duration == "" || in.Price <= 0 {
		return response.Error(c, "name, duration, and a positive price are required", fiber.StatusBadRequest, nil)
	}

	sp := &models.ServicePackage{
		Name:        in.Name,
		Description: strings.TrimSpace(in.Description),
		Price:       in.Price,
		Duration:    in.Duration,
		IsActive:    in.IsActive,
	}

	if err := h.repo.Create(c.Context(), sp); err != nil {
		return response.Error(c, "failed to create service package", fiber.StatusInternalServerError, nil)
	}

	return response.Success(c, sp, "service package created", fiber.StatusCreated)
}

// List handles GET /service_packages with optional ?id, ?q and pagination.
func (h *ServicePackageController) List(c *fiber.Ctx) error {
	// If query id is present, return that single resource
	if id := strings.TrimSpace(c.Query("id")); id != "" {
		sp, err := h.repo.FindByID(c.Context(), id)
		if err != nil {
			return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
		}
		if sp == nil {
			return response.Error(c, "service package not found", fiber.StatusNotFound, nil)
		}
		return response.Success(c, sp, "ok")
	}

	// Otherwise list with pagination + search
	page, limit := response.ParsePageLimit(c)
	q := strings.TrimSpace(c.Query("q"))
	items, total, err := h.repo.FindPaged(c.Context(), page, limit, q)
	if err != nil {
		return response.Error(c, "failed to list service packages", fiber.StatusInternalServerError, nil)
	}

	data := response.ListData[models.ServicePackage]{
		Items: items,
		Page:  page,
		Limit: limit,
		Total: total,
	}
	return response.Success(c, data, "ok")
}

// Update handles PUT /service_packages
func (h *ServicePackageController) Update(c *fiber.Ctx) error {
	var in models.UpdateServicePackageInput
	if err := c.BodyParser(&in); err != nil {
		return response.Error(c, "invalid body", fiber.StatusBadRequest, nil)
	}

	id := strings.TrimSpace(in.ID)
	if id == "" {
		return response.Error(c, "id is required in body", fiber.StatusBadRequest, nil)
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
	if in.Price != nil {
		if *in.Price <= 0 {
			return response.Error(c, "price must be positive", fiber.StatusBadRequest, nil)
		}
		updates = append(updates, bson.E{Key: "price", Value: *in.Price})
	}
	if in.Duration != nil {
		v := strings.TrimSpace(*in.Duration)
		if v == "" {
			return response.Error(c, "duration cannot be empty", fiber.StatusBadRequest, nil)
		}
		updates = append(updates, bson.E{Key: "duration", Value: v})
	}
	if in.IsActive != nil {
		updates = append(updates, bson.E{Key: "is_active", Value: *in.IsActive})
	}

	if len(updates) == 0 {
		return response.Error(c, "no fields to update", fiber.StatusBadRequest, nil)
	}

	sp, err := h.repo.UpdateByID(c.Context(), id, updates)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if sp == nil {
		return response.Error(c, "service package not found", fiber.StatusNotFound, nil)
	}

	return response.Success(c, sp, "service package updated")
}

// Delete handles DELETE /service_packages
func (h *ServicePackageController) Delete(c *fiber.Ctx) error {
	id := strings.TrimSpace(c.Query("id"))
	if id == "" {
		return response.Error(c, "id query param is required", fiber.StatusBadRequest, nil)
	}

	ok, err := h.repo.DeleteByID(c.Context(), id)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if !ok {
		return response.Error(c, "service package not found", fiber.StatusNotFound, nil)
	}

	return response.Success(c, true, "service package deleted")
}
