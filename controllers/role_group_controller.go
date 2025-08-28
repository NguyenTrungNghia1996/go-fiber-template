package controllers

import (
	"go-fiber-api/models"
	"go-fiber-api/repositories"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoleGroupController struct {
	Repo *repositories.RoleGroupRepository
}

func NewRoleGroupController(repo *repositories.RoleGroupRepository) *RoleGroupController {
	return &RoleGroupController{Repo: repo}
}

func (ctrl *RoleGroupController) GetRoleGroupDetail(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Status:  "error",
			Message: "Missing id",
			Data:    nil,
		})
	}
	group, err := ctrl.Repo.GetByID(c.Context(), id)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "role group not found" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(models.APIResponse{
			Status:  "error",
			Message: err.Error(),
			Data:    nil,
		})
	}
	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Get role group detail successfully",
		Data:    group.ToResponse(),
	})
}

func (ctrl *RoleGroupController) CreateRoleGroup(c *fiber.Ctx) error {
    var req struct {
        OrganizationID string                    `json:"organization_id"`
        Name           string                    `json:"name"`
        Description    string                    `json:"description"`
        Permission     []models.PermissionDetail `json:"permission"`
    }
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Status:  "error",
			Message: "Invalid data",
			Data:    nil,
		})
	}
    var orgID primitive.ObjectID
    if req.OrganizationID != "" {
        orgID, _ = primitive.ObjectIDFromHex(req.OrganizationID)
    }
    group := models.RoleGroup{
        OrganizationID: orgID,
        Name:           req.Name,
        Description:    req.Description,
        Permission:     req.Permission,
    }
	if err := ctrl.Repo.Create(c.Context(), &group); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Unable to create role group",
			Data:    nil,
		})
	}
	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Created role group successfully",
		Data:    group.ToResponse(),
	})
}

func (ctrl *RoleGroupController) GetRoleGroups(c *fiber.Ctx) error {
    search := c.Query("search")
    organizationID := c.Query("organization_id")
	page, _ := strconv.ParseInt(c.Query("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.Query("limit", "10"), 10, 64)

    groups, total, err := ctrl.Repo.GetAll(c.Context(), search, organizationID, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Cannot get role group list",
			Data:    nil,
		})
	}
	resp := make([]models.RoleGroupListItem, len(groups))
	for i, g := range groups {
		resp[i] = g.ToListItem()
	}
	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Get role group list successfully",
		Data: fiber.Map{
			"items":       resp,
			"totalrecord": total,
		},
	})
}

func (ctrl *RoleGroupController) UpdateRoleGroup(c *fiber.Ctx) error {
    var req struct {
        ID             string                    `json:"id"`
        OrganizationID string                    `json:"organization_id"`
        Name           string                    `json:"name"`
        Description    string                    `json:"description"`
        Permission     []models.PermissionDetail `json:"permission"`
    }
	if err := c.BodyParser(&req); err != nil || req.ID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Status:  "error",
			Message: "Invalid data",
			Data:    nil,
		})
	}
    var orgID primitive.ObjectID
    if req.OrganizationID != "" {
        orgID, _ = primitive.ObjectIDFromHex(req.OrganizationID)
    }
    group := models.RoleGroup{
        OrganizationID: orgID,
        Name:           req.Name,
        Description:    req.Description,
        Permission:     req.Permission,
    }
	if err := ctrl.Repo.UpdateByID(c.Context(), req.ID, &group); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: err.Error(),
			Data:    nil,
		})
	}
	group.ID, _ = primitive.ObjectIDFromHex(req.ID)
	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Updated role group successfully",
		Data:    group.ToResponse(),
	})
}

func (ctrl *RoleGroupController) DeleteRoleGroup(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Status:  "error",
			Message: "Missing id",
			Data:    nil,
		})
	}
	if err := ctrl.Repo.DeleteByID(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: err.Error(),
			Data:    nil,
		})
	}
	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Deleted role group successfully",
		Data:    nil,
	})
}
