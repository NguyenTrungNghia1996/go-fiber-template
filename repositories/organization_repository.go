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

type OrganizationRepository struct {
    collection *mongo.Collection
}

func NewOrganizationRepository(db *mongo.Database) *OrganizationRepository {
    return &OrganizationRepository{collection: db.Collection("organizations")}
}

func (r *OrganizationRepository) GetByID(ctx context.Context, id string) (*models.Organization, error) {
    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, err
    }
    var org models.Organization
    err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&org)
    if err == mongo.ErrNoDocuments {
        return nil, errors.New("organization not found")
    }
    if err != nil {
        return nil, err
    }
    return &org, nil
}

func (r *OrganizationRepository) Create(ctx context.Context, org *models.Organization) error {
    org.ID = primitive.NewObjectID()
    _, err := r.collection.InsertOne(ctx, org)
    return err
}

func (r *OrganizationRepository) GetAll(ctx context.Context, search string, page, limit int64) ([]models.Organization, int64, error) {
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

    var orgs []models.Organization
    for cursor.Next(ctx) {
        var o models.Organization
        if err := cursor.Decode(&o); err != nil {
            return nil, 0, err
        }
        orgs = append(orgs, o)
    }
    total, err := r.collection.CountDocuments(ctx, filter)
    if err != nil {
        return nil, 0, err
    }
    return orgs, total, nil
}

func (r *OrganizationRepository) UpdateByID(ctx context.Context, id string, org *models.Organization) error {
    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return err
    }
    update := bson.M{"$set": bson.M{
        "name":        org.Name,
        "description": org.Description,
    }}
    res, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
    if err != nil {
        return err
    }
    if res.MatchedCount == 0 {
        return errors.New("organization not found")
    }
    org.ID = objID
    return nil
}

func (r *OrganizationRepository) DeleteByID(ctx context.Context, id string) error {
    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return err
    }
    res, err := r.collection.DeleteOne(ctx, bson.M{"_id": objID})
    if err != nil {
        return err
    }
    if res.DeletedCount == 0 {
        return errors.New("organization not found")
    }
    return nil
}

