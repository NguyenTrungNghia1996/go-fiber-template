package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// PermissionDetail represents a permission for a menu item.
type PermissionDetail struct {
	Key             string `json:"key" bson:"key"`
	PermissionValue int64  `json:"permissionValue" bson:"permissionValue"`
}

// RoleGroup defines a group of permissions.
type RoleGroup struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description []PermissionDetail `json:"description" bson:"description"`
}

// RoleGroupResponse is used when returning role groups to clients.
type RoleGroupResponse struct {
	ID          primitive.ObjectID `json:"id"`
	Name        string             `json:"name"`
	Description []PermissionDetail `json:"description"`
}

// ToResponse converts a RoleGroup to RoleGroupResponse.
func (r RoleGroup) ToResponse() RoleGroupResponse {
	return RoleGroupResponse{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
	}
}
