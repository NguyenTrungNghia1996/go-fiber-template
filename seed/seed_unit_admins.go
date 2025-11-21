package seed

import (
	"context"
	"log"
	"time"

	"go-fiber-api/models"
	"go-fiber-api/repositories"
	"golang.org/x/crypto/bcrypt"
)

// SeedUnitAdmins ensures each existing unit has a default admin user (admin/admin).
// Only creates the user if it does not already exist.
func SeedUnitAdmins(unitRepo *repositories.UnitRepository, userRepo *repositories.UnitUserRepository) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	units, _, err := unitRepo.FindPaged(ctx, 0, 0, "")
	if err != nil {
		return err
	}
	for _, u := range units {
		// Only create default admin if no admin exists in this unit
		hasAdmin, err := userRepo.AnyAdminExists(ctx, u.ID)
		if err != nil {
			return err
		}
		if hasAdmin {
			continue
		}
		hash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if err := userRepo.Create(ctx, &models.UnitUser{
			UnitID:       u.ID,
			Username:     "admin",
			PasswordHash: string(hash),
			IsAdmin:      true,
			Name:         "Unit Admin",
		}); err != nil {
			// Log and continue to attempt others; tolerate duplicates racing
			log.Printf("seed unit admin error for unit %s: %v", u.ID, err)
		}
	}
	return nil
}
