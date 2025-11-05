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

type SuperAdminRepository struct {
    coll *mongo.Collection
}

func NewSuperAdminRepository(db *mongo.Database) (*SuperAdminRepository, error) {
    r := &SuperAdminRepository{coll: db.Collection("superadmins")}
    // Ensure unique index on username
    idx := mongo.IndexModel{
        Keys:    bson.D{{Key: "username", Value: 1}},
        Options: options.Index().SetUnique(true).SetBackground(true),
    }
    _, err := r.coll.Indexes().CreateOne(context.Background(), idx)
    if err != nil {
        return nil, err
    }
    return r, nil
}

func (r *SuperAdminRepository) Create(ctx context.Context, sa *models.SuperAdmin) error {
    now := time.Now().UTC()
    sa.ID = primitive.NilObjectID
    sa.CreatedAt = now
    sa.UpdatedAt = now
    res, err := r.coll.InsertOne(ctx, sa)
    if err != nil {
        return err
    }
    oid, ok := res.InsertedID.(primitive.ObjectID)
    if ok {
        sa.ID = oid
    }
    return nil
}

func (r *SuperAdminRepository) FindAll(ctx context.Context) ([]models.SuperAdmin, error) {
    cur, err := r.coll.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
    if err != nil {
        return nil, err
    }
    defer cur.Close(ctx)
    var out []models.SuperAdmin
    for cur.Next(ctx) {
        var sa models.SuperAdmin
        if err := cur.Decode(&sa); err != nil {
            return nil, err
        }
        out = append(out, sa)
    }
    return out, cur.Err()
}

func (r *SuperAdminRepository) FindPaged(ctx context.Context, page, limit int64, q string) ([]models.SuperAdmin, int64, error) {
    // Build filter for search across username, name, email using case-insensitive regex
    filter := bson.D{}
    if q != "" {
        safe := regexp.QuoteMeta(q)
        rx := primitive.Regex{Pattern: safe, Options: "i"}
        filter = bson.D{{Key: "$or", Value: bson.A{
            bson.D{{Key: "username", Value: rx}},
            bson.D{{Key: "name", Value: rx}},
            bson.D{{Key: "email", Value: rx}},
        }}}
    }
    // page==0 means return all data without pagination
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
    var out []models.SuperAdmin
    for cur.Next(ctx) {
        var sa models.SuperAdmin
        if err := cur.Decode(&sa); err != nil {
            return nil, 0, err
        }
        out = append(out, sa)
    }
    if err := cur.Err(); err != nil {
        return nil, 0, err
    }
    total, err := r.coll.CountDocuments(ctx, filter)
    if err != nil {
        return nil, 0, err
    }
    return out, total, nil
}

func (r *SuperAdminRepository) FindByID(ctx context.Context, id string) (*models.SuperAdmin, error) {
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, err
    }
    var sa models.SuperAdmin
    err = r.coll.FindOne(ctx, bson.D{{Key: "_id", Value: oid}}).Decode(&sa)
    if err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            return nil, nil
        }
        return nil, err
    }
    return &sa, nil
}

func (r *SuperAdminRepository) FindByUsername(ctx context.Context, username string) (*models.SuperAdmin, error) {
    var sa models.SuperAdmin
    err := r.coll.FindOne(ctx, bson.D{{Key: "username", Value: username}}).Decode(&sa)
    if err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            return nil, nil
        }
        return nil, err
    }
    return &sa, nil
}

func (r *SuperAdminRepository) UpdateByID(ctx context.Context, id string, updates bson.D) (*models.SuperAdmin, error) {
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, err
    }
    // Always bump updated_at
    updates = append(updates, bson.E{Key: "updated_at", Value: time.Now().UTC()})
    after := options.After
    opt := options.FindOneAndUpdate().SetReturnDocument(after)
    var updated models.SuperAdmin
    err = r.coll.FindOneAndUpdate(ctx, bson.D{{Key: "_id", Value: oid}}, bson.D{{Key: "$set", Value: updates}}, opt).Decode(&updated)
    if err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            return nil, nil
        }
        return nil, err
    }
    return &updated, nil
}

func (r *SuperAdminRepository) DeleteByID(ctx context.Context, id string) (bool, error) {
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
