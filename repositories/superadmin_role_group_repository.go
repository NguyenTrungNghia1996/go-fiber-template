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

// SuperAdminRoleGroupRepository manages role groups for super admins.
type SuperAdminRoleGroupRepository struct {
	coll *mongo.Collection
}

func NewSuperAdminRoleGroupRepository(db *mongo.Database) (*SuperAdminRoleGroupRepository, error) {
	r := &SuperAdminRoleGroupRepository{coll: db.Collection("super_admin_role_groups")}
	idx := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true).SetBackground(true),
	}
	if _, err := r.coll.Indexes().CreateOne(context.Background(), idx); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *SuperAdminRoleGroupRepository) Create(ctx context.Context, g *models.SuperAdminRoleGroup) error {
	now := time.Now().UTC()
	g.ID = primitive.NilObjectID
	g.CreatedAt = now
	g.UpdatedAt = now
	res, err := r.coll.InsertOne(ctx, g)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		g.ID = oid
	}
	return nil
}

func (r *SuperAdminRoleGroupRepository) FindByID(ctx context.Context, id string) (*models.SuperAdminRoleGroup, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var out models.SuperAdminRoleGroup
	err = r.coll.FindOne(ctx, bson.D{{Key: "_id", Value: oid}}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *SuperAdminRoleGroupRepository) FindPaged(ctx context.Context, page, limit int64, q string) ([]models.SuperAdminRoleGroup, int64, error) {
	filter := bson.D{}
	if q != "" {
		safe := regexp.QuoteMeta(q)
		rx := primitive.Regex{Pattern: safe, Options: "i"}
		filter = bson.D{{Key: "$or", Value: bson.A{
			bson.D{{Key: "name", Value: rx}},
			bson.D{{Key: "description", Value: rx}},
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
	var items []models.SuperAdminRoleGroup
	for cur.Next(ctx) {
		var it models.SuperAdminRoleGroup
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

func (r *SuperAdminRoleGroupRepository) UpdateByID(ctx context.Context, id string, updates bson.D) (*models.SuperAdminRoleGroup, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	updates = append(updates, bson.E{Key: "updated_at", Value: time.Now().UTC()})
	after := options.After
	var out models.SuperAdminRoleGroup
	err = r.coll.FindOneAndUpdate(ctx, bson.D{{Key: "_id", Value: oid}}, bson.D{{Key: "$set", Value: updates}}, options.FindOneAndUpdate().SetReturnDocument(after)).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *SuperAdminRoleGroupRepository) DeleteByID(ctx context.Context, id string) (bool, error) {
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

// FindExistingIDs returns a set of IDs that exist in the collection.
func (r *SuperAdminRoleGroupRepository) FindExistingIDs(ctx context.Context, ids []primitive.ObjectID) (map[primitive.ObjectID]struct{}, error) {
	out := make(map[primitive.ObjectID]struct{})
	if len(ids) == 0 {
		return out, nil
	}
	cur, err := r.coll.Find(ctx, bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var g models.SuperAdminRoleGroup
		if err := cur.Decode(&g); err != nil {
			return nil, err
		}
		out[g.ID] = struct{}{}
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// FindByIDs returns role groups matching the provided IDs (duplicates are ignored).
func (r *SuperAdminRoleGroupRepository) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]models.SuperAdminRoleGroup, error) {
	seen := make(map[primitive.ObjectID]struct{})
	unique := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		if id.IsZero() {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	if len(unique) == 0 {
		return []models.SuperAdminRoleGroup{}, nil
	}

	cur, err := r.coll.Find(ctx, bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: unique}}}})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []models.SuperAdminRoleGroup
	for cur.Next(ctx) {
		var g models.SuperAdminRoleGroup
		if err := cur.Decode(&g); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
