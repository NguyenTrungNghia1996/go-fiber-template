// package models
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username string             `json:"username" bson:"username"`
	Password string             `json:"password,omitempty" bson:"password"`
	Role     string             `json:"role" bson:"role"` // admin or member
}
