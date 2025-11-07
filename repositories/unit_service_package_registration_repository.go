package repositories

import (
	"context"
	"time"

	"go-fiber-api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UnitServicePackageRegistrationRepository struct {
	coll *mongo.Collection
}

func NewUnitServicePackageRegistrationRepository(db *mongo.Database) (*UnitServicePackageRegistrationRepository, error) {
	r := &UnitServicePackageRegistrationRepository{coll: db.Collection("unit_service_package_registrations")}
	// Ensure unique index on UnitID and ServicePackageID
	idx := mongo.IndexModel{
		Keys:    bson.D{{Key: "unit_id", Value: 1}, {Key: "service_package_id", Value: 1}},
		Options: options.Index().SetUnique(true).SetBackground(true),
	}
	_, err := r.coll.Indexes().CreateOne(context.Background(), idx)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *UnitServicePackageRegistrationRepository) Create(ctx context.Context, reg *models.UnitServicePackageRegistration) error {
	now := time.Now().UTC()
	reg.ID = primitive.NilObjectID
	reg.CreatedAt = now
	reg.UpdatedAt = now
	res, err := r.coll.InsertOne(ctx, reg)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		reg.ID = oid
	}
	return nil
}

func (r *UnitServicePackageRegistrationRepository) FindByID(ctx context.Context, id string) (*models.UnitServicePackageRegistration, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var reg models.UnitServicePackageRegistration
	if err := r.coll.FindOne(ctx, bson.D{{Key: "_id", Value: oid}}).Decode(&reg); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &reg, nil
}

func (r *UnitServicePackageRegistrationRepository) FindPaged(ctx context.Context, page, limit int64, unitID, servicePackageID string) ([]models.UnitServicePackageRegistration, int64, error) {
	filter := bson.D{}
	if unitID != "" {
		oid, err := primitive.ObjectIDFromHex(unitID)
		if err != nil {
			return nil, 0, err
		}
		filter = append(filter, bson.E{Key: "unit_id", Value: oid})
	}
	if servicePackageID != "" {
		oid, err := primitive.ObjectIDFromHex(servicePackageID)
		if err != nil {
			return nil, 0, err
		}
		filter = append(filter, bson.E{Key: "service_package_id", Value: oid})
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
	var items []models.UnitServicePackageRegistration
	for cur.Next(ctx) {
		var reg models.UnitServicePackageRegistration
		if err := cur.Decode(&reg); err != nil {
			return nil, 0, err
		}
		items = append(items, reg)
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

func (r *UnitServicePackageRegistrationRepository) UpdateByID(ctx context.Context, id string, updates bson.D) (*models.UnitServicePackageRegistration, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	updates = append(updates, bson.E{Key: "updated_at", Value: time.Now().UTC()})
	after := options.After
	var out models.UnitServicePackageRegistration
	err = r.coll.FindOneAndUpdate(ctx, bson.D{{Key: "_id", Value: oid}}, bson.D{{Key: "$set", Value: updates}}, options.FindOneAndUpdate().SetReturnDocument(after)).Decode(&out)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *UnitServicePackageRegistrationRepository) DeleteByID(ctx context.Context, id string) (bool, error) {
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
