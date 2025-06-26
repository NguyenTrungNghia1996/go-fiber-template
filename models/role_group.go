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
	Description string             `json:"description" bson:"description"`
	Permission  []PermissionDetail `json:"permission" bson:"permission"`
}

// RoleGroupResponse is used when returning role groups to clients.
type RoleGroupResponse struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Permission  []PermissionDetail `json:"permission"`
}

// RoleGroupListItem is a lightweight representation used when listing groups.
type RoleGroupListItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ToResponse converts a RoleGroup to RoleGroupResponse.
func (r RoleGroup) ToResponse() RoleGroupResponse {
	return RoleGroupResponse{
		ID:          r.ID.Hex(),
		Name:        r.Name,
		Description: r.Description,
		Permission:  r.Permission,
	}
}

// ToListItem converts a RoleGroup to RoleGroupListItem.
func (r RoleGroup) ToListItem() RoleGroupListItem {
	return RoleGroupListItem{
		ID:          r.ID.Hex(),
		Name:        r.Name,
		Description: r.Description,
	}
}
