package controllers

import (
	"strings"

	"go-fiber-api/models"
	"go-fiber-api/pkg/response"
	"go-fiber-api/repositories"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// UnitRoleGroupController handles CRUD for unit-scoped role groups.
type UnitRoleGroupController struct {
	repo     *repositories.UnitRoleGroupRepository
	userRepo *repositories.UnitUserRepository
}

func NewUnitRoleGroupController(repo *repositories.UnitRoleGroupRepository, userRepo *repositories.UnitUserRepository) *UnitRoleGroupController {
	return &UnitRoleGroupController{repo: repo, userRepo: userRepo}
}

// List supports GET /unit_role_groups with optional ?id and search/pagination.
func (h *UnitRoleGroupController) List(c *fiber.Ctx) error {
	if !isUnitAdmin(c) {
		return response.Error(c, "forbidden", fiber.StatusForbidden, nil)
	}
	unitOID, err := unitIDFromLocals(c)
	if err != nil {
		return response.Error(c, "unauthorized", fiber.StatusUnauthorized, nil)
	}
	if id := strings.TrimSpace(c.Query("id")); id != "" {
		g, err := h.repo.FindByIDWithinUnit(c.Context(), id, unitOID)
		if err != nil {
			return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
		}
		if g == nil {
			return response.Error(c, "not found", fiber.StatusNotFound, nil)
		}
		return response.Success(c, g, "ok")
	}
	page, limit := response.ParsePageLimit(c)
	q := strings.TrimSpace(c.Query("q"))
	items, total, err := h.repo.FindPagedByUnit(c.Context(), unitOID, page, limit, q)
	if err != nil {
		return response.Error(c, "failed to list", fiber.StatusInternalServerError, nil)
	}
	data := response.ListData[models.UnitRoleGroup]{
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

// Create handles POST /unit_role_groups.
func (h *UnitRoleGroupController) Create(c *fiber.Ctx) error {
	if !isUnitAdmin(c) {
		return response.Error(c, "forbidden", fiber.StatusForbidden, nil)
	}
	unitOID, err := unitIDFromLocals(c)
	if err != nil {
		return response.Error(c, "unauthorized", fiber.StatusUnauthorized, nil)
	}
	var in models.CreateUnitRoleGroupInput
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
	g := &models.UnitRoleGroup{
		UnitID:      unitOID,
		Name:        in.Name,
		Description: in.Description,
		Permissions: in.Permissions,
	}
	if err := h.repo.Create(c.Context(), g); err != nil {
		if strings.Contains(err.Error(), "E11000") {
			return response.Error(c, "role group name already exists in this unit", fiber.StatusConflict, nil)
		}
		return response.Error(c, "failed to create", fiber.StatusInternalServerError, nil)
	}
	return response.Success(c, g, "created", fiber.StatusCreated)
}

// Update handles PUT /unit_role_groups.
func (h *UnitRoleGroupController) Update(c *fiber.Ctx) error {
	if !isUnitAdmin(c) {
		return response.Error(c, "forbidden", fiber.StatusForbidden, nil)
	}
	unitOID, err := unitIDFromLocals(c)
	if err != nil {
		return response.Error(c, "unauthorized", fiber.StatusUnauthorized, nil)
	}
	var in models.UpdateUnitRoleGroupInput
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
	g, err := h.repo.UpdateByIDWithinUnit(c.Context(), id, unitOID, updates)
	if err != nil {
		if strings.Contains(err.Error(), "E11000") {
			return response.Error(c, "role group name already exists in this unit", fiber.StatusConflict, nil)
		}
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if g == nil {
		return response.Error(c, "not found", fiber.StatusNotFound, nil)
	}
	return response.Success(c, g, "updated")
}

// Delete handles DELETE /unit_role_groups?id=...
func (h *UnitRoleGroupController) Delete(c *fiber.Ctx) error {
	if !isUnitAdmin(c) {
		return response.Error(c, "forbidden", fiber.StatusForbidden, nil)
	}
	unitOID, err := unitIDFromLocals(c)
	if err != nil {
		return response.Error(c, "unauthorized", fiber.StatusUnauthorized, nil)
	}
	id := strings.TrimSpace(c.Query("id"))
	if id == "" {
		return response.Error(c, "id query param is required", fiber.StatusBadRequest, nil)
	}
	g, err := h.repo.FindByIDWithinUnit(c.Context(), id, unitOID)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if g == nil {
		return response.Error(c, "not found", fiber.StatusNotFound, nil)
	}
	if err := h.userRepo.PullRoleGroupFromAll(c.Context(), unitOID, g.ID); err != nil {
		return response.Error(c, "failed to detach role group from users", fiber.StatusInternalServerError, nil)
	}
	ok, err := h.repo.DeleteByIDWithinUnit(c.Context(), id, unitOID)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if !ok {
		return response.Error(c, "not found", fiber.StatusNotFound, nil)
	}
	return response.Success(c, true, "deleted")
}
