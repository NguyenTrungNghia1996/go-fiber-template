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

// SeedRoleGroups seeds default role groups if they don't already exist.
func SeedRoleGroups() {
	collection := config.DB.Collection("role_groups")

	adminID, _ := primitive.ObjectIDFromHex("685d01ab5e17ba55d0e349f2")
	adminGroup := models.RoleGroup{
		ID:          adminID,
		Name:        "admin",
		Description: "gioi thieu",
		Permission: []models.PermissionDetail{
			{Key: "menu", PermissionValue: 10},
			{Key: "menu-euoi92n7f0", PermissionValue: 0},
			{Key: "menu-byy4w5x6la", PermissionValue: 10},
		},
	}

	var existing models.RoleGroup
	err := collection.FindOne(context.TODO(), bson.M{"_id": adminID}).Decode(&existing)
	if err == mongo.ErrNoDocuments {
		if _, err := collection.InsertOne(context.TODO(), adminGroup); err != nil {
			fmt.Println("❌ Failed to seed role group:", err)
			return
		}
		fmt.Println("🚀 Role group seeded:", adminGroup.Name)
	} else if err == nil {
		fmt.Println("✅ Role group already exists:", adminGroup.Name)
	} else {
		fmt.Println("❌ Failed checking role group:", err)
	}
}
