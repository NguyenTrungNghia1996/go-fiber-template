package repositories

import (
	"context"
	"errors"
	"regexp"
	"time"

	"go-fiber-api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SuperAdminMenuRepository manages super admin menu items.
type SuperAdminMenuRepository struct {
	coll *mongo.Collection
}

func NewSuperAdminMenuRepository(db *mongo.Database) (*SuperAdminMenuRepository, error) {
	r := &SuperAdminMenuRepository{coll: db.Collection("super_admin_menus")}
	// Ensure unique index on key for stable lookups
	idx := mongo.IndexModel{
		Keys:    bson.D{{Key: "key", Value: 1}},
		Options: options.Index().SetUnique(true).SetBackground(true),
	}
	if _, err := r.coll.Indexes().CreateOne(context.Background(), idx); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *SuperAdminMenuRepository) Create(ctx context.Context, menu *models.SuperAdminMenu) error {
	now := time.Now().UTC()
	menu.ID = primitive.NilObjectID
	menu.CreatedAt = now
	menu.UpdatedAt = now
	res, err := r.coll.InsertOne(ctx, menu)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		menu.ID = oid
	}
	return nil
}

func (r *SuperAdminMenuRepository) FindByID(ctx context.Context, id string) (*models.SuperAdminMenu, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var out models.SuperAdminMenu
	err = r.coll.FindOne(ctx, bson.D{{Key: "_id", Value: oid}}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

// FindByKey performs a case-insensitive lookup by key.
func (r *SuperAdminMenuRepository) FindByKey(ctx context.Context, key string) (*models.SuperAdminMenu, error) {
	filter := bson.D{{Key: "key", Value: primitive.Regex{Pattern: "^" + regexp.QuoteMeta(key) + "$", Options: "i"}}}
	var out models.SuperAdminMenu
	err := r.coll.FindOne(ctx, filter).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

// FindPaged returns menus with optional pagination and text search.
func (r *SuperAdminMenuRepository) FindPaged(ctx context.Context, page, limit int64, q string) ([]models.SuperAdminMenu, int64, error) {
	filter := bson.D{}
	if q != "" {
		rx := primitive.Regex{Pattern: regexp.QuoteMeta(q), Options: "i"}
		filter = bson.D{{Key: "$or", Value: bson.A{
			bson.D{{Key: "title", Value: rx}},
			bson.D{{Key: "key", Value: rx}},
			bson.D{{Key: "url", Value: rx}},
		}}}
	}
	findOpt := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if page > 0 {
		if limit < 1 {
			limit = 10
		}
		skip := (page - 1) * limit
		findOpt.SetSkip(skip).SetLimit(limit)
	}

	cur, err := r.coll.Find(ctx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)

	var items []models.SuperAdminMenu
	for cur.Next(ctx) {
		var it models.SuperAdminMenu
		if err := cur.Decode(&it); err != nil {
			return nil, 0, err
		}
		items = append(items, it)
	}
	if err := cur.Err(); err != nil {
		return nil, 0, err
	}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *SuperAdminMenuRepository) UpdateByID(ctx context.Context, id string, updates bson.D) (*models.SuperAdminMenu, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	updates = append(updates, bson.E{Key: "updated_at", Value: time.Now().UTC()})
	after := options.After
	var out models.SuperAdminMenu
	err = r.coll.FindOneAndUpdate(ctx, bson.D{{Key: "_id", Value: oid}}, bson.D{{Key: "$set", Value: updates}}, options.FindOneAndUpdate().SetReturnDocument(after)).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *SuperAdminMenuRepository) DeleteByID(ctx context.Context, id string) (bool, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, err
	}
	res, err := r.coll.DeleteOne(ctx, bson.D{{Key: "_id", Value: oid}})
	if err != nil {
		return false, err
	}
	return res.DeletedCount > 0, nil
}

// CountByIDs returns how many menu IDs exist from the provided list.
func (r *SuperAdminMenuRepository) CountByIDs(ctx context.Context, ids []primitive.ObjectID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	return r.coll.CountDocuments(ctx, bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}})
}
