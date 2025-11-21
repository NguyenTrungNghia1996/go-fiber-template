package models

import "time"

// UnitServicePackageRegistration represents the registration of a service package by a unit.
type UnitServicePackageRegistration struct {
	ID               string     `gorm:"type:uuid;primaryKey" json:"id"`
	UnitID           string     `gorm:"type:uuid;not null;index:idx_unit_sp,unique" json:"unit_id"`
	ServicePackageID string     `gorm:"type:uuid;not null;index:idx_unit_sp,unique" json:"service_package_id"`
	StartAt          *time.Time `gorm:"type:timestamptz" json:"start_at,omitempty"`
	EndAt            *time.Time `gorm:"type:timestamptz" json:"end_at,omitempty"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// CreateUnitServicePackageRegistrationInput is the request payload for creating a registration.
type CreateUnitServicePackageRegistrationInput struct {
	UnitID           string `json:"unit_id"`
	ServicePackageID string `json:"service_package_id"`
	StartAt          string `json:"start_at,omitempty"` // RFC3339
	EndAt            string `json:"end_at,omitempty"`   // RFC3339
}

// UpdateUnitServicePackageRegistrationInput is the request payload for updating a registration.
type UpdateUnitServicePackageRegistrationInput struct {
	ID               string  `json:"id"`
	UnitID           *string `json:"unit_id,omitempty"`
	ServicePackageID *string `json:"service_package_id,omitempty"`
	StartAt          *string `json:"start_at,omitempty"` // RFC3339
	EndAt            *string `json:"end_at,omitempty"`   // RFC3339
}
