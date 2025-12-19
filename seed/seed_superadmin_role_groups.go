package seed

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-fiber-api/models"
	"go-fiber-api/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SeedSuperAdminRoleGroups ensures default super admin role groups exist with fixed IDs.
// Returns the list of ensured role group IDs for downstream seeding (e.g., default super admin assignment).
func SeedSuperAdminRoleGroups(repo *repositories.SuperAdminRoleGroupRepository) ([]primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	type permSeed struct {
		Key             string
		PermissionValue int64
	}
	type roleGroupSeed struct {
		ID          string
		Name        string
		Description string
		Permissions []permSeed
		CreatedAt   string
		UpdatedAt   string
	}

	groups := []roleGroupSeed{
		{
			ID:          "6943bf0100cbf143abbff0bb",
			Name:        "Full Access",
			Description: "Tất cả quyền cho super admin",
			Permissions: []permSeed{
				{Key: "menu", PermissionValue: 42},
				{Key: "menu-wp59tm6h4h", PermissionValue: 42},
			},
			CreatedAt: "2025-12-18T08:44:49.58Z",
			UpdatedAt: "2025-12-19T02:41:54.436Z",
		},
	}

	ensuredIDs := make([]primitive.ObjectID, 0, len(groups))
	for _, g := range groups {
		targetID, err := primitive.ObjectIDFromHex(g.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid role group id %s: %w", g.ID, err)
		}

		existing, err := repo.FindByID(ctx, g.ID)
		if err != nil {
			return nil, fmt.Errorf("check existing role group id %s: %w", g.ID, err)
		}
		if existing != nil {
			ensuredIDs = append(ensuredIDs, targetID)
			continue
		}

		existingName, err := repo.FindByName(ctx, g.Name)
		if err != nil {
			return nil, fmt.Errorf("check existing role group name %s: %w", g.Name, err)
		}
		if existingName != nil && existingName.ID != targetID {
			return nil, fmt.Errorf("role group name %s already exists with a different id %s", g.Name, existingName.ID.Hex())
		}

		createdAt, err := time.Parse(time.RFC3339Nano, g.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid created_at for role group %s: %w", g.Name, err)
		}
		updatedAt, err := time.Parse(time.RFC3339Nano, g.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid updated_at for role group %s: %w", g.Name, err)
		}

		perms := make([]models.SuperAdminMenuPermission, 0, len(g.Permissions))
		for _, p := range g.Permissions {
			perms = append(perms, models.SuperAdminMenuPermission{
				Key:             p.Key,
				PermissionValue: p.PermissionValue,
			})
		}

		group := models.SuperAdminRoleGroup{
			ID:          targetID,
			Name:        g.Name,
			Description: g.Description,
			Permissions: perms,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		}
		if err := repo.UpsertWithFixedID(ctx, group); err != nil {
			return nil, fmt.Errorf("seed super admin role group %s: %w", g.Name, err)
		}

		ensuredIDs = append(ensuredIDs, targetID)
	}

	log.Println("Seed: ensured default super admin role groups")
	return ensuredIDs, nil
}
