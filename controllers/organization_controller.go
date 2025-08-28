package controllers

import (
    "go-fiber-api/models"
    "go-fiber-api/repositories"
    "strconv"

    "github.com/gofiber/fiber/v2"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type OrganizationController struct {
    Repo *repositories.OrganizationRepository
}

func NewOrganizationController(repo *repositories.OrganizationRepository) *OrganizationController {
    return &OrganizationController{Repo: repo}
}

func (ctrl *OrganizationController) GetOrganizationDetail(c *fiber.Ctx) error {
    id := c.Query("id")
    if id == "" {
        return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
            Status:  "error",
            Message: "Missing id",
            Data:    nil,
        })
    }
    org, err := ctrl.Repo.GetByID(c.Context(), id)
    if err != nil {
        status := fiber.StatusInternalServerError
        if err.Error() == "organization not found" {
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
        Message: "Get organization detail successfully",
        Data:    org.ToResponse(),
    })
}

func (ctrl *OrganizationController) CreateOrganization(c *fiber.Ctx) error {
    var req struct {
        Name        string `json:"name"`
        Description string `json:"description"`
    }
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
            Status:  "error",
            Message: "Invalid data",
            Data:    nil,
        })
    }
    org := models.Organization{
        Name:        req.Name,
        Description: req.Description,
    }
    if err := ctrl.Repo.Create(c.Context(), &org); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
            Status:  "error",
            Message: "Unable to create organization",
            Data:    nil,
        })
    }
    return c.JSON(models.APIResponse{
        Status:  "success",
        Message: "Created organization successfully",
        Data:    org.ToResponse(),
    })
}

func (ctrl *OrganizationController) GetOrganizations(c *fiber.Ctx) error {
    search := c.Query("search")
    page, _ := strconv.ParseInt(c.Query("page", "1"), 10, 64)
    limit, _ := strconv.ParseInt(c.Query("limit", "10"), 10, 64)

    items, total, err := ctrl.Repo.GetAll(c.Context(), search, page, limit)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
            Status:  "error",
            Message: "Cannot get organization list",
            Data:    nil,
        })
    }

    resp := make([]models.OrganizationListItem, len(items))
    for i, it := range items {
        resp[i] = it.ToListItem()
    }

    return c.JSON(models.APIResponse{
        Status:  "success",
        Message: "Get organization list successfully",
        Data: fiber.Map{
            "items":       resp,
            "totalrecord": total,
        },
    })
}

func (ctrl *OrganizationController) UpdateOrganization(c *fiber.Ctx) error {
    var req struct {
        ID          string `json:"id"`
        Name        string `json:"name"`
        Description string `json:"description"`
    }
    if err := c.BodyParser(&req); err != nil || req.ID == "" {
        return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
            Status:  "error",
            Message: "Invalid data",
            Data:    nil,
        })
    }
    org := models.Organization{
        Name:        req.Name,
        Description: req.Description,
    }
    if err := ctrl.Repo.UpdateByID(c.Context(), req.ID, &org); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
            Status:  "error",
            Message: err.Error(),
            Data:    nil,
        })
    }
    org.ID, _ = primitive.ObjectIDFromHex(req.ID)
    return c.JSON(models.APIResponse{
        Status:  "success",
        Message: "Updated organization successfully",
        Data:    org.ToResponse(),
    })
}

func (ctrl *OrganizationController) DeleteOrganization(c *fiber.Ctx) error {
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
        Message: "Deleted organization successfully",
        Data:    nil,
    })
}

