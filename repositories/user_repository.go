package repositories

import (
	"context"
	"errors"
	"go-fiber-api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

// Tìm user theo username
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Tạo user mới
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = primitive.NewObjectID()
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

// Kiểm tra username đã tồn tại
func (r *UserRepository) IsUsernameExists(ctx context.Context, username string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"username": username})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetAll returns a paginated list of users filtered by username keyword.
// It also returns the total number of matched documents for pagination.
func (r *UserRepository) GetAll(ctx context.Context, search string, page, limit int64) ([]models.User, int64, error) {
	filter := bson.M{}
	if search != "" {
		filter["username"] = bson.M{"$regex": search, "$options": "i"}
	}

	projection := bson.M{"password": 0}
	opts := options.Find().SetProjection(projection)
	if limit > 0 {
		opts.SetLimit(limit).SetSkip((page - 1) * limit)
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Update password theo user ID
func (r *UserRepository) UpdatePassword(ctx context.Context, id string, hashedPassword string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"password": hashedPassword}}

	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}

// Lấy user theo ID
func (r *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var user models.User
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateByID updates the name, avatar URL, and role groups of a user by id and
// returns the updated document. Username and password cannot be changed here.
func (r *UserRepository) UpdateByID(ctx context.Context, id string, name string, urlAvatar string, roleGroups []primitive.ObjectID) (*models.User, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	update := bson.M{"$set": bson.M{
		"name":        name,
		"url_avatar":  urlAvatar,
		"role_groups": roleGroups,
	}}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated models.User
	err = r.collection.FindOneAndUpdate(ctx, bson.M{"_id": objID}, update, opts).Decode(&updated)
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	return &updated, nil
}
