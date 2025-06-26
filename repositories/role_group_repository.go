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

type RoleGroupRepository struct {
	collection *mongo.Collection
}

func NewRoleGroupRepository(db *mongo.Database) *RoleGroupRepository {
	return &RoleGroupRepository{collection: db.Collection("role_groups")}
}

func (r *RoleGroupRepository) GetByID(ctx context.Context, id string) (*models.RoleGroup, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var group models.RoleGroup
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&group)
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("role group not found")
	}
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *RoleGroupRepository) Create(ctx context.Context, group *models.RoleGroup) error {
	group.ID = primitive.NewObjectID()
	_, err := r.collection.InsertOne(ctx, group)
	return err
}

func (r *RoleGroupRepository) GetAll(ctx context.Context, search string, page, limit int64) ([]models.RoleGroup, int64, error) {
	filter := bson.M{}
	if search != "" {
		filter["name"] = bson.M{"$regex": search, "$options": "i"}
	}
	findOpts := options.Find()
	if limit > 0 {
		findOpts.SetLimit(limit).SetSkip((page - 1) * limit)
	}
	cursor, err := r.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var groups []models.RoleGroup
	for cursor.Next(ctx) {
		var g models.RoleGroup
		if err := cursor.Decode(&g); err != nil {
			return nil, 0, err
		}
		groups = append(groups, g)
	}
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return groups, total, nil
}

func (r *RoleGroupRepository) UpdateByID(ctx context.Context, id string, group *models.RoleGroup) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update := bson.M{"$set": bson.M{
		"name":        group.Name,
		"description": group.Description,
	}}
	res, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("role group not found")
	}
	group.ID = objID
	return nil
}

func (r *RoleGroupRepository) DeleteByID(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("role group not found")
	}
	return nil
}
