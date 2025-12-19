package seed

import (
	"context"
	"log"
	"time"

	"go-fiber-api/models"
	"go-fiber-api/repositories"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// SeedSuperAdmin ensures a default super admin account exists.
func SeedSuperAdmin(repo *repositories.SuperAdminRepository, roleGroupIDs []primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exist, err := repo.FindByUsername(ctx, "sa")
	if err != nil {
		return err
	}
	if exist != nil {
		updates := bson.D{}
		if !exist.IsAdmin {
			updates = append(updates, bson.E{Key: "is_admin", Value: true})
			exist.IsAdmin = true
		}
		if len(roleGroupIDs) > 0 {
			merged := mergeObjectIDs(exist.RoleGroupIDs, roleGroupIDs)
			if len(merged) != len(exist.RoleGroupIDs) {
				updates = append(updates, bson.E{Key: "role_group_ids", Value: merged})
				exist.RoleGroupIDs = merged
			}
		}
		if len(updates) > 0 {
			if _, err := repo.UpdateByID(ctx, exist.ID.Hex(), updates); err != nil {
				return err
			}
			log.Println("Seed: updated default super admin 'sa'")
			return nil
		}
		log.Println("Seed: default super admin 'sa' already exists")
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("sa123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	sa := &models.SuperAdmin{
		Username:     "sa",
		PasswordHash: string(hash),
		IsAdmin:      true,
		Name:         "Super Admin",
		Email:        "sa@example.com",
		RoleGroupIDs: mergeObjectIDs(nil, roleGroupIDs),
	}
	if err := repo.Create(ctx, sa); err != nil {
		return err
	}
	log.Println("Seed: created default super admin 'sa'")
	return nil
}

func mergeObjectIDs(existing []primitive.ObjectID, extras []primitive.ObjectID) []primitive.ObjectID {
	seen := make(map[primitive.ObjectID]struct{})
	out := make([]primitive.ObjectID, 0, len(existing)+len(extras))
	for _, id := range existing {
		if id.IsZero() {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	for _, id := range extras {
		if id.IsZero() {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
