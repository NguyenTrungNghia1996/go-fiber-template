package controllers

import (
	"go-fiber-api/models"
	"go-fiber-api/repositories"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type RoleGroupController struct {
	Repo *repositories.RoleGroupRepository
}

func NewRoleGroupController(repo *repositories.RoleGroupRepository) *RoleGroupController {
	return &RoleGroupController{Repo: repo}
}

func (ctrl *RoleGroupController) CreateRoleGroup(c *fiber.Ctx) error {
	var group models.RoleGroup
	if err := c.BodyParser(&group); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Status:  "error",
			Message: "Invalid data",
			Data:    nil,
		})
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
	page, _ := strconv.ParseInt(c.Query("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.Query("limit", "10"), 10, 64)

	groups, total, err := ctrl.Repo.GetAll(c.Context(), search, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Cannot get role group list",
			Data:    nil,
		})
	}
	resp := make([]models.RoleGroupResponse, len(groups))
	for i, g := range groups {
		resp[i] = g.ToResponse()
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
	var group models.RoleGroup
	if err := c.BodyParser(&group); err != nil || group.ID.IsZero() {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Status:  "error",
			Message: "Invalid data",
			Data:    nil,
		})
	}
	if err := ctrl.Repo.UpdateByID(c.Context(), group.ID.Hex(), &group); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: err.Error(),
			Data:    nil,
		})
	}
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
