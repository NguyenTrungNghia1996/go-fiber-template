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

var defaultOrgIDHex = "685f0aa0bd9eb34fceea7f4c"

// SeedOrganizations ensures a default organization exists for initial data.
func SeedOrganizations() {
    collection := config.DB.Collection("organizations")

    orgID, _ := primitive.ObjectIDFromHex(defaultOrgIDHex)
    defaultOrg := models.Organization{
        ID:          orgID,
        Name:        "Admin Organization",
        Description: "Seeded admin organization",
        Subdomain:   "admin",
    }

    var existing models.Organization
    err := collection.FindOne(context.TODO(), bson.M{"_id": orgID}).Decode(&existing)
    if err == mongo.ErrNoDocuments {
        if _, err := collection.InsertOne(context.TODO(), defaultOrg); err != nil {
            fmt.Println("❌ Failed to seed organization:", err)
            return
        }
        fmt.Println("🚀 Organization seeded:", defaultOrg.Name)
    } else if err == nil {
        fmt.Println("✅ Organization already exists:", defaultOrg.Name)
    } else {
        fmt.Println("❌ Failed checking organization:", err)
    }
}
