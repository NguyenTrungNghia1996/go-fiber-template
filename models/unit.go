package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Unit represents an organizational unit.
type Unit struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Subdomain   string             `bson:"subdomain" json:"subdomain"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	LogoURL     string             `bson:"logo_url,omitempty" json:"logo_url,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
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
	ID                string                           `json:"id"`
	Subdomain         *string                          `json:"subdomain,omitempty"`
	Name              *string                          `json:"name,omitempty"`
	Description       *string                          `json:"description,omitempty"`
	LogoURL           *string                          `json:"logo_url,omitempty"`
	ServicePackageIDs *[]string                        `json:"service_package_ids,omitempty"`
	ServicePackages   *[]CreateUnitServicePackageInput `json:"service_packages,omitempty"`
}

// CreateUnitAdminInput allows providing the first admin user of the new unit.
// If provided, username and password are required. The new user is created with is_admin=true.
type CreateUnitAdminInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	ImageURL string `json:"image_url"`
}
