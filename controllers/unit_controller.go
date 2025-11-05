package controllers

import (
    "regexp"
    "strings"

    "go-fiber-api/models"
    "go-fiber-api/pkg/response"
    "go-fiber-api/repositories"

    "github.com/gofiber/fiber/v2"
    "go.mongodb.org/mongo-driver/bson"
)

type UnitController struct {
    repo *repositories.UnitRepository
}

func NewUnitController(repo *repositories.UnitRepository) *UnitController {
    return &UnitController{repo: repo}
}

// List handles GET /units with optional ?id, ?q and pagination.
func (h *UnitController) List(c *fiber.Ctx) error {
    if id := strings.TrimSpace(c.Query("id")); id != "" {
        u, err := h.repo.FindByID(c.Context(), id)
        if err != nil {
            return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
        }
        if u == nil {
            return response.Error(c, "not found", fiber.StatusNotFound, nil)
        }
        return response.Success(c, u, "ok")
    }
    page, limit := response.ParsePageLimit(c)
    q := strings.TrimSpace(c.Query("q"))
    items, total, err := h.repo.FindPaged(c.Context(), page, limit, q)
    if err != nil {
        return response.Error(c, "failed to list", fiber.StatusInternalServerError, nil)
    }
    data := response.ListData[models.Unit]{
        Items: items,
        Page:  page,
        Limit: limit,
        Total: total,
    }
    return response.Success(c, data, "ok")
}

func (h *UnitController) Create(c *fiber.Ctx) error {
    var in models.CreateUnitInput
    if err := c.BodyParser(&in); err != nil {
        return response.Error(c, "invalid body", fiber.StatusBadRequest, nil)
    }
    in.Subdomain = strings.ToLower(strings.TrimSpace(in.Subdomain))
    in.Name = strings.TrimSpace(in.Name)
    if in.Subdomain == "" || in.Name == "" {
        return response.Error(c, "subdomain and name are required", fiber.StatusBadRequest, nil)
    }
    // Basic subdomain validation: letters, numbers, dashes only
    if !regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`).MatchString(in.Subdomain) {
        return response.Error(c, "invalid subdomain format", fiber.StatusBadRequest, nil)
    }
    u := &models.Unit{
        Subdomain:   in.Subdomain,
        Name:        in.Name,
        Description: strings.TrimSpace(in.Description),
        LogoURL:     strings.TrimSpace(in.LogoURL),
    }
    if err := h.repo.Create(c.Context(), u); err != nil {
        if strings.Contains(err.Error(), "E11000") {
            return response.Error(c, "subdomain already exists", fiber.StatusConflict, nil)
        }
        return response.Error(c, "failed to create", fiber.StatusInternalServerError, nil)
    }
    return response.Success(c, u, "created", fiber.StatusCreated)
}

func (h *UnitController) Update(c *fiber.Ctx) error {
    var in models.UpdateUnitInput
    if err := c.BodyParser(&in); err != nil {
        return response.Error(c, "invalid body", fiber.StatusBadRequest, nil)
    }
    id := strings.TrimSpace(in.ID)
    if id == "" {
        return response.Error(c, "id is required in body", fiber.StatusBadRequest, nil)
    }
    updates := bson.D{}
    if in.Subdomain != nil {
        v := strings.ToLower(strings.TrimSpace(*in.Subdomain))
        if v == "" {
            return response.Error(c, "subdomain cannot be empty", fiber.StatusBadRequest, nil)
        }
        if !regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`).MatchString(v) {
            return response.Error(c, "invalid subdomain format", fiber.StatusBadRequest, nil)
        }
        updates = append(updates, bson.E{Key: "subdomain", Value: v})
    }
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
    u, err := h.repo.UpdateByID(c.Context(), id, updates)
    if err != nil {
        if strings.Contains(err.Error(), "E11000") {
            return response.Error(c, "subdomain already exists", fiber.StatusConflict, nil)
        }
        return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
    }
    if u == nil {
        return response.Error(c, "not found", fiber.StatusNotFound, nil)
    }
    return response.Success(c, u, "updated")
}

func (h *UnitController) Delete(c *fiber.Ctx) error {
    id := strings.TrimSpace(c.Query("id"))
    if id == "" {
        return response.Error(c, "id query param is required", fiber.StatusBadRequest, nil)
    }
    ok, err := h.repo.DeleteByID(c.Context(), id)
    if err != nil {
        return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
    }
    if !ok {
        return response.Error(c, "not found", fiber.StatusNotFound, nil)
    }
    return response.Success(c, true, "deleted")
}
