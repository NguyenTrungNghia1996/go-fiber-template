package seed

import (
	"context"
	"fmt"

	"go-fiber-api/config"
	"go-fiber-api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// SeedMenus inserts default menu entries if they do not exist.
func SeedMenus() {
    collection := config.DB.Collection("menus")

    menus := []models.Menu{}

	id1, _ := primitive.ObjectIDFromHex("685d0554394165c3d5a0c625")
	menus = append(menus, models.Menu{
		ID:            id1,
		Title:         "Dashboard",
		Key:           "menu-euoi92n7f0",
		URL:           "/dashboard",
		Icon:          "ant-design:dashboard-outlined",
		ParentID:      primitive.NilObjectID,
		PermissionBit: 0,
	})

	id2, _ := primitive.ObjectIDFromHex("685d058f394165c3d5a0c626")
	menus = append(menus, models.Menu{
		ID:            id2,
		Title:         "Cài Đặt Quản Trị",
		Key:           "menu-byy4w5x6la",
		URL:           "/administration",
		Icon:          "ant-design:user-outlined",
		ParentID:      primitive.NilObjectID,
		PermissionBit: 2,
	})

	id3, _ := primitive.ObjectIDFromHex("685d05d3394165c3d5a0c627")
	menus = append(menus, models.Menu{
		ID:            id3,
		Title:         "Menu",
		Key:           "menu-1go08obucl",
		URL:           "/administration/menu",
		Icon:          "ant-design:menu-outlined",
		ParentID:      id2,
		PermissionBit: 0,
	})

	id4, _ := primitive.ObjectIDFromHex("685d0d98ba911d2a1d9f40ea")
	menus = append(menus, models.Menu{
		ID:            id4,
		Title:         "Nhóm Quyền",
		Key:           "menu-avmonb92aj",
		URL:           "/administration/role-group",
		Icon:          "ant-design:safety-outlined",
		ParentID:      id2,
		PermissionBit: 2,
	})

	id5, _ := primitive.ObjectIDFromHex("685e0007bd9eb34fceea7f4c")
    menus = append(menus, models.Menu{
        ID:            id5,
        Title:         "Người Dùng",
        Key:           "menu-w8bjq07960",
        URL:           "/administration/user",
        Icon:          "ant-design:user-outlined",
        ParentID:      id2,
        PermissionBit: 4,
    })

    // Organization management menu (shared across organizations)
    id6, _ := primitive.ObjectIDFromHex("68600107bd9eb34fceea7f4d")
    menus = append(menus, models.Menu{
        ID:            id6,
        Title:         "Tổ Chức",
        Key:           "menu-organization",
        URL:           "/administration/organization",
        Icon:          "ant-design:apartment-outlined",
        ParentID:      id2,
        PermissionBit: 8,
    })

	for _, m := range menus {
		var existing models.Menu
		err := collection.FindOne(context.TODO(), bson.M{"_id": m.ID}).Decode(&existing)
		if err == mongo.ErrNoDocuments {
			if _, err := collection.InsertOne(context.TODO(), m); err != nil {
				fmt.Println("❌ Failed to seed menu", m.Title, ":", err)
				continue
			}
			fmt.Println("🚀 Menu seeded:", m.Title)
		} else if err == nil {
			fmt.Println("✅ Menu already exists:", m.Title)
		} else {
			fmt.Println("❌ Failed checking menu", m.Title, ":", err)
		}
	}
}
