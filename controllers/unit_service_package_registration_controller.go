package controllers

import (
	"strings"
	"time"

	"go-fiber-api/models"
	"go-fiber-api/pkg/response"
	"go-fiber-api/repositories"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UnitServicePackageRegistrationController struct {
	repo     *repositories.UnitServicePackageRegistrationRepository
	unitRepo *repositories.UnitRepository
	spRepo   *repositories.ServicePackageRepository
}

func NewUnitServicePackageRegistrationController(
	repo *repositories.UnitServicePackageRegistrationRepository,
	unitRepo *repositories.UnitRepository,
	spRepo *repositories.ServicePackageRepository,
) *UnitServicePackageRegistrationController {
	return &UnitServicePackageRegistrationController{
		repo:     repo,
		unitRepo: unitRepo,
		spRepo:   spRepo,
	}
}

// Create handles POST /unit_service_package_registrations
func (h *UnitServicePackageRegistrationController) Create(c *fiber.Ctx) error {
	var in models.CreateUnitServicePackageRegistrationInput
	if err := c.BodyParser(&in); err != nil {
		return response.Error(c, "invalid body", fiber.StatusBadRequest, nil)
	}

	if in.UnitID == "" || in.ServicePackageID == "" {
		return response.Error(c, "unit_id and service_package_id are required", fiber.StatusBadRequest, nil)
	}

	startStr := strings.TrimSpace(in.StartAt)
	endStr := strings.TrimSpace(in.EndAt)

	unitOID, err := primitive.ObjectIDFromHex(in.UnitID)
	if err != nil {
		return response.Error(c, "invalid unit_id format", fiber.StatusBadRequest, nil)
	}
	servicePackageOID, err := primitive.ObjectIDFromHex(in.ServicePackageID)
	if err != nil {
		return response.Error(c, "invalid service_package_id format", fiber.StatusBadRequest, nil)
	}
	unit, err := h.unitRepo.FindByID(c.Context(), in.UnitID)
	if err != nil {
		return response.Error(c, "failed to load unit", fiber.StatusInternalServerError, nil)
	}
	if unit == nil {
		return response.Error(c, "unit not found", fiber.StatusNotFound, nil)
	}
	sp, err := h.spRepo.FindByID(c.Context(), in.ServicePackageID)
	if err != nil {
		return response.Error(c, "failed to load service package", fiber.StatusInternalServerError, nil)
	}
	if sp == nil {
		return response.Error(c, "service package not found", fiber.StatusNotFound, nil)
	}

	now := time.Now().UTC()
	var startAt time.Time
	if startStr == "" {
		startAt = now
	} else {
		t, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			return response.Error(c, "invalid start_at format, must be RFC3339", fiber.StatusBadRequest, nil)
		}
		startAt = t
	}

	var endAt time.Time
	if endStr != "" {
		t, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			return response.Error(c, "invalid end_at format, must be RFC3339", fiber.StatusBadRequest, nil)
		}
		endAt = t
		// Optional logical validation
		if !endAt.After(startAt) {
			return response.Error(c, "end_at must be after start_at", fiber.StatusBadRequest, nil)
		}
	}

	reg := &models.UnitServicePackageRegistration{
		UnitID:           unitOID,
		ServicePackageID: servicePackageOID,
		StartAt:          startAt,
		EndAt:            endAt,
	}

	if err := h.repo.Create(c.Context(), reg); err != nil {
		// handle duplicate key error
		if strings.Contains(err.Error(), "E11000") {
			return response.Error(c, "registration already exists for this unit and service package", fiber.StatusConflict, nil)
		}
		return response.Error(c, "failed to create registration", fiber.StatusInternalServerError, nil)
	}

	return response.Success(c, reg, "registration created", fiber.StatusCreated)
}

// List handles GET /unit_service_package_registrations with optional ?id, ?unit_id, ?service_package_id and pagination.
func (h *UnitServicePackageRegistrationController) List(c *fiber.Ctx) error {
	// If query id is present, return that single resource
	if id := strings.TrimSpace(c.Query("id")); id != "" {
		reg, err := h.repo.FindByID(c.Context(), id)
		if err != nil {
			return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
		}
		if reg == nil {
			return response.Error(c, "registration not found", fiber.StatusNotFound, nil)
		}
		return response.Success(c, reg, "ok")
	}

	// Otherwise list with pagination + search by unit_id or service_package_id
	page, limit := response.ParsePageLimit(c)
	unitID := strings.TrimSpace(c.Query("unit_id"))
	servicePackageID := strings.TrimSpace(c.Query("service_package_id"))

	items, total, err := h.repo.FindPaged(c.Context(), page, limit, unitID, servicePackageID)
	if err != nil {
		return response.Error(c, "failed to list registrations", fiber.StatusInternalServerError, nil)
	}

	data := response.ListData[models.UnitServicePackageRegistration]{
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

// Update handles PUT /unit_service_package_registrations
func (h *UnitServicePackageRegistrationController) Update(c *fiber.Ctx) error {
	var in models.UpdateUnitServicePackageRegistrationInput
	if err := c.BodyParser(&in); err != nil {
		return response.Error(c, "invalid body", fiber.StatusBadRequest, nil)
	}

	id := strings.TrimSpace(in.ID)
	if id == "" {
		return response.Error(c, "id is required in body", fiber.StatusBadRequest, nil)
	}

	existing, err := h.repo.FindByID(c.Context(), id)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if existing == nil {
		return response.Error(c, "registration not found", fiber.StatusNotFound, nil)
	}

	newStart := existing.StartAt
	newEnd := existing.EndAt

	updates := bson.D{}
	if in.UnitID != nil {
		unitOID, err := primitive.ObjectIDFromHex(*in.UnitID)
		if err != nil {
			return response.Error(c, "invalid unit_id format", fiber.StatusBadRequest, nil)
		}
		unit, err := h.unitRepo.FindByID(c.Context(), *in.UnitID)
		if err != nil {
			return response.Error(c, "failed to load unit", fiber.StatusInternalServerError, nil)
		}
		if unit == nil {
			return response.Error(c, "unit not found", fiber.StatusNotFound, nil)
		}
		updates = append(updates, bson.E{Key: "unit_id", Value: unitOID})
	}
	if in.ServicePackageID != nil {
		servicePackageOID, err := primitive.ObjectIDFromHex(*in.ServicePackageID)
		if err != nil {
			return response.Error(c, "invalid service_package_id format", fiber.StatusBadRequest, nil)
		}
		sp, err := h.spRepo.FindByID(c.Context(), *in.ServicePackageID)
		if err != nil {
			return response.Error(c, "failed to load service package", fiber.StatusInternalServerError, nil)
		}
		if sp == nil {
			return response.Error(c, "service package not found", fiber.StatusNotFound, nil)
		}
		updates = append(updates, bson.E{Key: "service_package_id", Value: servicePackageOID})
	}
	if in.StartAt != nil {
		startStr := strings.TrimSpace(*in.StartAt)
		if startStr == "" {
			return response.Error(c, "start_at cannot be empty", fiber.StatusBadRequest, nil)
		}
		t, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			return response.Error(c, "invalid start_at format, must be RFC3339", fiber.StatusBadRequest, nil)
		}
		newStart = t
		updates = append(updates, bson.E{Key: "start_at", Value: t})
	}
	if in.EndAt != nil {
		endStr := strings.TrimSpace(*in.EndAt)
		if endStr == "" {
			// Allow clearing end_at by passing empty string
			newEnd = time.Time{}
			updates = append(updates, bson.E{Key: "end_at", Value: time.Time{}})
		} else {
			t, err := time.Parse(time.RFC3339, endStr)
			if err != nil {
				return response.Error(c, "invalid end_at format, must be RFC3339", fiber.StatusBadRequest, nil)
			}
			newEnd = t
			updates = append(updates, bson.E{Key: "end_at", Value: t})
		}
	}

	if len(updates) == 0 {
		return response.Error(c, "no fields to update", fiber.StatusBadRequest, nil)
	}

	if !newEnd.IsZero() && !newStart.IsZero() && !newEnd.After(newStart) {
		return response.Error(c, "end_at must be after start_at", fiber.StatusBadRequest, nil)
	}

	reg, err := h.repo.UpdateByID(c.Context(), id, updates)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}

	return response.Success(c, reg, "registration updated")
}

// Delete handles DELETE /unit_service_package_registrations
func (h *UnitServicePackageRegistrationController) Delete(c *fiber.Ctx) error {
	id := strings.TrimSpace(c.Query("id"))
	if id == "" {
		return response.Error(c, "id query param is required", fiber.StatusBadRequest, nil)
	}

	ok, err := h.repo.DeleteByID(c.Context(), id)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if !ok {
		return response.Error(c, "registration not found", fiber.StatusNotFound, nil)
	}

	return response.Success(c, true, "registration deleted")
}
