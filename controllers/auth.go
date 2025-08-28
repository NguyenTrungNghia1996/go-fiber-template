package controllers

// This file contains authentication handlers used by the API.

import (
	"go-fiber-api/models"
	"go-fiber-api/repositories"
	"go-fiber-api/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuthController handles authentication requests.
type AuthController struct {
    Repo          *repositories.UserRepository
    RoleGroupRepo *repositories.RoleGroupRepository
    OrganizationRepo *repositories.OrganizationRepository
}

// NewAuthController creates a controller with the provided user repository.
func NewAuthController(repo *repositories.UserRepository, roleRepo *repositories.RoleGroupRepository, orgRepo *repositories.OrganizationRepository) *AuthController {
    return &AuthController{Repo: repo, RoleGroupRepo: roleRepo, OrganizationRepo: orgRepo}
}

// Login authenticates a user and returns a signed JWT on success.
func (ctrl *AuthController) Login(c *fiber.Ctx) error {
    var input struct {
        Username string `json:"username"`
        Password string `json:"password"`
        Subdomain string `json:"subdomain"`
    }
    if err := c.BodyParser(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
            Status:  "error",
            Message: "Invalid input",
            Data:    nil,
        })
    }

    // Resolve organization by subdomain
    org, err := ctrl.OrganizationRepo.FindBySubdomain(c.Context(), input.Subdomain)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(models.APIResponse{
            Status:  "error",
            Message: "Invalid organization/subdomain",
            Data:    nil,
        })
    }

    user, err := ctrl.Repo.FindByUsernameInOrg(c.Context(), input.Username, org.ID)
    if err != nil || !utils.CheckPasswordHash(input.Password, user.Password) {
        return c.Status(fiber.StatusUnauthorized).JSON(models.APIResponse{
            Status:  "error",
            Message: "Invalid credentials",
            Data:    nil,
        })
    }

	// Populate role group details for the logged in user
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

    token, _ := utils.GenerateJWT(user.ID.Hex())
    return c.JSON(models.APIResponse{
        Status:  "success",
        Message: "Login successful",
        Data: fiber.Map{
            "token": token,
            "user":  user.ToListItem(gm),
        },
    })
}
