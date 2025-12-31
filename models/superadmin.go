package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SuperAdmin represents a privileged administrative user.
type SuperAdmin struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Username     string               `bson:"username" json:"username"`
	PasswordHash string               `bson:"password_hash" json:"-"`
	IsAdmin      bool                 `bson:"is_admin" json:"is_admin"`
	Name         string               `bson:"name,omitempty" json:"name,omitempty"`
	Email        string               `bson:"email,omitempty" json:"email,omitempty"`
	ImageURL     string               `bson:"image_url,omitempty" json:"image_url,omitempty"`
	RoleGroupIDs []primitive.ObjectID `bson:"role_group_ids,omitempty" json:"role_group_ids,omitempty"`
	CreatedAt    time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time            `bson:"updated_at" json:"updated_at"`
}

// CreateSuperAdminInput is the request payload for creating a super admin.
type CreateSuperAdminInput struct {
	Username     string   `json:"username"`
	Password     string   `json:"password"`
	Name         string   `json:"name"`
	Email        string   `json:"email"`
	ImageURL     string   `json:"image_url"`
	IsAdmin      *bool    `json:"is_admin,omitempty"`
	RoleGroupIDs []string `json:"role_group_ids"`
}

// UpdateSuperAdminInput is the request payload for updating a super admin.
// All fields are optional; only provided fields will be updated.
type UpdateSuperAdminInput struct {
	ID           string    `json:"id"`
	Password     *string   `json:"password,omitempty"`
	Name         *string   `json:"name,omitempty"`
	Email        *string   `json:"email,omitempty"`
	ImageURL     *string   `json:"image_url,omitempty"`
	IsAdmin      *bool     `json:"is_admin,omitempty"`
	RoleGroupIDs *[]string `json:"role_group_ids,omitempty"`
}
