package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ServicePackage represents a service package offered by a unit.
type ServicePackage struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Price       float64            `bson:"price" json:"price"`
	Duration    string             `bson:"duration" json:"duration"` // e.g., "daily", "monthly", "yearly"
	IsActive    bool               `bson:"is_active" json:"is_active"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// CreateServicePackageInput is the request payload for creating a service package.
type CreateServicePackageInput struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Duration    string  `json:"duration"`
	IsActive    bool    `json:"is_active"`
}

// UpdateServicePackageInput is the request payload for updating a service package.
type UpdateServicePackageInput struct {
	ID          string   `json:"id"`
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Duration    *string  `json:"duration,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
}
