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
    regRepo *UnitServicePackageRegistrationRepository
}

func NewUnitRepository(db *mongo.Database, regRepo *UnitServicePackageRegistrationRepository) (*UnitRepository, error) {
    r := &UnitRepository{coll: db.Collection("units"), regRepo: regRepo}    // Ensure unique index on subdomain
    idx := mongo.IndexModel{
        Keys:    bson.D{{Key: "subdomain", Value: 1}},
        Options: options.Index().SetUnique(true).SetBackground(true),
    }
    if _, err := r.coll.Indexes().CreateOne(context.Background(), idx); err != nil {
        return nil, err
    }
    return r, nil
}

func (r *UnitRepository) Create(ctx context.Context, u *models.Unit, servicePackageIDs []string) error {
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

    // Create registrations for service packages
    for _, spID := range servicePackageIDs {
        spOID, err := primitive.ObjectIDFromHex(spID)
        if err != nil {
            return err // Or handle as a bad request
        }
        reg := &models.UnitServicePackageRegistration{
            UnitID: u.ID,
            ServicePackageID: spOID,
        }
        if err := r.regRepo.Create(ctx, reg); err != nil {
            return err
        }
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

// FindBySubdomain returns a Unit by its subdomain or nil if not found.
func (r *UnitRepository) FindBySubdomain(ctx context.Context, subdomain string) (*models.Unit, error) {
    var u models.Unit
    if err := r.coll.FindOne(ctx, bson.D{{Key: "subdomain", Value: subdomain}}).Decode(&u); err != nil {
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

func (r *UnitRepository) UpdateByID(ctx context.Context, id string, updates bson.D, servicePackageIDs []string) (*models.Unit, error) {
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

    // Update service package registrations
    if servicePackageIDs != nil {
        // Get existing registrations for this unit
        existingRegs, _, err := r.regRepo.FindPaged(ctx, 0, 0, id, "") // page=0, limit=0 to get all
        if err != nil {
            return nil, err
        }

        existingServicePackageIDs := make(map[string]struct{})
        for _, reg := range existingRegs {
            existingServicePackageIDs[reg.ServicePackageID.Hex()] = struct{}{}
        }

        newServicePackageIDs := make(map[string]struct{})
        for _, spID := range servicePackageIDs {
            newServicePackageIDs[spID] = struct{}{}
        }

        // Service packages to add
        for spID := range newServicePackageIDs {
            if _, exists := existingServicePackageIDs[spID]; !exists {
                spOID, err := primitive.ObjectIDFromHex(spID)
                if err != nil {
                    return nil, err
                }
                reg := &models.UnitServicePackageRegistration{
                    UnitID: out.ID,
                    ServicePackageID: spOID,
                }
                if err := r.regRepo.Create(ctx, reg); err != nil {
                    return nil, err
                }
            }
        }

        // Service packages to remove
        for spID := range existingServicePackageIDs {
            if _, exists := newServicePackageIDs[spID]; !exists {
                // Find and delete the specific registration
                spOID, err := primitive.ObjectIDFromHex(spID)
                if err != nil {
                    return nil, err
                }
                filter := bson.D{{Key: "unit_id", Value: out.ID}, {Key: "service_package_id", Value: spOID}}
                _, err = r.regRepo.coll.DeleteMany(ctx, filter)
                if err != nil {
                    return nil, err
                }
            }
        }
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
