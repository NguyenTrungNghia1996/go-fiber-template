package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"go-fiber-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ServicePackageRepository struct {
	db *gorm.DB
}

func NewServicePackageRepository(db *gorm.DB) *ServicePackageRepository {
	return &ServicePackageRepository{db: db}
}

func (r *ServicePackageRepository) Create(ctx context.Context, sp *models.ServicePackage) error {
	if sp.ID == "" {
		sp.ID = uuid.NewString()
	}
	if sp.Menus == nil {
		sp.Menus = models.JSONB[models.ServicePackageMenu]{}
	}
	return r.db.WithContext(ctx).Create(sp).Error
}

func (r *ServicePackageRepository) FindByID(ctx context.Context, id string) (*models.ServicePackage, error) {
	var sp models.ServicePackage
	err := r.db.WithContext(ctx).First(&sp, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sp, nil
}

// FindByIDs returns all service packages whose IDs are in the provided slice.
func (r *ServicePackageRepository) FindByIDs(ctx context.Context, ids []string) ([]models.ServicePackage, error) {
	if len(ids) == 0 {
		return []models.ServicePackage{}, nil
	}
	var items []models.ServicePackage
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ServicePackageRepository) FindPaged(ctx context.Context, page, limit int64, q string) ([]models.ServicePackage, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.ServicePackage{})
	if q != "" {
		like := "%" + strings.ToLower(q) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", like, like)
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

	var items []models.ServicePackage
	if err := query.Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, 0, err
	}
	if page == 0 {
		limit = int64(len(items))
	}
	return items, total, nil
}

func (r *ServicePackageRepository) UpdateByID(ctx context.Context, id string, updates map[string]interface{}) (*models.ServicePackage, error) {
	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}
	updates["updated_at"] = time.Now().UTC()

	var sp models.ServicePackage
	tx := r.db.WithContext(ctx).Model(&models.ServicePackage{}).
		Where("id = ?", id).
		Clauses(clause.Returning{}).
		Updates(updates).
		Scan(&sp)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, nil
	}
	return &sp, nil
}

func (r *ServicePackageRepository) DeleteByID(ctx context.Context, id string) (bool, error) {
	tx := r.db.WithContext(ctx).Delete(&models.ServicePackage{}, "id = ?", id)
	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected > 0, nil
}
