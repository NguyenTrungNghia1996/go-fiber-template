package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UnitServicePackageRegistration represents the registration of a service package by a unit.
type UnitServicePackageRegistration struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UnitID           primitive.ObjectID `bson:"unit_id" json:"unit_id"`
	ServicePackageID primitive.ObjectID `bson:"service_package_id" json:"service_package_id"`
	StartAt          time.Time          `bson:"start_at" json:"start_at"`
	EndAt            time.Time          `bson:"end_at" json:"end_at"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

// CreateUnitServicePackageRegistrationInput is the request payload for creating a registration.
type CreateUnitServicePackageRegistrationInput struct {
	UnitID           string `json:"unit_id"`
	ServicePackageID string `json:"service_package_id"`
	StartAt          string `json:"start_at,omitempty"` // RFC3339
	EndAt            string `json:"end_at,omitempty"`   // RFC3339
}

// UpdateUnitServicePackageRegistrationInput is the request payload for updating a registration.
type UpdateUnitServicePackageRegistrationInput struct {
	ID               string  `json:"id"`
	UnitID           *string `json:"unit_id,omitempty"`
	ServicePackageID *string `json:"service_package_id,omitempty"`
	StartAt          *string `json:"start_at,omitempty"` // RFC3339
	EndAt            *string `json:"end_at,omitempty"`   // RFC3339
}
