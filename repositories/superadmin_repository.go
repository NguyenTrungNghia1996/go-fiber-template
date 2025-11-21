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

type SuperAdminRepository struct {
	db *gorm.DB
}

func NewSuperAdminRepository(db *gorm.DB) *SuperAdminRepository {
	return &SuperAdminRepository{db: db}
}

func (r *SuperAdminRepository) Create(ctx context.Context, sa *models.SuperAdmin) error {
	if sa.ID == "" {
		sa.ID = uuid.NewString()
	}
	return r.db.WithContext(ctx).Create(sa).Error
}

func (r *SuperAdminRepository) FindAll(ctx context.Context) ([]models.SuperAdmin, error) {
	var out []models.SuperAdmin
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *SuperAdminRepository) FindPaged(ctx context.Context, page, limit int64, q string) ([]models.SuperAdmin, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.SuperAdmin{})
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
	var out []models.SuperAdmin
	if err := query.Order("created_at DESC").Find(&out).Error; err != nil {
		return nil, 0, err
	}
	if page == 0 {
		limit = int64(len(out))
	}
	return out, total, nil
}

func (r *SuperAdminRepository) FindByID(ctx context.Context, id string) (*models.SuperAdmin, error) {
	var sa models.SuperAdmin
	err := r.db.WithContext(ctx).First(&sa, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sa, nil
}

func (r *SuperAdminRepository) FindByUsername(ctx context.Context, username string) (*models.SuperAdmin, error) {
	var sa models.SuperAdmin
	err := r.db.WithContext(ctx).First(&sa, "username = ?", username).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sa, nil
}

func (r *SuperAdminRepository) UpdateByID(ctx context.Context, id string, updates map[string]interface{}) (*models.SuperAdmin, error) {
	if len(updates) == 0 {
		return nil, fmt.Errorf("no updates provided")
	}
	updates["updated_at"] = gorm.Expr("NOW()")

	var sa models.SuperAdmin
	tx := r.db.WithContext(ctx).Model(&models.SuperAdmin{}).
		Where("id = ?", id).
		Clauses(clause.Returning{}).
		Updates(updates).
		Scan(&sa)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, nil
	}
	return &sa, nil
}

func (r *SuperAdminRepository) DeleteByID(ctx context.Context, id string) (bool, error) {
	tx := r.db.WithContext(ctx).Delete(&models.SuperAdmin{}, "id = ?", id)
	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected > 0, nil
}
