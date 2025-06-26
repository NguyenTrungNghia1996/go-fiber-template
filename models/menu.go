package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Menu struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title         string             `json:"title" bson:"title"`
	Key           string             `json:"key" bson:"key"`
	URL           string             `json:"url" bson:"url"`
	Icon          string             `json:"icon" bson:"icon"`
	ParentID      primitive.ObjectID `json:"parent_Id" bson:"parent_Id"`
	PermissionBit int64              `json:"permissionBit" bson:"permissionBit"`
}
