package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Menu represents a navigation entry.
type Menu struct {
    ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    Title         string             `json:"title" bson:"title"`
    Key           string             `json:"key" bson:"key"`
    URL           string             `json:"url" bson:"url"`
    Icon          string             `json:"icon" bson:"icon"`
    ParentID      primitive.ObjectID `json:"parent_Id" bson:"parent_Id"`
    PermissionBit int64              `json:"permissionBit" bson:"permissionBit"`
}

// MenuResponse is used when returning menus to clients.
type MenuResponse struct {
    ID            primitive.ObjectID  `json:"id"`
    Title         string              `json:"title"`
    Key           string              `json:"key"`
    URL           string              `json:"url"`
    Icon          string              `json:"icon"`
    ParentID      *primitive.ObjectID `json:"parent_Id"`
    PermissionBit int64               `json:"permissionBit"`
}

// ToResponse converts a Menu to a MenuResponse, setting ParentID to nil when it is zero.
func (m Menu) ToResponse() MenuResponse {
    var pid *primitive.ObjectID
    if !m.ParentID.IsZero() {
        pid = &m.ParentID
    }
    return MenuResponse{
        ID:            m.ID,
        Title:         m.Title,
        Key:           m.Key,
        URL:           m.URL,
        Icon:          m.Icon,
        ParentID:      pid,
        PermissionBit: m.PermissionBit,
    }
}
