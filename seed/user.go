package seed

import (
	"context"
	"fmt"
	"go-fiber-api/config"
	"go-fiber-api/models"
	"go-fiber-api/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func SeedAdminUser() {
    collection := config.DB.Collection("users")
    var existing models.User
    err := collection.FindOne(context.TODO(), bson.M{"username": "admin"}).Decode(&existing)
    if err != mongo.ErrNoDocuments {
        fmt.Println("✅ Admin user already exists.")
        return
    }
    password, _ := utils.HashPassword("admin123")
    groupID, _ := primitive.ObjectIDFromHex("685d01ab5e17ba55d0e349f2")
    orgID, _ := primitive.ObjectIDFromHex(defaultOrgIDHex)
    admin := models.User{
        OrganizationID: orgID,
        Username:   "admin",
        Password:   password,
        Name:       "Administrator",
        UrlAvatar:  "",
        RoleGroups: []primitive.ObjectID{groupID},
    }

	_, err = collection.InsertOne(context.TODO(), admin)
	if err != nil {
		fmt.Println("❌ Failed to seed admin:", err)
		return
	}
	fmt.Println("🚀 Admin user seeded successfully: username=admin password=admin123")
}

// SeedDefaultUser creates a regular user account if not already present.
func SeedDefaultUser() {
    collection := config.DB.Collection("users")
    var existing models.User
    err := collection.FindOne(context.TODO(), bson.M{"username": "user"}).Decode(&existing)
    if err != mongo.ErrNoDocuments {
        fmt.Println("✅ Regular user already exists.")
        return
    }
    password, _ := utils.HashPassword("user123")
    orgID, _ := primitive.ObjectIDFromHex(defaultOrgIDHex)
    user := models.User{
        OrganizationID: orgID,
        Username:   "user",
        Password:   password,
        Name:       "Default User",
        UrlAvatar:  "",
        RoleGroups: []primitive.ObjectID{},
    }

	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		fmt.Println("❌ Failed to seed user:", err)
		return
	}
	fmt.Println("🚀 Regular user seeded successfully: username=user password=user123")
}
