package models

import "time"

// Unit represents an organizational unit.
type Unit struct {
	ID          string    `gorm:"type:uuid;primaryKey" json:"id"`
	Subdomain   string    `gorm:"uniqueIndex;not null" json:"subdomain"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `gorm:"not null;default:''" json:"description,omitempty"`
	LogoURL     string    `gorm:"not null;default:''" json:"logo_url,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// CreateUnitServicePackageInput allows attaching service packages with optional active window.
// Dates should be RFC3339; empty values mean "no constraint".
type CreateUnitServicePackageInput struct {
	ServicePackageID string `json:"service_package_id"`
	StartAt          string `json:"start_at,omitempty"` // RFC3339
	EndAt            string `json:"end_at,omitempty"`   // RFC3339
}

type CreateUnitInput struct {
	Subdomain   string `json:"subdomain"`
	Name        string `json:"name"`
	Description string `json:"description"`
	LogoURL     string `json:"logo_url"`
	// Deprecated: prefer service_packages to include date ranges.
	ServicePackageIDs []string                        `json:"service_package_ids,omitempty"`
	ServicePackages   []CreateUnitServicePackageInput `json:"service_packages,omitempty"`
	AdminUser         *CreateUnitAdminInput           `json:"admin_user,omitempty"`
}

type UpdateUnitInput struct {
	ID                string    `json:"id"`
	Subdomain         *string   `json:"subdomain,omitempty"`
	Name              *string   `json:"name,omitempty"`
	Description       *string   `json:"description,omitempty"`
	LogoURL           *string   `json:"logo_url,omitempty"`
	ServicePackageIDs *[]string `json:"service_package_ids,omitempty"`
}

// CreateUnitAdminInput allows providing the first admin user of the new unit.
// If provided, username and password are required. The new user is created with is_admin=true.
type CreateUnitAdminInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}
