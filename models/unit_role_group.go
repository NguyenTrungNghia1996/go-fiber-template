package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UnitRoleGroup represents a permission group scoped to a single unit (tenant).
type UnitRoleGroup struct {
	ID          primitive.ObjectID         `bson:"_id,omitempty" json:"id"`
	UnitID      primitive.ObjectID         `bson:"unit_id" json:"unit_id"`
	Name        string                     `bson:"name" json:"name"`
	Description string                     `bson:"description,omitempty" json:"description,omitempty"`
	Permissions []SuperAdminMenuPermission `bson:"permissions,omitempty" json:"permissions,omitempty"`
	CreatedAt   time.Time                  `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time                  `bson:"updated_at" json:"updated_at"`
}

// CreateUnitRoleGroupInput is the payload for creating a unit role group.
type CreateUnitRoleGroupInput struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Permissions []SuperAdminMenuPermission `json:"permissions"`
}

// UpdateUnitRoleGroupInput is the payload for updating a unit role group.
type UpdateUnitRoleGroupInput struct {
	ID          string                      `json:"id"`
	Name        *string                     `json:"name,omitempty"`
	Description *string                     `json:"description,omitempty"`
	Permissions *[]SuperAdminMenuPermission `json:"permissions,omitempty"`
}
