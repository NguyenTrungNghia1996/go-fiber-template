package controllers

import (
	"errors"
	"strings"

	"go-fiber-api/models"
	"go-fiber-api/pkg/response"
	"go-fiber-api/repositories"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SuperAdminRoleGroupController handles CRUD operations for super admin role groups.
type SuperAdminRoleGroupController struct {
	repo           *repositories.SuperAdminRoleGroupRepository
	superAdminRepo *repositories.SuperAdminRepository
}

func NewSuperAdminRoleGroupController(repo *repositories.SuperAdminRoleGroupRepository, superAdminRepo *repositories.SuperAdminRepository) *SuperAdminRoleGroupController {
	return &SuperAdminRoleGroupController{
		repo:           repo,
		superAdminRepo: superAdminRepo,
	}
}

// List supports GET /superadmin_role_groups with ?id or pagination + search.
func (h *SuperAdminRoleGroupController) List(c *fiber.Ctx) error {
	if id := strings.TrimSpace(c.Query("id")); id != "" {
		g, err := h.repo.FindByID(c.Context(), id)
		if err != nil {
			return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
		}
		if g == nil {
			return response.Error(c, "role group not found", fiber.StatusNotFound, nil)
		}
		return response.Success(c, g, "ok")
	}
	page, limit := response.ParsePageLimit(c)
	q := strings.TrimSpace(c.Query("q"))
	items, total, err := h.repo.FindPaged(c.Context(), page, limit, q)
	if err != nil {
		return response.Error(c, "failed to list role groups", fiber.StatusInternalServerError, nil)
	}
	data := response.ListData[models.SuperAdminRoleGroup]{
		Items: items,
		Page:  page,
		Limit: limit,
		Total: total,
	}
	if page == 0 {
		data.Limit = total
	}
	return response.Success(c, data, "ok")
}

// Create handles POST /superadmin_role_groups.
func (h *SuperAdminRoleGroupController) Create(c *fiber.Ctx) error {
	var in models.CreateSuperAdminRoleGroupInput
	if err := c.BodyParser(&in); err != nil {
		return response.Error(c, "invalid body", fiber.StatusBadRequest, nil)
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	if in.Name == "" {
		return response.Error(c, "name is required", fiber.StatusBadRequest, nil)
	}
	if err := validatePermissions(in.Permissions); err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	g := &models.SuperAdminRoleGroup{
		Name:        in.Name,
		Description: in.Description,
		Permissions: in.Permissions,
	}
	if err := h.repo.Create(c.Context(), g); err != nil {
		if strings.Contains(err.Error(), "E11000") {
			return response.Error(c, "role group name already exists", fiber.StatusConflict, nil)
		}
		return response.Error(c, "failed to create role group", fiber.StatusInternalServerError, nil)
	}
	return response.Success(c, g, "created", fiber.StatusCreated)
}

// Update handles PUT /superadmin_role_groups.
func (h *SuperAdminRoleGroupController) Update(c *fiber.Ctx) error {
	var in models.UpdateSuperAdminRoleGroupInput
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
	if in.Permissions != nil {
		if err := validatePermissions(*in.Permissions); err != nil {
			return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
		}
		updates = append(updates, bson.E{Key: "permissions", Value: *in.Permissions})
	}
	if len(updates) == 0 {
		return response.Error(c, "no fields to update", fiber.StatusBadRequest, nil)
	}
	g, err := h.repo.UpdateByID(c.Context(), id, updates)
	if err != nil {
		if strings.Contains(err.Error(), "E11000") {
			return response.Error(c, "role group name already exists", fiber.StatusConflict, nil)
		}
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if g == nil {
		return response.Error(c, "role group not found", fiber.StatusNotFound, nil)
	}
	return response.Success(c, g, "updated")
}

// Delete handles DELETE /superadmin_role_groups?id=...
func (h *SuperAdminRoleGroupController) Delete(c *fiber.Ctx) error {
	id := strings.TrimSpace(c.Query("id"))
	if id == "" {
		return response.Error(c, "id query param is required", fiber.StatusBadRequest, nil)
	}
	g, err := h.repo.FindByID(c.Context(), id)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if g == nil {
		return response.Error(c, "role group not found", fiber.StatusNotFound, nil)
	}
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return response.Error(c, "invalid id", fiber.StatusBadRequest, nil)
	}
	if err := h.superAdminRepo.PullRoleGroupFromAll(c.Context(), oid); err != nil {
		return response.Error(c, "failed to detach role group from super admins", fiber.StatusInternalServerError, nil)
	}
	ok, err := h.repo.DeleteByID(c.Context(), id)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if !ok {
		return response.Error(c, "role group not found", fiber.StatusNotFound, nil)
	}
	return response.Success(c, true, "deleted")
}

func validatePermissions(perms []models.SuperAdminMenuPermission) error {
	seen := make(map[string]struct{})
	for _, p := range perms {
		key := strings.TrimSpace(p.Key)
		if key == "" {
			return errors.New("permissions cannot contain empty key")
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
	}
	return nil
}
