package repositories

import (
	"context"
	"errors"
	"time"

	"go-fiber-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UnitServicePackageRegistrationRepository struct {
	db *gorm.DB
}

func NewUnitServicePackageRegistrationRepository(db *gorm.DB) *UnitServicePackageRegistrationRepository {
	return &UnitServicePackageRegistrationRepository{db: db}
}

func (r *UnitServicePackageRegistrationRepository) Create(ctx context.Context, reg *models.UnitServicePackageRegistration) error {
	return r.create(ctx, r.db, reg)
}

func (r *UnitServicePackageRegistrationRepository) create(ctx context.Context, db *gorm.DB, reg *models.UnitServicePackageRegistration) error {
	if reg.ID == "" {
		reg.ID = uuid.NewString()
	}
	return db.WithContext(ctx).Create(reg).Error
}

func (r *UnitServicePackageRegistrationRepository) FindByID(ctx context.Context, id string) (*models.UnitServicePackageRegistration, error) {
	var reg models.UnitServicePackageRegistration
	err := r.db.WithContext(ctx).First(&reg, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &reg, nil
}

func (r *UnitServicePackageRegistrationRepository) FindPaged(ctx context.Context, page, limit int64, unitID, servicePackageID string) ([]models.UnitServicePackageRegistration, int64, error) {
	return r.findPaged(ctx, r.db, page, limit, unitID, servicePackageID)
}

func (r *UnitServicePackageRegistrationRepository) findPaged(ctx context.Context, db *gorm.DB, page, limit int64, unitID, servicePackageID string) ([]models.UnitServicePackageRegistration, int64, error) {
	query := db.WithContext(ctx).Model(&models.UnitServicePackageRegistration{})
	if unitID != "" {
		query = query.Where("unit_id = ?", unitID)
	}
	if servicePackageID != "" {
		query = query.Where("service_package_id = ?", servicePackageID)
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

	var items []models.UnitServicePackageRegistration
	if err := query.Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, 0, err
	}
	if page == 0 {
		limit = int64(len(items))
	}
	return items, total, nil
}

func (r *UnitServicePackageRegistrationRepository) UpdateByID(ctx context.Context, id string, updates map[string]interface{}) (*models.UnitServicePackageRegistration, error) {
	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}
	updates["updated_at"] = time.Now().UTC()

	var reg models.UnitServicePackageRegistration
	tx := r.db.WithContext(ctx).Model(&models.UnitServicePackageRegistration{}).
		Where("id = ?", id).
		Clauses(clause.Returning{}).
		Updates(updates).
		Scan(&reg)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, nil
	}
	return &reg, nil
}

func (r *UnitServicePackageRegistrationRepository) DeleteByID(ctx context.Context, id string) (bool, error) {
	tx := r.db.WithContext(ctx).Delete(&models.UnitServicePackageRegistration{}, "id = ?", id)
	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected > 0, nil
}

func (r *UnitServicePackageRegistrationRepository) deleteByUnitAndServicePackage(ctx context.Context, db *gorm.DB, unitID, spID string) error {
	return db.WithContext(ctx).Where("unit_id = ? AND service_package_id = ?", unitID, spID).
		Delete(&models.UnitServicePackageRegistration{}).Error
}
