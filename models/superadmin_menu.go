package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SuperAdminMenu represents a menu item available to super admins.
type SuperAdminMenu struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title      string             `bson:"title" json:"title"`
	Key        string             `bson:"key" json:"key"`
	URL        string             `bson:"url" json:"url"`
	Icon       string             `bson:"icon" json:"icon"`
	ParentID   primitive.ObjectID `bson:"parent_id,omitempty" json:"parent_id,omitempty"`
	Permission int64              `bson:"permission" json:"permission"`
	Active     bool               `bson:"active" json:"active"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

// CreateSuperAdminMenuInput defines the payload for creating a super admin menu item.
type CreateSuperAdminMenuInput struct {
	Title      string `json:"title"`
	Key        string `json:"key"`
	URL        string `json:"url"`
	Icon       string `json:"icon"`
	ParentID   string `json:"parent_id,omitempty"`
	Permission int64  `json:"permission"`
	Active     bool   `json:"active"`
}

// UpdateSuperAdminMenuInput defines the payload for updating a super admin menu item.
type UpdateSuperAdminMenuInput struct {
	ID         string  `json:"id"`
	Title      *string `json:"title,omitempty"`
	Key        *string `json:"key,omitempty"`
	URL        *string `json:"url,omitempty"`
	Icon       *string `json:"icon,omitempty"`
	ParentID   *string `json:"parent_id,omitempty"`
	Permission *int64  `json:"permission,omitempty"`
	Active     *bool   `json:"active,omitempty"`
}
