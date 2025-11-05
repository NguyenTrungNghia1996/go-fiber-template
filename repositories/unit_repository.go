package repositories

import (
    "context"
    "regexp"
    "time"

    "go-fiber-api/models"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type UnitRepository struct {
    coll *mongo.Collection
}

func NewUnitRepository(db *mongo.Database) (*UnitRepository, error) {
    r := &UnitRepository{coll: db.Collection("units")}
    // Ensure unique index on subdomain
    idx := mongo.IndexModel{
        Keys:    bson.D{{Key: "subdomain", Value: 1}},
        Options: options.Index().SetUnique(true).SetBackground(true),
    }
    if _, err := r.coll.Indexes().CreateOne(context.Background(), idx); err != nil {
        return nil, err
    }
    return r, nil
}

func (r *UnitRepository) Create(ctx context.Context, u *models.Unit) error {
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

func (r *UnitRepository) FindByID(ctx context.Context, id string) (*models.Unit, error) {
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, err
    }
    var u models.Unit
    if err := r.coll.FindOne(ctx, bson.D{{Key: "_id", Value: oid}}).Decode(&u); err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, nil
        }
        return nil, err
    }
    return &u, nil
}

func (r *UnitRepository) FindPaged(ctx context.Context, page, limit int64, q string) ([]models.Unit, int64, error) {
    // Build filter for search across code, name, description using regex
    filter := bson.D{}
    if q != "" {
        safe := regexp.QuoteMeta(q)
        rx := primitive.Regex{Pattern: safe, Options: "i"}
        filter = bson.D{{Key: "$or", Value: bson.A{
            bson.D{{Key: "subdomain", Value: rx}},
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
    var items []models.Unit
    for cur.Next(ctx) {
        var u models.Unit
        if err := cur.Decode(&u); err != nil {
            return nil, 0, err
        }
        items = append(items, u)
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

func (r *UnitRepository) UpdateByID(ctx context.Context, id string, updates bson.D) (*models.Unit, error) {
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, err
    }
    updates = append(updates, bson.E{Key: "updated_at", Value: time.Now().UTC()})
    after := options.After
    var out models.Unit
    err = r.coll.FindOneAndUpdate(ctx, bson.D{{Key: "_id", Value: oid}}, bson.D{{Key: "$set", Value: updates}}, options.FindOneAndUpdate().SetReturnDocument(after)).Decode(&out)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, nil
        }
        return nil, err
    }
    return &out, nil
}

func (r *UnitRepository) DeleteByID(ctx context.Context, id string) (bool, error) {
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
