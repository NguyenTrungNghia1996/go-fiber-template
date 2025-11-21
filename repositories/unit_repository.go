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

type UnitRepository struct {
	db      *gorm.DB
	regRepo *UnitServicePackageRegistrationRepository
}

func NewUnitRepository(db *gorm.DB, regRepo *UnitServicePackageRegistrationRepository) *UnitRepository {
	return &UnitRepository{db: db, regRepo: regRepo}
}

func (r *UnitRepository) Create(ctx context.Context, u *models.Unit, servicePackageRegs []models.UnitServicePackageRegistration) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if u.ID == "" {
			u.ID = uuid.NewString()
		}
		if err := tx.Create(u).Error; err != nil {
			return err
		}
		for _, reg := range servicePackageRegs {
			regCopy := reg
			regCopy.UnitID = u.ID
			if err := r.regRepo.create(ctx, tx, &regCopy); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *UnitRepository) FindByID(ctx context.Context, id string) (*models.Unit, error) {
	var u models.Unit
	err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// FindBySubdomain returns a Unit by its subdomain or nil if not found.
func (r *UnitRepository) FindBySubdomain(ctx context.Context, subdomain string) (*models.Unit, error) {
	var u models.Unit
	err := r.db.WithContext(ctx).First(&u, "subdomain = ?", subdomain).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UnitRepository) FindPaged(ctx context.Context, page, limit int64, q string) ([]models.Unit, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Unit{})
	if q != "" {
		like := "%" + strings.ToLower(q) + "%"
		query = query.Where("LOWER(subdomain) LIKE ? OR LOWER(name) LIKE ? OR LOWER(description) LIKE ?", like, like, like)
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

	var items []models.Unit
	if err := query.Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, 0, err
	}
	if page == 0 {
		limit = int64(len(items))
	}
	return items, total, nil
}

func (r *UnitRepository) UpdateByID(ctx context.Context, id string, updates map[string]interface{}, servicePackageIDs []string) (*models.Unit, error) {
	var out models.Unit
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			updates["updated_at"] = time.Now().UTC()
			res := tx.Model(&models.Unit{}).
				Where("id = ?", id).
				Clauses(clause.Returning{}).
				Updates(updates).
				Scan(&out)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
		} else {
			if err := tx.First(&out, "id = ?", id).Error; err != nil {
				return err
			}
		}

		if servicePackageIDs != nil {
			var existingRegs []models.UnitServicePackageRegistration
			if err := tx.Where("unit_id = ?", id).Find(&existingRegs).Error; err != nil {
				return err
			}
			existingMap := make(map[string]struct{})
			for _, reg := range existingRegs {
				existingMap[reg.ServicePackageID] = struct{}{}
			}
			newSet := make(map[string]struct{})
			for _, spID := range servicePackageIDs {
				newSet[spID] = struct{}{}
				if _, exists := existingMap[spID]; exists {
					continue
				}
				reg := &models.UnitServicePackageRegistration{
					UnitID:           out.ID,
					ServicePackageID: spID,
				}
				if err := r.regRepo.create(ctx, tx, reg); err != nil {
					return err
				}
			}
			for spID := range existingMap {
				if _, keep := newSet[spID]; keep {
					continue
				}
				if err := r.regRepo.deleteByUnitAndServicePackage(ctx, tx, out.ID, spID); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *UnitRepository) DeleteByID(ctx context.Context, id string) (bool, error) {
	tx := r.db.WithContext(ctx).Delete(&models.Unit{}, "id = ?", id)
	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected > 0, nil
}
