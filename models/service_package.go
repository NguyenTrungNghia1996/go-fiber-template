package models

import (
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ObjectIDOrNil is an ObjectID that gracefully decodes legacy numeric/null values to a zero ObjectID.
type ObjectIDOrNil primitive.ObjectID

func (o ObjectIDOrNil) MarshalJSON() ([]byte, error) {
	return primitive.ObjectID(o).MarshalJSON()
}

func (o ObjectIDOrNil) MarshalBSONValue() (bsontype.Type, []byte, error) {
	oid := primitive.ObjectID(o)
	return bsontype.ObjectID, oid[:], nil
}

func (o *ObjectIDOrNil) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	switch t {
	case bsontype.Null, bsontype.Undefined:
		*o = ObjectIDOrNil(primitive.NilObjectID)
		return nil
	case bsontype.ObjectID:
		if len(data) != 12 {
			return fmt.Errorf("invalid objectid length %d", len(data))
		}
		var oid primitive.ObjectID
		copy(oid[:], data)
		*o = ObjectIDOrNil(oid)
		return nil
	case bsontype.String:
		rv := bson.RawValue{Type: t, Value: data}
		str := rv.StringValue()
		str = strings.TrimSpace(str)
		if str == "" {
			*o = ObjectIDOrNil(primitive.NilObjectID)
			return nil
		}
		oid, err := primitive.ObjectIDFromHex(str)
		if err != nil {
			return err
		}
		*o = ObjectIDOrNil(oid)
		return nil
	case bsontype.Int32, bsontype.Int64, bsontype.Double, bsontype.Boolean:
		*o = ObjectIDOrNil(primitive.NilObjectID)
		return nil
	default:
		return fmt.Errorf("cannot decode objectid from type %s", t)
	}
}

// ServicePackage represents a service package offered by a unit.
type ServicePackage struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name        string               `bson:"name" json:"name"`
	Description string               `bson:"description,omitempty" json:"description,omitempty"`
	Price       float64              `bson:"price" json:"price"`
	Duration    string               `bson:"duration" json:"duration"` // e.g., "daily", "monthly", "yearly"
	IsActive    bool                 `bson:"is_active" json:"is_active"`
	Menus       []ServicePackageMenu `bson:"menus,omitempty" json:"menus,omitempty"`
	CreatedAt   time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time            `bson:"updated_at" json:"updated_at"`
}

// ServicePackageMenu describes a menu entry available within a service package.
type ServicePackageMenu struct {
	ID         primitive.ObjectID `bson:"id,omitempty" json:"id"`
	Title      string             `bson:"title" json:"title"`
	Key        string             `bson:"key" json:"key"`
	URL        string             `bson:"url" json:"url"`
	Icon       string             `bson:"icon" json:"icon"`
	ParentID   ObjectIDOrNil      `bson:"parent_id,omitempty" json:"parent_id,omitempty"`
	Permission int64              `bson:"permission" json:"permission"`
	Active     bool               `bson:"active" json:"active"`
}

// ServicePackageMenuPayload captures incoming menu payloads for service packages.
type ServicePackageMenuPayload struct {
	ID         string `json:"id,omitempty"`
	Title      string `json:"title"`
	Key        string `json:"key"`
	URL        string `json:"url"`
	Icon       string `json:"icon"`
	ParentID   string `json:"parent_id,omitempty"`
	Permission int64  `json:"permission"`
	Active     bool   `json:"active"`
}

// CreateServicePackageInput is the request payload for creating a service package.
type CreateServicePackageInput struct {
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	Price       float64                     `json:"price"`
	Duration    string                      `json:"duration"`
	IsActive    bool                        `json:"is_active"`
	Menus       []ServicePackageMenuPayload `json:"menus,omitempty"`
}

// UpdateServicePackageInput is the request payload for updating a service package.
type UpdateServicePackageInput struct {
	ID          string                       `json:"id"`
	Name        *string                      `json:"name,omitempty"`
	Description *string                      `json:"description,omitempty"`
	Price       *float64                     `json:"price,omitempty"`
	Duration    *string                      `json:"duration,omitempty"`
	IsActive    *bool                        `json:"is_active,omitempty"`
	Menus       *[]ServicePackageMenuPayload `json:"menus,omitempty"`
}

// CreateServicePackageMenuInput is the request payload for creating a menu under a service package.
type CreateServicePackageMenuInput struct {
	ServicePackageID string `json:"service_package_id"`
	Title            string `json:"title"`
	Key              string `json:"key"`
	URL              string `json:"url"`
	Icon             string `json:"icon"`
	ParentID         string `json:"parent_id,omitempty"`
	Permission       int64  `json:"permission"`
	Active           bool   `json:"active"`
}

// UpdateServicePackageMenuInput is the request payload for updating a specific menu.
type UpdateServicePackageMenuInput struct {
	ID               string  `json:"id"`
	ServicePackageID string  `json:"service_package_id"`
	Title            *string `json:"title,omitempty"`
	Key              *string `json:"key,omitempty"`
	URL              *string `json:"url,omitempty"`
	Icon             *string `json:"icon,omitempty"`
	ParentID         *string `json:"parent_id,omitempty"`
	Permission       *int64  `json:"permission,omitempty"`
	Active           *bool   `json:"active,omitempty"`
}
