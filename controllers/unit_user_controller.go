package controllers

import (
	"strings"

	"go-fiber-api/models"
	"go-fiber-api/pkg/response"
	"go-fiber-api/repositories"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type UnitUserController struct {
	repo *repositories.UnitUserRepository
}

func NewUnitUserController(repo *repositories.UnitUserRepository) *UnitUserController {
	return &UnitUserController{repo: repo}
}

func isUnitAdmin(c *fiber.Ctx) bool {
	v := c.Locals("is_admin")
	b, _ := v.(bool)
	return b
}

func unitIDFromLocals(c *fiber.Ctx) (primitive.ObjectID, error) {
	v := c.Locals("unit_id")
	s, _ := v.(string)
	return primitive.ObjectIDFromHex(s)
}

// List supports:
// - GET /unit_users?id=<id> -> detail within unit (admin-only)
// - GET /unit_users -> list (admin-only) with pagination and optional q
func (h *UnitUserController) List(c *fiber.Ctx) error {
	if !isUnitAdmin(c) {
		return response.Error(c, "forbidden", fiber.StatusForbidden, nil)
	}
	unitOID, err := unitIDFromLocals(c)
	if err != nil {
		return response.Error(c, "unauthorized", fiber.StatusUnauthorized, nil)
	}
	if id := strings.TrimSpace(c.Query("id")); id != "" {
		u, err := h.repo.FindByIDWithinUnit(c.Context(), id, unitOID)
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
	items, total, err := h.repo.FindPagedByUnit(c.Context(), unitOID, page, limit, q)
	if err != nil {
		return response.Error(c, "failed to list", fiber.StatusInternalServerError, nil)
	}
	data := response.ListData[models.UnitUser]{
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

// Create: POST /unit_users
func (h *UnitUserController) Create(c *fiber.Ctx) error {
	if !isUnitAdmin(c) {
		return response.Error(c, "forbidden", fiber.StatusForbidden, nil)
	}
	unitOID, err := unitIDFromLocals(c)
	if err != nil {
		return response.Error(c, "unauthorized", fiber.StatusUnauthorized, nil)
	}
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		IsAdmin  bool   `json:"is_admin"`
	}
	if err := c.BodyParser(&in); err != nil {
		return response.Error(c, "invalid body", fiber.StatusBadRequest, nil)
	}
	in.Username = strings.TrimSpace(in.Username)
	in.Password = strings.TrimSpace(in.Password)
	if in.Username == "" || in.Password == "" {
		return response.Error(c, "username and password are required", fiber.StatusBadRequest, nil)
	}
	// hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return response.Error(c, "failed to hash password", fiber.StatusInternalServerError, nil)
	}
	u := &models.UnitUser{
		UnitID:       unitOID,
		Username:     in.Username,
		PasswordHash: string(hash),
		IsAdmin:      in.IsAdmin,
		Name:         strings.TrimSpace(in.Name),
		Email:        strings.TrimSpace(in.Email),
	}
	if err := h.repo.Create(c.Context(), u); err != nil {
		if strings.Contains(err.Error(), "E11000") {
			return response.Error(c, "username already exists", fiber.StatusConflict, nil)
		}
		return response.Error(c, "failed to create", fiber.StatusInternalServerError, nil)
	}
	return response.Success(c, u, "created", fiber.StatusCreated)
}

// Update: PUT /unit_users (id required in body)
func (h *UnitUserController) Update(c *fiber.Ctx) error {
	if !isUnitAdmin(c) {
		return response.Error(c, "forbidden", fiber.StatusForbidden, nil)
	}
	unitOID, err := unitIDFromLocals(c)
	if err != nil {
		return response.Error(c, "unauthorized", fiber.StatusUnauthorized, nil)
	}
	var in struct {
		ID       string  `json:"id"`
		Username *string `json:"username,omitempty"`
		Password *string `json:"password,omitempty"`
		Name     *string `json:"name,omitempty"`
		Email    *string `json:"email,omitempty"`
		IsAdmin  *bool   `json:"is_admin,omitempty"`
	}
	if err := c.BodyParser(&in); err != nil {
		return response.Error(c, "invalid body", fiber.StatusBadRequest, nil)
	}
	id := strings.TrimSpace(in.ID)
	if id == "" {
		return response.Error(c, "id is required in body", fiber.StatusBadRequest, nil)
	}
	updates := bson.D{}
	if in.Username != nil {
		v := strings.TrimSpace(*in.Username)
		if v == "" {
			return response.Error(c, "username cannot be empty", fiber.StatusBadRequest, nil)
		}
		updates = append(updates, bson.E{Key: "username", Value: v})
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
	if in.Name != nil {
		updates = append(updates, bson.E{Key: "name", Value: strings.TrimSpace(*in.Name)})
	}
	if in.Email != nil {
		updates = append(updates, bson.E{Key: "email", Value: strings.TrimSpace(*in.Email)})
	}
	if in.IsAdmin != nil {
		// Prevent demoting the last admin of the unit
		if !*in.IsAdmin {
			existing, err := h.repo.FindByIDWithinUnit(c.Context(), id, unitOID)
			if err != nil {
				return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
			}
			if existing == nil {
				return response.Error(c, "not found", fiber.StatusNotFound, nil)
			}
			if existing.IsAdmin {
				cnt, err := h.repo.AdminCount(c.Context(), unitOID)
				if err != nil {
					return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
				}
				if cnt <= 1 {
					return response.Error(c, "cannot remove admin role from the last admin of the unit", fiber.StatusConflict, nil)
				}
			}
		}
		updates = append(updates, bson.E{Key: "is_admin", Value: *in.IsAdmin})
	}
	if len(updates) == 0 {
		return response.Error(c, "no fields to update", fiber.StatusBadRequest, nil)
	}
	u, err := h.repo.UpdateByIDWithinUnit(c.Context(), id, unitOID, updates)
	if err != nil {
		if strings.Contains(err.Error(), "E11000") {
			return response.Error(c, "username already exists", fiber.StatusConflict, nil)
		}
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if u == nil {
		return response.Error(c, "not found", fiber.StatusNotFound, nil)
	}
	return response.Success(c, u, "updated")
}

// Delete: DELETE /unit_users?id=<id>
func (h *UnitUserController) Delete(c *fiber.Ctx) error {
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
	// Prevent deleting the last admin of the unit
	existing, err := h.repo.FindByIDWithinUnit(c.Context(), id, unitOID)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if existing == nil {
		return response.Error(c, "not found", fiber.StatusNotFound, nil)
	}
	if existing.IsAdmin {
		cnt, err := h.repo.AdminCount(c.Context(), unitOID)
		if err != nil {
			return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
		}
		if cnt <= 1 {
			return response.Error(c, "cannot delete the last admin of the unit", fiber.StatusConflict, nil)
		}
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
