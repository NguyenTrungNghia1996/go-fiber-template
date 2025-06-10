// package models
package models

type User struct {
	ID       string `json:"id" bson:"_id,omitempty"`
	Username string `json:"username" bson:"username"`
	Password string `json:"password,omitempty" bson:"password"`
	Email    string `json:"email" bson:"email"`
	Role     string `json:"role" bson:"role"`           // admin or member
	PersonID string `json:"person_id" bson:"person_id"` // ID của giáo viên
}
