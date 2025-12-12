package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SuperAdminRoleGroup represents a permission group for super admins.
type SuperAdminRoleGroup struct {
	ID          primitive.ObjectID         `bson:"_id,omitempty" json:"id"`
	Name        string                     `bson:"name" json:"name"`
	Description string                     `bson:"description,omitempty" json:"description,omitempty"`
	Permissions []SuperAdminMenuPermission `bson:"permissions,omitempty" json:"permissions,omitempty"`
	CreatedAt   time.Time                  `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time                  `bson:"updated_at" json:"updated_at"`
}

// SuperAdminMenuPermission captures a permission entry for a menu key.
type SuperAdminMenuPermission struct {
	Key             string `bson:"key" json:"key"`
	PermissionValue int64  `bson:"permission_value" json:"permissionValue"`
}

// CreateSuperAdminRoleGroupInput is the payload for creating a role group.
type CreateSuperAdminRoleGroupInput struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Permissions []SuperAdminMenuPermission `json:"permissions"`
}

// UpdateSuperAdminRoleGroupInput is the payload for updating a role group.
type UpdateSuperAdminRoleGroupInput struct {
	ID          string                      `json:"id"`
	Name        *string                     `json:"name,omitempty"`
	Description *string                     `json:"description,omitempty"`
	Permissions *[]SuperAdminMenuPermission `json:"permissions,omitempty"`
}
