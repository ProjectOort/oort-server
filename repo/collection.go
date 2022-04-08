package repo

import (
	"context"
	"fmt"
	"github.com/ProjectOort/oort-server/biz/asteroid"
	"github.com/ProjectOort/oort-server/biz/collection"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// compile-time interface implementation check.
var _ collection.Repo = (*CollectionRepo)(nil)

var (
	_CollectionCollection = "collection"
)

type CollectionRepo struct {
	_mongo *mongo.Database
}

func NewCollectionRepo(_mongo *mongo.Database) *CollectionRepo {
	return &CollectionRepo{_mongo: _mongo}
}

func (x *CollectionRepo) Get(ctx context.Context, collectionID primitive.ObjectID) (*collection.Collection, error) {
	var c collection.Collection
	err := x._mongo.Collection(_CollectionCollection).
		FindOne(ctx, bson.D{{"_id", collectionID}}).Decode(&c)
	return &c, err
}

func (x *CollectionRepo) Create(ctx context.Context, collection *collection.Collection) error {
	_, err := x._mongo.Collection(_CollectionCollection).InsertOne(ctx, collection)
	return err
}

func (x *CollectionRepo) Update(ctx context.Context, collection *collection.Collection) (*collection.Collection, error) {
	var updateFields bson.D
	updateFields = append(updateFields, bson.E{Key: "updated_time", Value: collection.UpdatedTime})
	if collection.Name != "" {
		updateFields = append(updateFields, bson.E{Key: "name", Value: collection.Name})
	}
	if collection.Description != "" {
		updateFields = append(updateFields, bson.E{Key: "description", Value: collection.Description})
	}
	result, err := x._mongo.Collection(_CollectionCollection).
		UpdateByID(ctx, collection.ID, bson.D{{"$set", updateFields}})
	fmt.Println(result.UpsertedCount)
	return nil, err
}

func (x *CollectionRepo) Delete(ctx context.Context, collectionID primitive.ObjectID) error {
	result, err := x._mongo.Collection(_CollectionCollection).
		UpdateByID(ctx, collectionID, bson.D{{"$set", bson.D{{"state", false}}}})
	fmt.Println(result.UpsertedCount)
	return err
}

func (x *CollectionRepo) List(ctx context.Context, ownerID primitive.ObjectID) ([]*collection.Collection, error) {
	result, err := x._mongo.Collection(_CollectionCollection).
		Find(ctx, bson.D{{"owner_id", ownerID}, {"state", true}},
			options.Find().SetProjection(bson.D{{"items", 0}}))
	if err != nil {
		return nil, err
	}
	cols := make([]*collection.Collection, 0)
	for result.Next(ctx) {
		var col collection.Collection
		err := result.Decode(&col)
		if err != nil {
			return nil, err
		}
		cols = append(cols, &col)
	}
	return cols, nil
}

func (x *CollectionRepo) PushItem(ctx context.Context, collectionID primitive.ObjectID, itemID primitive.ObjectID) error {
	_, err := x._mongo.Collection(_CollectionCollection).
		UpdateByID(ctx, collectionID,
			bson.D{{"$push", bson.D{
				{"items", itemID}},
			}},
		)
	return err
}

func (x *CollectionRepo) PopItem(ctx context.Context, collectionID primitive.ObjectID, itemID primitive.ObjectID) error {
	_, err := x._mongo.Collection(_CollectionCollection).UpdateOne(ctx, bson.D{
		{"_id", collectionID},
	}, bson.D{
		{"$pull", bson.D{
			{"items", itemID},
		}},
	})
	return err
}

func (x *CollectionRepo) ListItems(ctx context.Context, collectionID primitive.ObjectID) ([]*collection.Item, error) {
	result := x._mongo.Collection(_CollectionCollection).
		FindOne(ctx, bson.D{{"_id", collectionID}}, options.FindOne().SetProjection(bson.D{{"items", 1}}))
	if result.Err() != nil {
		return nil, result.Err()
	}
	var col collection.Collection
	err := result.Decode(&col)
	if err != nil {
		return nil, err
	}
	if len(col.Items) == 0 {
		return []*collection.Item{}, nil
	}

	asteroidsResult, err := x._mongo.Collection(_AsteroidCollection).
		Find(ctx, bson.D{{"_id", bson.D{{"$in", col.Items}}}},
			options.Find().SetProjection(bson.D{{"content", 0}}))
	if err != nil {
		return nil, err
	}
	items := make([]*collection.Item, 0)
	for asteroidsResult.Next(ctx) {
		var a asteroid.Asteroid
		err := asteroidsResult.Decode(&a)
		if err != nil {
			return nil, err
		}
		items = append(items, &collection.Item{Asteroid: a})
	}
	return items, nil
}
