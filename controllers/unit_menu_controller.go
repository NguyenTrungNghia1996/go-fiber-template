package controllers

import (
	"strings"
	"time"

	"go-fiber-api/models"
	"go-fiber-api/pkg/response"
	"go-fiber-api/repositories"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UnitMenuController exposes APIs for unit menus derived from registered service packages.
type UnitMenuController struct {
	regRepo *repositories.UnitServicePackageRegistrationRepository
	spRepo  *repositories.ServicePackageRepository
}

func NewUnitMenuController(regRepo *repositories.UnitServicePackageRegistrationRepository, spRepo *repositories.ServicePackageRepository) *UnitMenuController {
	return &UnitMenuController{
		regRepo: regRepo,
		spRepo:  spRepo,
	}
}

// List returns the menus available for the current unit based on its active service package registrations.
// GET /unit_menus[?page&limit&q]
func (h *UnitMenuController) List(c *fiber.Ctx) error {
	unitIDStr, _ := c.Locals("unit_id").(string)
	if strings.TrimSpace(unitIDStr) == "" {
		return response.Error(c, "unauthorized", fiber.StatusUnauthorized, nil)
	}

	page, limit := response.ParsePageLimit(c)
	q := strings.ToLower(strings.TrimSpace(c.Query("q")))

	// Load all registrations for this unit (page=0 => all)
	regs, _, err := h.regRepo.FindPaged(c.Context(), 0, 0, unitIDStr, "")
	if err != nil {
		return response.Error(c, "failed to load registrations", fiber.StatusInternalServerError, nil)
	}

	now := time.Now().UTC()
	activeIDsSet := make(map[primitive.ObjectID]struct{})
	for _, reg := range regs {
		if !reg.StartAt.IsZero() && reg.StartAt.After(now) {
			continue
		}
		if !reg.EndAt.IsZero() && reg.EndAt.Before(now) {
			continue
		}
		activeIDsSet[reg.ServicePackageID] = struct{}{}
	}

	if len(activeIDsSet) == 0 {
		data := response.ListData[models.ServicePackageMenu]{
			Items: []models.ServicePackageMenu{},
			Page:  page,
			Limit: limit,
			Total: 0,
		}
		return response.Success(c, data, "ok")
	}

	var activeIDs []primitive.ObjectID
	for id := range activeIDsSet {
		activeIDs = append(activeIDs, id)
	}

	sps, err := h.spRepo.FindByIDs(c.Context(), activeIDs)
	if err != nil {
		return response.Error(c, "failed to load service packages", fiber.StatusInternalServerError, nil)
	}

	menuKeySeen := make(map[string]struct{})
	var menus []models.ServicePackageMenu
	for _, sp := range sps {
		for _, m := range sp.Menus {
			// Simple case-insensitive search over title/key/url when q is provided
			if q != "" {
				lq := q
				if !strings.Contains(strings.ToLower(m.Title), lq) &&
					!strings.Contains(strings.ToLower(m.Key), lq) &&
					!strings.Contains(strings.ToLower(m.URL), lq) {
					continue
				}
			}

			dk := strings.TrimSpace(m.Key)
			if dk == "" {
				dk = strings.TrimSpace(m.URL)
			}
			if dk != "" {
				if _, exists := menuKeySeen[dk]; exists {
					continue
				}
				menuKeySeen[dk] = struct{}{}
			}
			menus = append(menus, m)
		}
	}

	total := int64(len(menus))
	// No pagination requested -> return all
	if page == 0 {
		data := response.ListData[models.ServicePackageMenu]{
			Items: menus,
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
	paged := menus[start:end]

	data := response.ListData[models.ServicePackageMenu]{
		Items: paged,
		Page:  page,
		Limit: limit,
		Total: total,
	}
	return response.Success(c, data, "ok")
}

