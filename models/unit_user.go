package models

import "time"

// UnitUser represents a user that belongs to a specific Unit (tenant).
type UnitUser struct {
	ID           string    `gorm:"type:uuid;primaryKey" json:"id"`
	UnitID       string    `gorm:"type:uuid;not null;index:idx_unit_user,unique" json:"unit_id"`
	Username     string    `gorm:"not null;index:idx_unit_user,unique" json:"username"`
	PasswordHash string    `gorm:"not null" json:"-"`
	IsAdmin      bool      `gorm:"not null;default:false" json:"is_admin"`
	Name         string    `gorm:"not null;default:''" json:"name,omitempty"`
	Email        string    `gorm:"not null;default:''" json:"email,omitempty"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
