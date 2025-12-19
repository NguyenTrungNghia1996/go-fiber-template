package controllers

import (
	"log"
	"regexp"
	"strings"
	"time"

	"go-fiber-api/models"
	"go-fiber-api/pkg/cloudflare"
	"go-fiber-api/pkg/response"
	"go-fiber-api/repositories"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type UnitController struct {
	repo     *repositories.UnitRepository
	regRepo  *repositories.UnitServicePackageRegistrationRepository
	spRepo   *repositories.ServicePackageRepository
	userRepo *repositories.UnitUserRepository
	dns      *cloudflare.DNSClient
	skipDNS  bool
}

func NewUnitController(repo *repositories.UnitRepository, regRepo *repositories.UnitServicePackageRegistrationRepository, spRepo *repositories.ServicePackageRepository, userRepo *repositories.UnitUserRepository, dns *cloudflare.DNSClient, skipDNS bool) *UnitController {
	return &UnitController{repo: repo, regRepo: regRepo, spRepo: spRepo, userRepo: userRepo, dns: dns, skipDNS: skipDNS}
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
		// Resolve service packages for this unit
		regs, _, err := h.regRepo.FindPaged(c.Context(), 0, 0, id, "")
		if err != nil {
			return response.Error(c, "failed to resolve packages", fiber.StatusInternalServerError, nil)
		}
		spCache := map[string]string{}
		var sps []models.ServicePackageBasic
		for _, r := range regs {
			spID := r.ServicePackageID.Hex()
			name, ok := spCache[spID]
			if !ok {
				sp, err := h.spRepo.FindByID(c.Context(), spID)
				if err != nil {
					return response.Error(c, "failed to resolve packages", fiber.StatusInternalServerError, nil)
				}
				if sp == nil { // skip dangling
					continue
				}
				name = sp.Name
				spCache[spID] = name
			}
			sps = append(sps, models.ServicePackageBasic{ID: spID, Name: name})
		}
		dto := models.UnitDTO{Unit: *u, ServicePackages: sps}
		return response.Success(c, dto, "ok")
	}
	page, limit := response.ParsePageLimit(c)
	q := strings.TrimSpace(c.Query("q"))
	items, total, err := h.repo.FindPaged(c.Context(), page, limit, q)
	if err != nil {
		return response.Error(c, "failed to list", fiber.StatusInternalServerError, nil)
	}
	// For list, resolve packages per unit with simple caching per request
	spNameCache := map[string]string{}
	var out []models.UnitDTO
	for i := range items {
		u := items[i]
		regs, _, err := h.regRepo.FindPaged(c.Context(), 0, 0, u.ID.Hex(), "")
		if err != nil {
			return response.Error(c, "failed to resolve packages", fiber.StatusInternalServerError, nil)
		}
		var sps []models.ServicePackageBasic
		for _, r := range regs {
			spID := r.ServicePackageID.Hex()
			name, ok := spNameCache[spID]
			if !ok {
				sp, err := h.spRepo.FindByID(c.Context(), spID)
				if err != nil {
					return response.Error(c, "failed to resolve packages", fiber.StatusInternalServerError, nil)
				}
				if sp == nil {
					continue
				}
				name = sp.Name
				spNameCache[spID] = name
			}
			sps = append(sps, models.ServicePackageBasic{ID: spID, Name: name})
		}
		out = append(out, models.UnitDTO{Unit: u, ServicePackages: sps})
	}
	data := response.ListData[models.UnitDTO]{
		Items: out,
		Page:  page,
		Limit: limit,
		Total: total,
	}
	if page == 0 {
		data.Limit = total
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
	if !h.skipDNS && h.dns == nil {
		return response.Error(c, "cloudflare dns is not configured", fiber.StatusInternalServerError, nil)
	}
	existing, err := h.repo.FindBySubdomain(c.Context(), in.Subdomain)
	if err != nil {
		return response.Error(c, "failed to check subdomain", fiber.StatusInternalServerError, nil)
	}
	if existing != nil {
		return response.Error(c, "subdomain already exists", fiber.StatusConflict, nil)
	}
	u := &models.Unit{
		Subdomain:   in.Subdomain,
		Name:        in.Name,
		Description: strings.TrimSpace(in.Description),
		LogoURL:     strings.TrimSpace(in.LogoURL),
	}
	// Build service package registrations (support legacy service_package_ids without dates)
	var regs []models.UnitServicePackageRegistration
	seen := map[string]struct{}{}
	if len(in.ServicePackages) > 0 {
		for _, sp := range in.ServicePackages {
			spID := strings.TrimSpace(sp.ServicePackageID)
			if spID == "" {
				return response.Error(c, "service_packages.service_package_id is required", fiber.StatusBadRequest, nil)
			}
			if _, err := primitive.ObjectIDFromHex(spID); err != nil {
				return response.Error(c, "invalid service_packages.service_package_id", fiber.StatusBadRequest, nil)
			}
			if _, duplicate := seen[spID]; duplicate {
				return response.Error(c, "duplicate service_package_id", fiber.StatusBadRequest, nil)
			}
			seen[spID] = struct{}{}

			var startAt, endAt time.Time
			if strings.TrimSpace(sp.StartAt) != "" {
				t, err := time.Parse(time.RFC3339, sp.StartAt)
				if err != nil {
					return response.Error(c, "invalid start_at (use RFC3339)", fiber.StatusBadRequest, nil)
				}
				startAt = t.UTC()
			}
			if strings.TrimSpace(sp.EndAt) != "" {
				t, err := time.Parse(time.RFC3339, sp.EndAt)
				if err != nil {
					return response.Error(c, "invalid end_at (use RFC3339)", fiber.StatusBadRequest, nil)
				}
				endAt = t.UTC()
			}
			if !startAt.IsZero() && !endAt.IsZero() && endAt.Before(startAt) {
				return response.Error(c, "end_at cannot be before start_at", fiber.StatusBadRequest, nil)
			}
			oid, _ := primitive.ObjectIDFromHex(spID)
			regs = append(regs, models.UnitServicePackageRegistration{
				ServicePackageID: oid,
				StartAt:          startAt,
				EndAt:            endAt,
			})
		}
	} else {
		for _, id := range in.ServicePackageIDs {
			if _, err := primitive.ObjectIDFromHex(id); err != nil {
				return response.Error(c, "invalid service_package_ids entry", fiber.StatusBadRequest, nil)
			}
			if _, duplicate := seen[id]; duplicate {
				return response.Error(c, "duplicate service_package_id", fiber.StatusBadRequest, nil)
			}
			seen[id] = struct{}{}
			oid, _ := primitive.ObjectIDFromHex(id)
			regs = append(regs, models.UnitServicePackageRegistration{
				ServicePackageID: oid,
			})
		}
	}
	var createdDNS bool
	if !h.skipDNS {
		_, createdDNS, err = h.dns.EnsureCNAME(c.Context(), in.Subdomain)
		if err != nil {
			return response.Error(c, "failed to provision subdomain", fiber.StatusInternalServerError, fiber.Map{"reason": err.Error()})
		}
	}
	if err := h.repo.Create(c.Context(), u, regs); err != nil {
		if !h.skipDNS && createdDNS {
			_ = h.dns.DeleteByName(c.Context(), in.Subdomain)
		}
		if strings.Contains(err.Error(), "E11000") {
			return response.Error(c, "subdomain already exists", fiber.StatusConflict, nil)
		}
		return response.Error(c, "failed to create", fiber.StatusInternalServerError, nil)
	}
	// Create the first admin user for the unit
	if h.userRepo == nil {
		_, _ = h.repo.DeleteByID(c.Context(), u.ID.Hex())
		_ = h.regRepo.DeleteByUnitID(c.Context(), u.ID)
		if !h.skipDNS && createdDNS {
			_ = h.dns.DeleteByName(c.Context(), in.Subdomain)
		}
		return response.Error(c, "user repository not configured", fiber.StatusInternalServerError, nil)
	}

	rollback := func() {
		_, _ = h.repo.DeleteByID(c.Context(), u.ID.Hex())
		_ = h.regRepo.DeleteByUnitID(c.Context(), u.ID)
		if !h.skipDNS && createdDNS {
			_ = h.dns.DeleteByName(c.Context(), in.Subdomain)
		}
	}

	if in.AdminUser != nil {
		au := *in.AdminUser
		au.Username = strings.TrimSpace(au.Username)
		au.Password = strings.TrimSpace(au.Password)
		if au.Username == "" || au.Password == "" {
			rollback()
			return response.Error(c, "admin_user.username and admin_user.password are required", fiber.StatusBadRequest, nil)
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(au.Password), bcrypt.DefaultCost)
		if err != nil {
			rollback()
			return response.Error(c, "failed to hash admin password", fiber.StatusInternalServerError, nil)
		}
		if err := h.userRepo.Create(c.Context(), &models.UnitUser{
			UnitID:       u.ID,
			Username:     au.Username,
			PasswordHash: string(hash),
			IsAdmin:      true,
			Name:         strings.TrimSpace(au.Name),
			Email:        strings.TrimSpace(au.Email),
		}); err != nil {
			rollback()
			if strings.Contains(err.Error(), "E11000") {
				return response.Error(c, "admin username already exists for this unit", fiber.StatusConflict, nil)
			}
			return response.Error(c, "failed to create admin user", fiber.StatusInternalServerError, fiber.Map{"reason": err.Error()})
		}
	} else {
		hash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		if err != nil {
			rollback()
			return response.Error(c, "failed to hash default admin password", fiber.StatusInternalServerError, nil)
		}
		if err := h.userRepo.Create(c.Context(), &models.UnitUser{
			UnitID:       u.ID,
			Username:     "admin",
			PasswordHash: string(hash),
			IsAdmin:      true,
			Name:         "Unit Admin",
			Email:        "",
		}); err != nil {
			rollback()
			if strings.Contains(err.Error(), "E11000") {
				return response.Error(c, "default admin username already exists", fiber.StatusConflict, nil)
			}
			return response.Error(c, "failed to create default admin user", fiber.StatusInternalServerError, fiber.Map{"reason": err.Error()})
		}
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

	existing, err := h.repo.FindByID(c.Context(), id)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if existing == nil {
		return response.Error(c, "not found", fiber.StatusNotFound, nil)
	}
	oldSubdomain := existing.Subdomain

	updates := bson.D{}
	subdomainChanged := false
	newSubdomain := oldSubdomain
	if in.Subdomain != nil {
		v := strings.ToLower(strings.TrimSpace(*in.Subdomain))
		if v == "" {
			return response.Error(c, "subdomain cannot be empty", fiber.StatusBadRequest, nil)
		}
		if !regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`).MatchString(v) {
			return response.Error(c, "invalid subdomain format", fiber.StatusBadRequest, nil)
		}
		if v != oldSubdomain {
			subdomainChanged = true
			newSubdomain = v
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
	if len(updates) == 0 && in.ServicePackageIDs == nil && in.ServicePackages == nil {
		return response.Error(c, "no fields to update", fiber.StatusBadRequest, nil)
	}
	// Validate service package payload
	// Only one of service_package_ids or service_packages may be provided.
	if in.ServicePackageIDs != nil && in.ServicePackages != nil {
		return response.Error(c, "provide either service_package_ids or service_packages, not both", fiber.StatusBadRequest, nil)
	}
	// servicePackageIDs preserves legacy semantics:
	// - nil => not provided (no change)
	// - empty slice => clear all registrations
	var servicePackageIDs []string
	var servicePackageRegs []models.UnitServicePackageRegistration
	if in.ServicePackages != nil {
		seen := map[string]struct{}{}
		for _, sp := range *in.ServicePackages {
			spID := strings.TrimSpace(sp.ServicePackageID)
			if spID == "" {
				return response.Error(c, "service_packages.service_package_id is required", fiber.StatusBadRequest, nil)
			}
			if _, err := primitive.ObjectIDFromHex(spID); err != nil {
				return response.Error(c, "invalid service_packages.service_package_id", fiber.StatusBadRequest, nil)
			}
			if _, dup := seen[spID]; dup {
				return response.Error(c, "duplicate service_package_id", fiber.StatusBadRequest, nil)
			}
			seen[spID] = struct{}{}

			var startAt, endAt time.Time
			if s := strings.TrimSpace(sp.StartAt); s != "" {
				t, err := time.Parse(time.RFC3339, s)
				if err != nil {
					return response.Error(c, "invalid service_packages.start_at (use RFC3339)", fiber.StatusBadRequest, nil)
				}
				startAt = t.UTC()
			}
			if s := strings.TrimSpace(sp.EndAt); s != "" {
				t, err := time.Parse(time.RFC3339, s)
				if err != nil {
					return response.Error(c, "invalid service_packages.end_at (use RFC3339)", fiber.StatusBadRequest, nil)
				}
				endAt = t.UTC()
			}
			if !startAt.IsZero() && !endAt.IsZero() && endAt.Before(startAt) {
				return response.Error(c, "end_at cannot be before start_at", fiber.StatusBadRequest, nil)
			}
			oid, _ := primitive.ObjectIDFromHex(spID)
			servicePackageRegs = append(servicePackageRegs, models.UnitServicePackageRegistration{
				ServicePackageID: oid,
				StartAt:          startAt,
				EndAt:            endAt,
			})
		}
		servicePackageIDs = nil
	} else if in.ServicePackageIDs != nil {
		for _, id := range *in.ServicePackageIDs {
			if _, err := primitive.ObjectIDFromHex(id); err != nil {
				return response.Error(c, "invalid service_package_ids entry", fiber.StatusBadRequest, nil)
			}
		}
		servicePackageIDs = *in.ServicePackageIDs
		servicePackageRegs = nil
	} else {
		servicePackageIDs = nil
		servicePackageRegs = nil
	}

	if subdomainChanged {
		if !h.skipDNS && h.dns == nil {
			return response.Error(c, "cloudflare dns is not configured", fiber.StatusInternalServerError, nil)
		}
		dup, err := h.repo.FindBySubdomain(c.Context(), newSubdomain)
		if err != nil {
			return response.Error(c, "failed to check subdomain", fiber.StatusInternalServerError, nil)
		}
		if dup != nil && dup.ID != existing.ID {
			return response.Error(c, "subdomain already exists", fiber.StatusConflict, nil)
		}
	}

	var dnsCreated bool
	if subdomainChanged && !h.skipDNS {
		_, dnsCreated, err = h.dns.EnsureCNAME(c.Context(), newSubdomain)
		if err != nil {
			return response.Error(c, "failed to provision subdomain", fiber.StatusInternalServerError, fiber.Map{"reason": err.Error()})
		}
	}

	u, err := h.repo.UpdateByID(c.Context(), id, updates, servicePackageIDs, servicePackageRegs)
	if err != nil {
		if subdomainChanged && !h.skipDNS && dnsCreated {
			_ = h.dns.DeleteByName(c.Context(), newSubdomain)
		}
		if strings.Contains(err.Error(), "E11000") {
			return response.Error(c, "subdomain already exists", fiber.StatusConflict, nil)
		}
		return response.Error(c, err.Error(), fiber.StatusBadRequest, nil)
	}
	if u == nil {
		if subdomainChanged && !h.skipDNS && dnsCreated {
			_ = h.dns.DeleteByName(c.Context(), newSubdomain)
		}
		return response.Error(c, "not found", fiber.StatusNotFound, nil)
	}
	if subdomainChanged && !h.skipDNS {
		if err := h.dns.DeleteByName(c.Context(), oldSubdomain); err != nil {
			log.Printf("failed to clean up old subdomain %s: %v", oldSubdomain, err)
		}
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
