package repositories

import (
	"context"

	"go-fiber-api/config"
	"go-fiber-api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Tìm user theo username
func FindUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := config.DB.Collection("users").FindOne(context.TODO(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Tạo user mới và gắn với giáo viên (PersonID)
func CreateUser(user *models.User) error {
	user.ID = primitive.NewObjectID()
	_, err := config.DB.Collection("users").InsertOne(context.TODO(), user)
	return err
}

// Kiểm tra Username already exists
func IsUsernameExists(username string) (bool, error) {
	count, err := config.DB.Collection("users").CountDocuments(context.TODO(), bson.M{"username": username})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Lấy danh sách user theo role (nếu có)
func GetUsersByRole(role string) ([]models.User, error) {
	filter := bson.M{}
	if role != "" {
		filter["role"] = role
	}

	// Projection: loại bỏ trường password
	projection := bson.M{
		"password": 0, // 0 = không lấy trường này
	}

	opts := options.Find().SetProjection(projection)

	cursor, err := config.DB.Collection("users").Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var users []models.User
	for cursor.Next(context.TODO()) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
func UpdateUserPersonID(id string, personID string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"person_id": personID}}
	_, err = config.DB.Collection("users").UpdateOne(context.TODO(), filter, update)
	return err
}

func UpdateUserPassword(id string, hashedPassword string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"password": hashedPassword}}
	_, err = config.DB.Collection("users").UpdateOne(context.TODO(), filter, update)
	return err
}

// Lấy user theo ID
func FindUserByID(id string) (*models.User, error) {
	var user models.User
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = config.DB.Collection("users").FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
