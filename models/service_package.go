package models

import "time"

// ServicePackage represents a service package offered by a unit.
type ServicePackage struct {
	ID          string                    `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string                    `gorm:"not null" json:"name"`
	Description string                    `gorm:"not null;default:''" json:"description,omitempty"`
	Price       float64                   `gorm:"not null" json:"price"`
	Duration    string                    `gorm:"not null" json:"duration"` // e.g., "daily", "monthly", "yearly"
	IsActive    bool                      `gorm:"not null;default:true" json:"is_active"`
	Menus       JSONB[ServicePackageMenu] `gorm:"type:jsonb;not null;default:'[]'" json:"menus,omitempty"`
	CreatedAt   time.Time                 `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time                 `gorm:"autoUpdateTime" json:"updated_at"`
}

// ServicePackageMenu describes a menu entry available within a service package.
type ServicePackageMenu struct {
	Title      string `json:"title"`
	Key        string `json:"key"`
	URL        string `json:"url"`
	Icon       string `json:"icon"`
	ParentID   int64  `json:"parent_id"`
	Permission int64  `json:"permission"`
	Active     bool   `json:"active"`
}

// CreateServicePackageInput is the request payload for creating a service package.
type CreateServicePackageInput struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Price       float64              `json:"price"`
	Duration    string               `json:"duration"`
	IsActive    bool                 `json:"is_active"`
	Menus       []ServicePackageMenu `json:"menus,omitempty"`
}

// UpdateServicePackageInput is the request payload for updating a service package.
type UpdateServicePackageInput struct {
	ID          string                `json:"id"`
	Name        *string               `json:"name,omitempty"`
	Description *string               `json:"description,omitempty"`
	Price       *float64              `json:"price,omitempty"`
	Duration    *string               `json:"duration,omitempty"`
	IsActive    *bool                 `json:"is_active,omitempty"`
	Menus       *[]ServicePackageMenu `json:"menus,omitempty"`
}
