package controllers

import (
	"strings"

	"go-fiber-api/models"
	"go-fiber-api/pkg/response"
	"go-fiber-api/repositories"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ServicePackageMenuController manages menu items of a service package (admin-only).
type ServicePackageMenuController struct {
	repo *repositories.ServicePackageRepository
}

func NewServicePackageMenuController(repo *repositories.ServicePackageRepository) *ServicePackageMenuController {
	return &ServicePackageMenuController{repo: repo}
}

// List returns menus for a service package or a single menu when id is provided.
// GET /service_package_menus?service_package_id=<id>[&id=&page=&limit=&q=]
func (h *ServicePackageMenuController) List(c *fiber.Ctx) error {
	spID := strings.TrimSpace(c.Query("service_package_id"))
	if spID == "" {
		return response.Error(c, "service_package_id query param is required", fiber.StatusBadRequest, nil)
	}
	if _, err := primitive.ObjectIDFromHex(spID); err != nil {
		return response.Error(c, "invalid service_package_id", fiber.StatusBadRequest, nil)
	}

	sp, err := h.repo.FindByID(c.Context(), spID)
	if err != nil {
		return response.Error(c, "failed to load service package", fiber.StatusInternalServerError, nil)
	}
	if sp == nil {
		return response.Error(c, "service package not found", fiber.StatusNotFound, nil)
	}

	// Backfill missing menu IDs for existing data.
	if ensureMenuIDs(sp.Menus) {
		updated, err := h.repo.UpdateByID(c.Context(), spID, bson.D{{Key: "menus", Value: sp.Menus}})
		if err != nil {
			return response.Error(c, "failed to normalize service package menus", fiber.StatusInternalServerError, nil)
		}
		if updated != nil {
			sp = updated
		}
	}

	// Detail request when menu id is provided
	if id := strings.TrimSpace(c.Query("id")); id != "" {
		menuID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return response.Error(c, "invalid id", fiber.StatusBadRequest, nil)
		}
		for _, menu := range sp.Menus {
			if menu.ID == menuID {
				return response.Success(c, menu, "ok")
			}
		}
		return response.Error(c, "menu not found", fiber.StatusNotFound, nil)
	}

	page, limit := response.ParsePageLimit(c)
	q := strings.ToLower(strings.TrimSpace(c.Query("q")))

	var filtered []models.ServicePackageMenu
	for _, menu := range sp.Menus {
		if q != "" {
			if !strings.Contains(strings.ToLower(menu.Title), q) &&
				!strings.Contains(strings.ToLower(menu.Key), q) &&
				!strings.Contains(strings.ToLower(menu.URL), q) {
				continue
			}
		}
		filtered = append(filtered, menu)
	}

	total := int64(len(filtered))
	if page == 0 {
		data := response.ListData[models.ServicePackageMenu]{
			Items: filtered,
			Page:  0,
			Limit: total,
			Total: total,
		}
		return response.Success(c, data, "ok")
	}

	if limit < 1 {
		limit = 10
	}
	start := (page - 1) * limit
	if start >= total {
		data := response.ListData[models.ServicePackageMenu]{
			Items: []models.ServicePackageMenu{},
			Page:  page,
			Limit: limit,
			Total: total,
		}
		return response.Success(c, data, "ok")
	}
	end := start + limit
	if end > total {
		end = total
	}
	startIdx := int(start)
	endIdx := int(end)
	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx > len(filtered) {
		endIdx = len(filtered)
	}
	paged := filtered[startIdx:endIdx]

	data := response.ListData[models.ServicePackageMenu]{
		Items: paged,
		Page:  page,
		Limit: limit,
		Total: total,
	}
	return response.Success(c, data, "ok")
}

// Create adds a new menu to a service package.
// POST /service_package_menus
func (h *ServicePackageMenuController) Create(c *fiber.Ctx) error {
	var in models.CreateServicePackageMenuInput
	if err := c.BodyParser(&in); err != nil {
		return response.Error(c, "invalid body", fiber.StatusBadRequest, nil)
	}
	in.ServicePackageID = strings.TrimSpace(in.ServicePackageID)
	in.Title = strings.TrimSpace(in.Title)
	in.Key = strings.TrimSpace(in.Key)
	in.URL = strings.TrimSpace(in.URL)
	in.Icon = strings.TrimSpace(in.Icon)

	if in.ServicePackageID == "" {
		return response.Error(c, "service_package_id is required", fiber.StatusBadRequest, nil)
	}
	if _, err := primitive.ObjectIDFromHex(in.ServicePackageID); err != nil {
		return response.Error(c, "invalid service_package_id", fiber.StatusBadRequest, nil)
	}
	if in.Title == "" || in.Key == "" || in.URL == "" {
		return response.Error(c, "title, key, and url are required", fiber.StatusBadRequest, nil)
	}

	sp, err := h.repo.FindByID(c.Context(), in.ServicePackageID)
	if err != nil {
		return response.Error(c, "failed to load service package", fiber.StatusInternalServerError, nil)
	}
	if sp == nil {
		return response.Error(c, "service package not found", fiber.StatusNotFound, nil)
	}

	ensureMenuIDs(sp.Menus)

	for _, menu := range sp.Menus {
		if strings.EqualFold(strings.TrimSpace(menu.Key), in.Key) {
			return response.Error(c, "menu key already exists in this service package", fiber.StatusBadRequest, nil)
		}
	}
	parentID, err := parseServicePackageParentID(in.ParentID)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}

	menu := models.ServicePackageMenu{
		ID:         primitive.NewObjectID(),
		Title:      in.Title,
		Key:        in.Key,
		URL:        in.URL,
		Icon:       in.Icon,
		ParentID:   parentID,
		Permission: in.Permission,
		Active:     in.Active,
	}
	sp.Menus = append(sp.Menus, menu)

	updated, err := h.repo.UpdateByID(c.Context(), in.ServicePackageID, bson.D{{Key: "menus", Value: sp.Menus}})
	if err != nil {
		return response.Error(c, "failed to add menu", fiber.StatusInternalServerError, nil)
	}
	if updated == nil {
		return response.Error(c, "service package not found", fiber.StatusNotFound, nil)
	}

	return response.Success(c, menu, "menu added", fiber.StatusCreated)
}

// Update modifies a menu of a service package.
// PUT /service_package_menus
func (h *ServicePackageMenuController) Update(c *fiber.Ctx) error {
	var in models.UpdateServicePackageMenuInput
	if err := c.BodyParser(&in); err != nil {
		return response.Error(c, "invalid body", fiber.StatusBadRequest, nil)
	}
	in.ID = strings.TrimSpace(in.ID)
	in.ServicePackageID = strings.TrimSpace(in.ServicePackageID)

	if in.ID == "" || in.ServicePackageID == "" {
		return response.Error(c, "id and service_package_id are required", fiber.StatusBadRequest, nil)
	}
	menuID, err := primitive.ObjectIDFromHex(in.ID)
	if err != nil {
		return response.Error(c, "invalid id", fiber.StatusBadRequest, nil)
	}
	if _, err := primitive.ObjectIDFromHex(in.ServicePackageID); err != nil {
		return response.Error(c, "invalid service_package_id", fiber.StatusBadRequest, nil)
	}

	if in.Title == nil && in.Key == nil && in.URL == nil && in.Icon == nil && in.ParentID == nil && in.Permission == nil && in.Active == nil {
		return response.Error(c, "no fields to update", fiber.StatusBadRequest, nil)
	}

	sp, err := h.repo.FindByID(c.Context(), in.ServicePackageID)
	if err != nil {
		return response.Error(c, "failed to load service package", fiber.StatusInternalServerError, nil)
	}
	if sp == nil {
		return response.Error(c, "service package not found", fiber.StatusNotFound, nil)
	}

	ensureMenuIDs(sp.Menus)

	idx := -1
	for i, menu := range sp.Menus {
		if menu.ID == menuID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return response.Error(c, "menu not found", fiber.StatusNotFound, nil)
	}

	menu := sp.Menus[idx]
	updatedAny := false

	if in.Title != nil {
		v := strings.TrimSpace(*in.Title)
		if v == "" {
			return response.Error(c, "title cannot be empty", fiber.StatusBadRequest, nil)
		}
		menu.Title = v
		updatedAny = true
	}
	if in.Key != nil {
		v := strings.TrimSpace(*in.Key)
		if v == "" {
			return response.Error(c, "key cannot be empty", fiber.StatusBadRequest, nil)
		}
		for i, m := range sp.Menus {
			if i == idx {
				continue
			}
			if strings.EqualFold(strings.TrimSpace(m.Key), v) {
				return response.Error(c, "menu key already exists in this service package", fiber.StatusBadRequest, nil)
			}
		}
		menu.Key = v
		updatedAny = true
	}
	if in.URL != nil {
		v := strings.TrimSpace(*in.URL)
		if v == "" {
			return response.Error(c, "url cannot be empty", fiber.StatusBadRequest, nil)
		}
		menu.URL = v
		updatedAny = true
	}
	if in.Icon != nil {
		menu.Icon = strings.TrimSpace(*in.Icon)
		updatedAny = true
	}
	if in.ParentID != nil {
		parentID, err := parseServicePackageParentID(*in.ParentID)
		if err != nil {
			return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
		}
		menu.ParentID = parentID
		updatedAny = true
	}
	if in.Permission != nil {
		menu.Permission = *in.Permission
		updatedAny = true
	}
	if in.Active != nil {
		menu.Active = *in.Active
		updatedAny = true
	}

	if !updatedAny {
		return response.Error(c, "no fields to update", fiber.StatusBadRequest, nil)
	}

	sp.Menus[idx] = menu
	updated, err := h.repo.UpdateByID(c.Context(), in.ServicePackageID, bson.D{{Key: "menus", Value: sp.Menus}})
	if err != nil {
		return response.Error(c, "failed to update menu", fiber.StatusInternalServerError, nil)
	}
	if updated == nil {
		return response.Error(c, "service package not found", fiber.StatusNotFound, nil)
	}

	return response.Success(c, menu, "menu updated")
}

// Delete removes a menu from a service package.
// DELETE /service_package_menus?id=<menuId>&service_package_id=<packageId>
func (h *ServicePackageMenuController) Delete(c *fiber.Ctx) error {
	id := strings.TrimSpace(c.Query("id"))
	spID := strings.TrimSpace(c.Query("service_package_id"))
	if id == "" || spID == "" {
		return response.Error(c, "id and service_package_id are required", fiber.StatusBadRequest, nil)
	}
	menuID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return response.Error(c, "invalid id", fiber.StatusBadRequest, nil)
	}
	if _, err := primitive.ObjectIDFromHex(spID); err != nil {
		return response.Error(c, "invalid service_package_id", fiber.StatusBadRequest, nil)
	}

	sp, err := h.repo.FindByID(c.Context(), spID)
	if err != nil {
		return response.Error(c, "failed to load service package", fiber.StatusInternalServerError, nil)
	}
	if sp == nil {
		return response.Error(c, "service package not found", fiber.StatusNotFound, nil)
	}

	ensureMenuIDs(sp.Menus)

	idx := -1
	for i, menu := range sp.Menus {
		if menu.ID == menuID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return response.Error(c, "menu not found", fiber.StatusNotFound, nil)
	}

	sp.Menus = append(sp.Menus[:idx], sp.Menus[idx+1:]...)

	updated, err := h.repo.UpdateByID(c.Context(), spID, bson.D{{Key: "menus", Value: sp.Menus}})
	if err != nil {
		return response.Error(c, "failed to delete menu", fiber.StatusInternalServerError, nil)
	}
	if updated == nil {
		return response.Error(c, "service package not found", fiber.StatusNotFound, nil)
	}

	return response.Success(c, true, "menu deleted")
}
