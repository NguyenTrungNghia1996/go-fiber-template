package controllers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-fiber-api/models"
	"go-fiber-api/pkg/auth"
	"go-fiber-api/pkg/response"
	"go-fiber-api/repositories"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type SuperAdminController struct {
	repo          *repositories.SuperAdminRepository
	roleGroupRepo *repositories.SuperAdminRoleGroupRepository
}

func NewSuperAdminController(repo *repositories.SuperAdminRepository, roleGroupRepo *repositories.SuperAdminRoleGroupRepository) *SuperAdminController {
	return &SuperAdminController{repo: repo, roleGroupRepo: roleGroupRepo}
}

func (h *SuperAdminController) Create(c *fiber.Ctx) error {
	var in models.CreateSuperAdminInput
	if err := c.BodyParser(&in); err != nil {
		return response.Error(c, "invalid body", fiber.StatusBadRequest, nil)
	}
	in.Username = strings.TrimSpace(in.Username)
	in.Password = strings.TrimSpace(in.Password)
	if in.Username == "" || in.Password == "" {
		return response.Error(c, "username and password are required", fiber.StatusBadRequest, nil)
	}
	roleGroupIDs, err := h.parseRoleGroupIDs(c.Context(), in.RoleGroupIDs)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	// Check existing username
	if exists, _ := h.repo.FindByUsername(c.Context(), in.Username); exists != nil {
		return response.Error(c, "username already exists", fiber.StatusConflict, nil)
	}
	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return response.Error(c, "failed to hash password", fiber.StatusInternalServerError, nil)
	}
	sa := &models.SuperAdmin{
		Username:     in.Username,
		PasswordHash: string(hash),
		Name:         strings.TrimSpace(in.Name),
		Email:        strings.TrimSpace(in.Email),
		RoleGroupIDs: roleGroupIDs,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	if err := h.repo.Create(c.Context(), sa); err != nil {
		// handle duplicate key
		if strings.Contains(err.Error(), "E11000") {
			return response.Error(c, "username already exists", fiber.StatusConflict, nil)
		}
		return response.Error(c, "failed to create", fiber.StatusInternalServerError, nil)
	}
	return response.Success(c, sa, "created", fiber.StatusCreated)
}

func (h *SuperAdminController) List(c *fiber.Ctx) error {
	// If query id is present, return that single resource
	if id := strings.TrimSpace(c.Query("id")); id != "" {
		sa, err := h.repo.FindByID(c.Context(), id)
		if err != nil {
			return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
		}
		if sa == nil {
			return response.Error(c, "not found", fiber.StatusNotFound, nil)
		}
		return response.Success(c, sa, "ok")
	}
	// Otherwise list with pagination + search
	page, limit := response.ParsePageLimit(c)
	q := strings.TrimSpace(c.Query("q"))
	items, total, err := h.repo.FindPaged(c.Context(), page, limit, q)
	if err != nil {
		return response.Error(c, "failed to list", fiber.StatusInternalServerError, nil)
	}
	data := response.ListData[models.SuperAdmin]{
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

func (h *SuperAdminController) Update(c *fiber.Ctx) error {
	var in models.UpdateSuperAdminInput
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
		updates = append(updates, bson.E{Key: "name", Value: v})
	}
	if in.Email != nil {
		v := strings.TrimSpace(*in.Email)
		updates = append(updates, bson.E{Key: "email", Value: v})
	}
	if in.Password != nil {
		if strings.TrimSpace(*in.Password) == "" {
			return response.Error(c, "password cannot be empty", fiber.StatusBadRequest, nil)
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(*in.Password), bcrypt.DefaultCost)
		if err != nil {
			return response.Error(c, "failed to hash password", fiber.StatusInternalServerError, nil)
		}
		updates = append(updates, bson.E{Key: "password_hash", Value: string(hash)})
	}
	if in.RoleGroupIDs != nil {
		roleGroupIDs, err := h.parseRoleGroupIDs(c.Context(), *in.RoleGroupIDs)
		if err != nil {
			return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
		}
		updates = append(updates, bson.E{Key: "role_group_ids", Value: roleGroupIDs})
	}
	if len(updates) == 0 {
		return response.Error(c, "no fields to update", fiber.StatusBadRequest, nil)
	}
	updated, err := h.repo.UpdateByID(c.Context(), id, updates)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return response.Error(c, "timeout", fiber.StatusRequestTimeout, nil)
		}
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if updated == nil {
		return response.Error(c, "not found", fiber.StatusNotFound, nil)
	}
	return response.Success(c, updated, "updated")
}

func (h *SuperAdminController) Delete(c *fiber.Ctx) error {
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

// Login authenticates a super admin and returns a JWT token.
func (h *SuperAdminController) Login(c *fiber.Ctx) error {
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&in); err != nil {
		return response.Error(c, "invalid body", fiber.StatusBadRequest, nil)
	}
	in.Username = strings.TrimSpace(in.Username)
	in.Password = strings.TrimSpace(in.Password)
	if in.Username == "" || in.Password == "" {
		return response.Error(c, "username and password are required", fiber.StatusBadRequest, nil)
	}
	sa, err := h.repo.FindByUsername(c.Context(), in.Username)
	if err != nil {
		return response.Error(c, "authentication failed", fiber.StatusUnauthorized, nil)
	}
	if sa == nil {
		return response.Error(c, "authentication failed", fiber.StatusUnauthorized, nil)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(sa.PasswordHash), []byte(in.Password)); err != nil {
		return response.Error(c, "authentication failed", fiber.StatusUnauthorized, nil)
	}
	token, exp, err := auth.GenerateToken(sa.ID.Hex(), 24*time.Hour)
	if err != nil {
		return response.Error(c, "failed to issue token", fiber.StatusInternalServerError, nil)
	}
	// minimal user payload
	user := fiber.Map{
		"id":             sa.ID.Hex(),
		"username":       sa.Username,
		"name":           sa.Name,
		"email":          sa.Email,
		"role_group_ids": sa.RoleGroupIDs,
	}
	return response.Success(c, fiber.Map{
		"token":      token,
		"expires_at": exp.Format(time.RFC3339),
		"user":       user,
	}, "logged in")
}

func (h *SuperAdminController) parseRoleGroupIDs(ctx context.Context, ids []string) ([]primitive.ObjectID, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	seen := make(map[primitive.ObjectID]struct{})
	out := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			return nil, errors.New("role_group_ids cannot contain empty values")
		}
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, fmt.Errorf("invalid role_group_id: %s", id)
		}
		if _, ok := seen[oid]; ok {
			continue
		}
		seen[oid] = struct{}{}
		out = append(out, oid)
	}
	existing, err := h.roleGroupRepo.FindExistingIDs(ctx, out)
	if err != nil {
		return nil, err
	}
	if len(existing) != len(out) {
		return nil, errors.New("one or more role_group_ids do not exist")
	}
	return out, nil
}
