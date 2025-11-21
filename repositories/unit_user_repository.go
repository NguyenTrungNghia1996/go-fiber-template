package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go-fiber-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UnitUserRepository manages users that belong to units (tenants).
type UnitUserRepository struct {
	db *gorm.DB
}

func NewUnitUserRepository(db *gorm.DB) *UnitUserRepository {
	return &UnitUserRepository{db: db}
}

func (r *UnitUserRepository) Create(ctx context.Context, u *models.UnitUser) error {
	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *UnitUserRepository) FindByUsernameAndUnitID(ctx context.Context, username string, unitID string) (*models.UnitUser, error) {
	var u models.UnitUser
	err := r.db.WithContext(ctx).First(&u, "username = ? AND unit_id = ?", username, unitID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// FindByIDWithinUnit returns a user by id scoped to a unit.
func (r *UnitUserRepository) FindByIDWithinUnit(ctx context.Context, id string, unitID string) (*models.UnitUser, error) {
	var u models.UnitUser
	err := r.db.WithContext(ctx).First(&u, "id = ? AND unit_id = ?", id, unitID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// FindPagedByUnit lists users within a unit with optional search and pagination.
func (r *UnitUserRepository) FindPagedByUnit(ctx context.Context, unitID string, page, limit int64, q string) ([]models.UnitUser, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.UnitUser{}).Where("unit_id = ?", unitID)
	if q != "" {
		like := "%" + strings.ToLower(q) + "%"
		query = query.Where("LOWER(username) LIKE ? OR LOWER(name) LIKE ? OR LOWER(email) LIKE ?", like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page > 0 {
		if limit < 1 {
			limit = 10
		}
		offset := (page - 1) * limit
		query = query.Limit(int(limit)).Offset(int(offset))
	}

	var items []models.UnitUser
	if err := query.Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, 0, err
	}
	if page == 0 {
		limit = int64(len(items))
	}
	return items, total, nil
}

// UpdateByIDWithinUnit updates a user within a unit and returns the updated record.
func (r *UnitUserRepository) UpdateByIDWithinUnit(ctx context.Context, id string, unitID string, updates map[string]interface{}) (*models.UnitUser, error) {
	if len(updates) == 0 {
		return nil, fmt.Errorf("no updates provided")
	}
	updates["updated_at"] = gorm.Expr("NOW()")

	var u models.UnitUser
	tx := r.db.WithContext(ctx).Model(&models.UnitUser{}).
		Where("id = ? AND unit_id = ?", id, unitID).
		Clauses(clause.Returning{}).
		Updates(updates).
		Scan(&u)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, nil
	}
	return &u, nil
}

// DeleteByIDWithinUnit deletes a user within a unit and returns whether a record was deleted.
func (r *UnitUserRepository) DeleteByIDWithinUnit(ctx context.Context, id string, unitID string) (bool, error) {
	tx := r.db.WithContext(ctx).Where("id = ? AND unit_id = ?", id, unitID).Delete(&models.UnitUser{})
	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected > 0, nil
}

// AnyAdminExists checks whether there is at least one admin user in the unit.
func (r *UnitUserRepository) AnyAdminExists(ctx context.Context, unitID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.UnitUser{}).Where("unit_id = ? AND is_admin = TRUE", unitID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// AdminCount returns number of admin users in a unit.
func (r *UnitUserRepository) AdminCount(ctx context.Context, unitID string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.UnitUser{}).Where("unit_id = ? AND is_admin = TRUE", unitID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
