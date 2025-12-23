package models

import "time"

// ServicePackageBasic is a lightweight projection for responses.
type ServicePackageBasic struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	StartAt time.Time `json:"start_at,omitempty"`
	EndAt   time.Time `json:"end_at,omitempty"`
}

// UnitDTO extends Unit with resolved service packages for read endpoints.
// Embedding Unit flattens its fields in JSON output.
type UnitDTO struct {
	Unit
	ServicePackages []ServicePackageBasic `json:"service_packages"`
}
