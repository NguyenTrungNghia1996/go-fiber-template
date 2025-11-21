package models

import "time"

// SuperAdmin represents a privileged administrative user.
type SuperAdmin struct {
	ID           string    `gorm:"type:uuid;primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Name         string    `gorm:"not null;default:''" json:"name,omitempty"`
	Email        string    `gorm:"not null;default:''" json:"email,omitempty"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// CreateSuperAdminInput is the request payload for creating a super admin.
type CreateSuperAdminInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

// UpdateSuperAdminInput is the request payload for updating a super admin.
// All fields are optional; only provided fields will be updated.
type UpdateSuperAdminInput struct {
	ID       string  `json:"id"`
	Password *string `json:"password,omitempty"`
	Name     *string `json:"name,omitempty"`
	Email    *string `json:"email,omitempty"`
}
