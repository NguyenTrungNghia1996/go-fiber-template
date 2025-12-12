package controllers

import (
	"strings"

	"go-fiber-api/models"
	"go-fiber-api/pkg/response"
	"go-fiber-api/repositories"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// SuperAdminMenuController exposes CRUD endpoints for super admin menus.
type SuperAdminMenuController struct {
	repo *repositories.SuperAdminMenuRepository
}

func NewSuperAdminMenuController(repo *repositories.SuperAdminMenuRepository) *SuperAdminMenuController {
	return &SuperAdminMenuController{repo: repo}
}

// List handles GET /superadmin_menus with optional id/q pagination.
func (h *SuperAdminMenuController) List(c *fiber.Ctx) error {
	if id := strings.TrimSpace(c.Query("id")); id != "" {
		menu, err := h.repo.FindByID(c.Context(), id)
		if err != nil {
			return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
		}
		if menu == nil {
			return response.Error(c, "menu not found", fiber.StatusNotFound, nil)
		}
		return response.Success(c, menu, "ok")
	}

	page, limit := response.ParsePageLimit(c)
	q := strings.TrimSpace(c.Query("q"))
	items, total, err := h.repo.FindPaged(c.Context(), page, limit, q)
	if err != nil {
		return response.Error(c, "failed to list menus", fiber.StatusInternalServerError, nil)
	}

	data := response.ListData[models.SuperAdminMenu]{
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

// Create handles POST /superadmin_menus to add a menu item.
func (h *SuperAdminMenuController) Create(c *fiber.Ctx) error {
	var in models.CreateSuperAdminMenuInput
	if err := c.BodyParser(&in); err != nil {
		return response.Error(c, "invalid body", fiber.StatusBadRequest, nil)
	}
	in.Title = strings.TrimSpace(in.Title)
	in.Key = strings.TrimSpace(in.Key)
	in.URL = strings.TrimSpace(in.URL)
	in.Icon = strings.TrimSpace(in.Icon)

	if in.Title == "" || in.Key == "" || in.URL == "" {
		return response.Error(c, "title, key, and url are required", fiber.StatusBadRequest, nil)
	}

	if existing, err := h.repo.FindByKey(c.Context(), in.Key); err != nil {
		return response.Error(c, "failed to check existing menus", fiber.StatusInternalServerError, nil)
	} else if existing != nil {
		return response.Error(c, "menu key already exists", fiber.StatusBadRequest, nil)
	}

	menu := &models.SuperAdminMenu{
		Title:      in.Title,
		Key:        in.Key,
		URL:        in.URL,
		Icon:       in.Icon,
		ParentID:   in.ParentID,
		Permission: in.Permission,
		Active:     in.Active,
	}

	if err := h.repo.Create(c.Context(), menu); err != nil {
		return response.Error(c, "failed to create menu", fiber.StatusInternalServerError, nil)
	}

	return response.Success(c, menu, "menu created", fiber.StatusCreated)
}

// Update handles PUT /superadmin_menus to modify a menu item.
func (h *SuperAdminMenuController) Update(c *fiber.Ctx) error {
	var in models.UpdateSuperAdminMenuInput
	if err := c.BodyParser(&in); err != nil {
		return response.Error(c, "invalid body", fiber.StatusBadRequest, nil)
	}
	id := strings.TrimSpace(in.ID)
	if id == "" {
		return response.Error(c, "id is required in body", fiber.StatusBadRequest, nil)
	}

	updates := bson.D{}
	if in.Title != nil {
		v := strings.TrimSpace(*in.Title)
		if v == "" {
			return response.Error(c, "title cannot be empty", fiber.StatusBadRequest, nil)
		}
		updates = append(updates, bson.E{Key: "title", Value: v})
	}
	if in.Key != nil {
		v := strings.TrimSpace(*in.Key)
		if v == "" {
			return response.Error(c, "key cannot be empty", fiber.StatusBadRequest, nil)
		}
		if existing, err := h.repo.FindByKey(c.Context(), v); err != nil {
			return response.Error(c, "failed to check existing menus", fiber.StatusInternalServerError, nil)
		} else if existing != nil && existing.ID.Hex() != id {
			return response.Error(c, "menu key already exists", fiber.StatusBadRequest, nil)
		}
		updates = append(updates, bson.E{Key: "key", Value: v})
	}
	if in.URL != nil {
		v := strings.TrimSpace(*in.URL)
		if v == "" {
			return response.Error(c, "url cannot be empty", fiber.StatusBadRequest, nil)
		}
		updates = append(updates, bson.E{Key: "url", Value: v})
	}
	if in.Icon != nil {
		updates = append(updates, bson.E{Key: "icon", Value: strings.TrimSpace(*in.Icon)})
	}
	if in.ParentID != nil {
		updates = append(updates, bson.E{Key: "parent_id", Value: *in.ParentID})
	}
	if in.Permission != nil {
		updates = append(updates, bson.E{Key: "permission", Value: *in.Permission})
	}
	if in.Active != nil {
		updates = append(updates, bson.E{Key: "active", Value: *in.Active})
	}

	if len(updates) == 0 {
		return response.Error(c, "no fields to update", fiber.StatusBadRequest, nil)
	}

	menu, err := h.repo.UpdateByID(c.Context(), id, updates)
	if err != nil {
		return response.Error(c, "failed to update menu", fiber.StatusInternalServerError, nil)
	}
	if menu == nil {
		return response.Error(c, "menu not found", fiber.StatusNotFound, nil)
	}

	return response.Success(c, menu, "menu updated")
}

// Delete handles DELETE /superadmin_menus?id=...
func (h *SuperAdminMenuController) Delete(c *fiber.Ctx) error {
	id := strings.TrimSpace(c.Query("id"))
	if id == "" {
		return response.Error(c, "id query param is required", fiber.StatusBadRequest, nil)
	}

	ok, err := h.repo.DeleteByID(c.Context(), id)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if !ok {
		return response.Error(c, "menu not found", fiber.StatusNotFound, nil)
	}

	return response.Success(c, true, "menu deleted")
}
