package seed

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go-fiber-api/models"
	"go-fiber-api/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SeedSuperAdminMenus ensures default super admin menus exist with fixed IDs/keys.
func SeedSuperAdminMenus(repo *repositories.SuperAdminMenuRepository) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	type menuSeed struct {
		ID         string
		Title      string
		Key        string
		URL        string
		Icon       string
		ParentID   string
		Permission int64
		Active     bool
		CreatedAt  string
		UpdatedAt  string
	}

	menus := []menuSeed{
		{
			ID:         "6944b3896034bc07712c1fd6",
			Title:      "Quản lý",
			Key:        "menu-wp59tm6h4h",
			URL:        "/admin",
			Icon:       "ant-design:setting-outlined",
			ParentID:   "000000000000000000000000",
			Permission: 0,
			Active:     true,
			CreatedAt:  "2025-12-19T02:08:09.045Z",
			UpdatedAt:  "2025-12-19T02:08:09.045Z",
		},
		{
			ID:         "6944b3ba6034bc07712c1fd7",
			Title:      "Menu",
			Key:        "menu-kksc1edbms",
			URL:        "/admin/menu",
			Icon:       "ant-design:menu-outlined",
			ParentID:   "6944b3896034bc07712c1fd6",
			Permission: 0,
			Active:     true,
			CreatedAt:  "2025-12-19T02:08:58.046Z",
			UpdatedAt:  "2025-12-19T02:08:58.046Z",
		},
		{
			ID:         "6944b4026034bc07712c1fd9",
			Title:      "Người dùng",
			Key:        "menu-9swf8sikdo",
			URL:        "/admin/user",
			Icon:       "ant-design:user-outlined",
			ParentID:   "6944b3896034bc07712c1fd6",
			Permission: 4,
			Active:     true,
			CreatedAt:  "2025-12-19T02:10:10.719Z",
			UpdatedAt:  "2025-12-19T02:10:10.719Z",
		},
		{
			ID:         "6944b3e36034bc07712c1fd8",
			Title:      "Nhóm quyền",
			Key:        "menu-ehxi97xuxf",
			URL:        "/admin/role_groups",
			Icon:       "ant-design:usergroup-add-outlined",
			ParentID:   "6944b3896034bc07712c1fd6",
			Permission: 2,
			Active:     true,
			CreatedAt:  "2025-12-19T02:09:39.896Z",
			UpdatedAt:  "2025-12-19T02:09:39.896Z",
		},
	}

	for _, m := range menus {
		// Only insert if menu does not already exist (by ID or key)
		existing, err := repo.FindByID(ctx, m.ID)
		if err != nil {
			return fmt.Errorf("check existing menu id %s: %w", m.ID, err)
		}
		if existing != nil {
			continue
		}
		existingKey, err := repo.FindByKey(ctx, m.Key)
		if err != nil {
			return fmt.Errorf("check existing menu key %s: %w", m.Key, err)
		}
		if existingKey != nil {
			continue
		}

		id, err := primitive.ObjectIDFromHex(m.ID)
		if err != nil {
			return fmt.Errorf("invalid menu id %s: %w", m.ID, err)
		}
		parentID := primitive.NilObjectID
		if strings.TrimSpace(m.ParentID) != "" {
			pid, err := primitive.ObjectIDFromHex(m.ParentID)
			if err != nil {
				return fmt.Errorf("invalid parent id for menu %s: %w", m.Key, err)
			}
			parentID = pid
		}
		createdAt, err := time.Parse(time.RFC3339Nano, m.CreatedAt)
		if err != nil {
			return fmt.Errorf("invalid created_at for menu %s: %w", m.Key, err)
		}
		updatedAt, err := time.Parse(time.RFC3339Nano, m.UpdatedAt)
		if err != nil {
			return fmt.Errorf("invalid updated_at for menu %s: %w", m.Key, err)
		}

		menu := models.SuperAdminMenu{
			ID:         id,
			Title:      m.Title,
			Key:        m.Key,
			URL:        m.URL,
			Icon:       m.Icon,
			ParentID:   parentID,
			Permission: m.Permission,
			Active:     m.Active,
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
		}
		if err := repo.UpsertWithFixedID(ctx, menu); err != nil {
			return fmt.Errorf("seed super admin menu %s: %w", m.Key, err)
		}
	}

	log.Println("Seed: ensured default super admin menus")
	return nil
}
