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

// UnitRoleGroupRepository manages role groups scoped to a unit.
type UnitRoleGroupRepository struct {
	coll *mongo.Collection
}

func NewUnitRoleGroupRepository(db *mongo.Database) *UnitRoleGroupRepository {
	r := &UnitRoleGroupRepository{coll: db.Collection("unit_role_groups")}
	// Unique name within a unit
	idx := mongo.IndexModel{
		Keys:    bson.D{{Key: "unit_id", Value: 1}, {Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true).SetBackground(true),
	}
	_, _ = r.coll.Indexes().CreateOne(context.Background(), idx)
	return r
}

func (r *UnitRoleGroupRepository) Create(ctx context.Context, g *models.UnitRoleGroup) error {
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

func (r *UnitRoleGroupRepository) FindByIDWithinUnit(ctx context.Context, id string, unitID primitive.ObjectID) (*models.UnitRoleGroup, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var out models.UnitRoleGroup
	err = r.coll.FindOne(ctx, bson.D{{Key: "_id", Value: oid}, {Key: "unit_id", Value: unitID}}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *UnitRoleGroupRepository) FindPagedByUnit(ctx context.Context, unitID primitive.ObjectID, page, limit int64, q string) ([]models.UnitRoleGroup, int64, error) {
	filter := bson.D{{Key: "unit_id", Value: unitID}}
	if q != "" {
		rx := primitive.Regex{Pattern: regexp.QuoteMeta(q), Options: "i"}
		filter = append(filter, bson.E{Key: "$or", Value: bson.A{
			bson.D{{Key: "name", Value: rx}},
			bson.D{{Key: "description", Value: rx}},
		}})
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
	var items []models.UnitRoleGroup
	for cur.Next(ctx) {
		var it models.UnitRoleGroup
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

func (r *UnitRoleGroupRepository) UpdateByIDWithinUnit(ctx context.Context, id string, unitID primitive.ObjectID, updates bson.D) (*models.UnitRoleGroup, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	updates = append(updates, bson.E{Key: "updated_at", Value: time.Now().UTC()})
	after := options.After
	var out models.UnitRoleGroup
	err = r.coll.FindOneAndUpdate(ctx, bson.D{{Key: "_id", Value: oid}, {Key: "unit_id", Value: unitID}}, bson.D{{Key: "$set", Value: updates}}, options.FindOneAndUpdate().SetReturnDocument(after)).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *UnitRoleGroupRepository) DeleteByIDWithinUnit(ctx context.Context, id string, unitID primitive.ObjectID) (bool, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, err
	}
	res, err := r.coll.DeleteOne(ctx, bson.D{{Key: "_id", Value: oid}, {Key: "unit_id", Value: unitID}})
	if err != nil {
		return false, err
	}
	return res.DeletedCount > 0, nil
}

// FindExistingIDsInUnit returns which ids exist for the given unit.
func (r *UnitRoleGroupRepository) FindExistingIDsInUnit(ctx context.Context, unitID primitive.ObjectID, ids []primitive.ObjectID) (map[primitive.ObjectID]struct{}, error) {
	out := make(map[primitive.ObjectID]struct{})
	if len(ids) == 0 {
		return out, nil
	}
	cur, err := r.coll.Find(ctx, bson.D{{Key: "unit_id", Value: unitID}, {Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var g models.UnitRoleGroup
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

// FindByIDsWithinUnit returns role groups in the given unit that match the provided IDs (deduped).
func (r *UnitRoleGroupRepository) FindByIDsWithinUnit(ctx context.Context, unitID primitive.ObjectID, ids []primitive.ObjectID) ([]models.UnitRoleGroup, error) {
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
		return []models.UnitRoleGroup{}, nil
	}

	cur, err := r.coll.Find(ctx, bson.D{
		{Key: "unit_id", Value: unitID},
		{Key: "_id", Value: bson.D{{Key: "$in", Value: unique}}},
	})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []models.UnitRoleGroup
	for cur.Next(ctx) {
		var g models.UnitRoleGroup
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
