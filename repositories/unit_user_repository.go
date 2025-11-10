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

// UnitUserRepository manages users that belong to units (tenants).
type UnitUserRepository struct {
    coll *mongo.Collection
}

func NewUnitUserRepository(db *mongo.Database) *UnitUserRepository {
    r := &UnitUserRepository{coll: db.Collection("unit_users")}
    // Unique index on (unit_id, username)
    idx := mongo.IndexModel{
        Keys:    bson.D{{Key: "unit_id", Value: 1}, {Key: "username", Value: 1}},
        Options: options.Index().SetUnique(true).SetBackground(true),
    }
    _, _ = r.coll.Indexes().CreateOne(context.Background(), idx)
    return r
}

func (r *UnitUserRepository) Create(ctx context.Context, u *models.UnitUser) error {
    now := time.Now().UTC()
    u.ID = primitive.NilObjectID
    u.CreatedAt = now
    u.UpdatedAt = now
    res, err := r.coll.InsertOne(ctx, u)
    if err != nil {
        return err
    }
    if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
        u.ID = oid
    }
    return nil
}

func (r *UnitUserRepository) FindByUsernameAndUnitID(ctx context.Context, username string, unitID primitive.ObjectID) (*models.UnitUser, error) {
    var u models.UnitUser
    err := r.coll.FindOne(ctx, bson.D{{Key: "username", Value: username}, {Key: "unit_id", Value: unitID}}).Decode(&u)
    if err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            return nil, nil
        }
        return nil, err
    }
    return &u, nil
}

// FindByIDWithinUnit returns a user by id scoped to a unit.
func (r *UnitUserRepository) FindByIDWithinUnit(ctx context.Context, id string, unitID primitive.ObjectID) (*models.UnitUser, error) {
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, err
    }
    var u models.UnitUser
    err = r.coll.FindOne(ctx, bson.D{{Key: "_id", Value: oid}, {Key: "unit_id", Value: unitID}}).Decode(&u)
    if err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            return nil, nil
        }
        return nil, err
    }
    return &u, nil
}

// FindPagedByUnit lists users within a unit with optional search and pagination.
func (r *UnitUserRepository) FindPagedByUnit(ctx context.Context, unitID primitive.ObjectID, page, limit int64, q string) ([]models.UnitUser, int64, error) {
    filter := bson.D{{Key: "unit_id", Value: unitID}}
    if q != "" {
        // search username, name, email (case-insensitive, escaped)
        safe := primitive.Regex{Pattern: regexp.QuoteMeta(q), Options: "i"}
        filter = append(filter, bson.E{Key: "$or", Value: bson.A{
            bson.D{{Key: "username", Value: safe}},
            bson.D{{Key: "name", Value: safe}},
            bson.D{{Key: "email", Value: safe}},
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
    var items []models.UnitUser
    for cur.Next(ctx) {
        var it models.UnitUser
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

// UpdateByIDWithinUnit updates a user within a unit and returns the updated doc.
func (r *UnitUserRepository) UpdateByIDWithinUnit(ctx context.Context, id string, unitID primitive.ObjectID, updates bson.D) (*models.UnitUser, error) {
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, err
    }
    updates = append(updates, bson.E{Key: "updated_at", Value: time.Now().UTC()})
    after := options.After
    var out models.UnitUser
    err = r.coll.FindOneAndUpdate(ctx, bson.D{{Key: "_id", Value: oid}, {Key: "unit_id", Value: unitID}}, bson.D{{Key: "$set", Value: updates}}, options.FindOneAndUpdate().SetReturnDocument(after)).Decode(&out)
    if err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            return nil, nil
        }
        return nil, err
    }
    return &out, nil
}

// DeleteByIDWithinUnit deletes a user within a unit and returns whether a document was deleted.
func (r *UnitUserRepository) DeleteByIDWithinUnit(ctx context.Context, id string, unitID primitive.ObjectID) (bool, error) {
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

// AnyAdminExists checks whether there is at least one admin user in the unit.
func (r *UnitUserRepository) AnyAdminExists(ctx context.Context, unitID primitive.ObjectID) (bool, error) {
    count, err := r.coll.CountDocuments(ctx, bson.D{{Key: "unit_id", Value: unitID}, {Key: "is_admin", Value: true}})
    if err != nil {
        return false, err
    }
    return count > 0, nil
}
