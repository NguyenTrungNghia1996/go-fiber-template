package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UnitUser represents a user that belongs to a specific Unit (tenant).
type UnitUser struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	UnitID       primitive.ObjectID   `bson:"unit_id" json:"unit_id"`
	Username     string               `bson:"username" json:"username"`
	PasswordHash string               `bson:"password_hash" json:"-"`
	IsAdmin      bool                 `bson:"is_admin" json:"is_admin"`
	RoleGroupIDs []primitive.ObjectID `bson:"role_group_ids,omitempty" json:"role_group_ids,omitempty"`
	Name         string               `bson:"name,omitempty" json:"name,omitempty"`
	Email        string               `bson:"email,omitempty" json:"email,omitempty"`
	ImageURL     string               `bson:"image_url,omitempty" json:"image_url,omitempty"`
	CreatedAt    time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time            `bson:"updated_at" json:"updated_at"`
}
