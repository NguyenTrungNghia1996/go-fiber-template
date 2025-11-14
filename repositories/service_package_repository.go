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

type ServicePackageRepository struct {
	coll *mongo.Collection
}

func NewServicePackageRepository(db *mongo.Database) *ServicePackageRepository {
	return &ServicePackageRepository{coll: db.Collection("service_packages")}
}

func (r *ServicePackageRepository) Create(ctx context.Context, sp *models.ServicePackage) error {
	now := time.Now().UTC()
	sp.ID = primitive.NilObjectID
	sp.CreatedAt = now
	sp.UpdatedAt = now
	res, err := r.coll.InsertOne(ctx, sp)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		sp.ID = oid
	}
	return nil
}

func (r *ServicePackageRepository) FindByID(ctx context.Context, id string) (*models.ServicePackage, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var sp models.ServicePackage
	if err := r.coll.FindOne(ctx, bson.D{{Key: "_id", Value: oid}}).Decode(&sp); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &sp, nil
}

// FindByIDs returns all service packages whose IDs are in the provided slice.
func (r *ServicePackageRepository) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]models.ServicePackage, error) {
	if len(ids) == 0 {
		return []models.ServicePackage{}, nil
	}
	filter := bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}}
	cur, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var items []models.ServicePackage
	for cur.Next(ctx) {
		var sp models.ServicePackage
		if err := cur.Decode(&sp); err != nil {
			return nil, err
		}
		items = append(items, sp)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return items, nil
}


func (r *ServicePackageRepository) FindPaged(ctx context.Context, page, limit int64, q string) ([]models.ServicePackage, int64, error) {
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
	var items []models.ServicePackage
	for cur.Next(ctx) {
		var sp models.ServicePackage
		if err := cur.Decode(&sp); err != nil {
			return nil, 0, err
		}
		items = append(items, sp)
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

func (r *ServicePackageRepository) UpdateByID(ctx context.Context, id string, updates bson.D) (*models.ServicePackage, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	updates = append(updates, bson.E{Key: "updated_at", Value: time.Now().UTC()})
	after := options.After
	var out models.ServicePackage
	err = r.coll.FindOneAndUpdate(ctx, bson.D{{Key: "_id", Value: oid}}, bson.D{{Key: "$set", Value: updates}}, options.FindOneAndUpdate().SetReturnDocument(after)).Decode(&out)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *ServicePackageRepository) DeleteByID(ctx context.Context, id string) (bool, error) {
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
