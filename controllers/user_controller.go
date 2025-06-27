package controllers

// This file defines handlers for user-related endpoints.

import (
	"go-fiber-api/models"
	"go-fiber-api/repositories"
	"go-fiber-api/utils"

	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserController struct {
	Repo          *repositories.UserRepository
	RoleGroupRepo *repositories.RoleGroupRepository
}

// NewUserController creates a controller with the given repository.
func NewUserController(repo *repositories.UserRepository, roleRepo *repositories.RoleGroupRepository) *UserController {
	return &UserController{Repo: repo, RoleGroupRepo: roleRepo}
}

// CreateUser handles POST /api/users and registers a new account.
func (ctrl *UserController) CreateUser(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Status:  "error",
			Message: "Invalid data",
			Data:    nil,
		})
	}

	// Check username exists
	exists, err := ctrl.Repo.IsUsernameExists(c.Context(), user.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Error checking username",
			Data:    nil,
		})
	}
	if exists {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Status:  "error",
			Message: "Username already exists",
			Data:    nil,
		})
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Failed to hash password",
			Data:    nil,
		})
	}
	user.Password = hashedPassword

	// Create user
	if err := ctrl.Repo.Create(c.Context(), &user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Unable to create user",
			Data:    nil,
		})
	}

	user.Password = ""
	groups, err := ctrl.RoleGroupRepo.GetByIDs(c.Context(), user.RoleGroups)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Cannot get role groups",
			Data:    nil,
		})
	}
	gm := make(map[primitive.ObjectID]models.RoleGroupListItem, len(groups))
	for _, g := range groups {
		gm[g.ID] = g.ToListItem()
	}

	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Created user successfully",
		Data:    user.ToListItem(gm),
	})
}

// GetUsers returns a paginated list of users.
func (ctrl *UserController) GetUsers(c *fiber.Ctx) error {
	search := c.Query("search")
	page, _ := strconv.ParseInt(c.Query("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.Query("limit", "10"), 10, 64)

	users, total, err := ctrl.Repo.GetAll(c.Context(), search, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Cannot get user list",
			Data:    nil,
		})
	}

	// Collect all role group IDs across users for lookup
	idSet := map[primitive.ObjectID]struct{}{}
	for _, u := range users {
		for _, id := range u.RoleGroups {
			idSet[id] = struct{}{}
		}
	}
	ids := make([]primitive.ObjectID, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}

	groups, err := ctrl.RoleGroupRepo.GetByIDs(c.Context(), ids)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Cannot get role groups",
			Data:    nil,
		})
	}

	groupMap := make(map[primitive.ObjectID]models.RoleGroupListItem, len(groups))
	for _, g := range groups {
		groupMap[g.ID] = g.ToListItem()
	}

	resp := make([]models.UserListItem, len(users))
	for i, u := range users {
		resp[i] = u.ToListItem(groupMap)
	}

	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Get user list successfully",
		Data: fiber.Map{
			"items": resp,
			"total": total,
		},
	})
}

// PUT /api/users/password
// ChangeUserPassword allows an authenticated user to update their password.
func (ctrl *UserController) ChangeUserPassword(c *fiber.Ctx) error {
	var body struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := c.BodyParser(&body); err != nil || body.OldPassword == "" || body.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Status:  "error",
			Message: "Invalid data",
			Data:    nil,
		})
	}

	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	id, _ := claims["id"].(string)

	user, err := ctrl.Repo.FindByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.APIResponse{
			Status:  "error",
			Message: "User not found",
			Data:    nil,
		})
	}

	if !utils.CheckPasswordHash(body.OldPassword, user.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIResponse{
			Status:  "error",
			Message: "Old password is incorrect",
			Data:    nil,
		})
	}

	hashed, err := utils.HashPassword(body.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Unable to encrypt password",
			Data:    nil,
		})
	}

	err = ctrl.Repo.UpdatePassword(c.Context(), id, hashed)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Unable to update password",
			Data:    nil,
		})
	}

	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Password changed successfully",
		Data:    nil,
	})
}

// UpdateUser handles PUT /api/users to update user information
func (ctrl *UserController) UpdateUser(c *fiber.Ctx) error {
	var req struct {
		ID         primitive.ObjectID   `json:"id"`
		Name       string               `json:"name"`
		UrlAvatar  string               `json:"url_avatar"`
		RoleGroups []primitive.ObjectID `json:"role_groups"`
	}
	if err := c.BodyParser(&req); err != nil || req.ID.IsZero() {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Status:  "error",
			Message: "Invalid data",
			Data:    nil,
		})
	}

	user, err := ctrl.Repo.UpdateByID(c.Context(), req.ID.Hex(), req.Name, req.UrlAvatar, req.RoleGroups)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "user not found" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(models.APIResponse{
			Status:  "error",
			Message: err.Error(),
			Data:    nil,
		})
	}
	user.Password = ""
	groups, err := ctrl.RoleGroupRepo.GetByIDs(c.Context(), user.RoleGroups)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Cannot get role groups",
			Data:    nil,
		})
	}
	gm := make(map[primitive.ObjectID]models.RoleGroupListItem, len(groups))
	for _, g := range groups {
		gm[g.ID] = g.ToListItem()
	}
	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Updated user successfully",
		Data:    user.ToListItem(gm),
	})
}

// GetCurrentUser returns profile information for the authenticated user.
func (ctrl *UserController) GetCurrentUser(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	id, _ := claims["id"].(string)

	user, err := ctrl.Repo.FindByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.APIResponse{
			Status:  "error",
			Message: "User not found",
			Data:    nil,
		})
	}
	user.Password = ""
	groups, err := ctrl.RoleGroupRepo.GetByIDs(c.Context(), user.RoleGroups)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Cannot get role groups",
			Data:    nil,
		})
	}
	gm := make(map[primitive.ObjectID]models.RoleGroupListItem, len(groups))
	for _, g := range groups {
		gm[g.ID] = g.ToListItem()
	}

	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Get profile successfully",
		Data:    user.ToListItem(gm),
	})
}

// GetUserPermissions aggregates permissions from the user's role groups.
func (ctrl *UserController) GetUserPermissions(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	id, _ := claims["id"].(string)

	user, err := ctrl.Repo.FindByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.APIResponse{
			Status:  "error",
			Message: "User not found",
			Data:    nil,
		})
	}

	groups, err := ctrl.RoleGroupRepo.GetByIDs(c.Context(), user.RoleGroups)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Cannot get role groups",
			Data:    nil,
		})
	}

	permMap := map[string]int64{}
	for _, g := range groups {
		for _, p := range g.Permission {
			permMap[p.Key] |= p.PermissionValue
		}
	}

	permissions := make([]models.PermissionDetail, 0, len(permMap))
	for k, v := range permMap {
		permissions = append(permissions, models.PermissionDetail{Key: k, PermissionValue: v})
	}

	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Get permissions successfully",
		Data: fiber.Map{
			"permission": permissions,
		},
	})
}
