// package models
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID         primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Username   string               `json:"username" bson:"username"`
	Password   string               `json:"password,omitempty" bson:"password"`
	Name       string               `json:"name" bson:"name"`
	RoleGroups []primitive.ObjectID `json:"role_groups" bson:"role_groups"`
}

// UserListItem represents a user with role group details populated.
type UserListItem struct {
	ID         string              `json:"id"`
	Username   string              `json:"username"`
	Name       string              `json:"name"`
	RoleGroups []RoleGroupListItem `json:"role_groups"`
}

// ToListItem converts a User to UserListItem with role group details.
func (u User) ToListItem(groups map[primitive.ObjectID]RoleGroupListItem) UserListItem {
	rg := make([]RoleGroupListItem, 0, len(u.RoleGroups))
	for _, id := range u.RoleGroups {
		if g, ok := groups[id]; ok {
			rg = append(rg, g)
		}
	}
	return UserListItem{
		ID:         u.ID.Hex(),
		Username:   u.Username,
		Name:       u.Name,
		RoleGroups: rg,
	}
}
