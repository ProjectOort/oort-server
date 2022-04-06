package repo

import (
	"context"

	"github.com/ProjectOort/oort-server/biz/asteroid"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// compile-time interface implementation check.
var _ asteroid.Repo = (*AsteroidRepo)(nil)

const (
	_AsteroidCollection = "asteroid"
)

type AsteroidRepo struct {
	_mongo *mongo.Database
	_neo4j neo4j.Driver
}

func NewAsteroidRepo(_mongo *mongo.Database, _neo4j neo4j.Driver) *AsteroidRepo {
	return &AsteroidRepo{
		_mongo: _mongo,
		_neo4j: _neo4j,
	}
}

func (x *AsteroidRepo) Create(ctx context.Context, a *asteroid.Asteroid, linkFromIDs []primitive.ObjectID, linkToIDs []primitive.ObjectID) error {
	_, err := x._mongo.Collection(_AsteroidCollection).InsertOne(ctx, a)
	if err != nil {
		return err
	}
	neo4jSession := x._neo4j.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer neo4jSession.Close()

	neo4jCallback := func(tx neo4j.Transaction) (interface{}, error) {
		createNodeCypher := "CREATE (a: Asteroid {id: $id, state: $state, authorId: $authorId, createdTime: $createdTime})"
		result, err := tx.Run(createNodeCypher, map[string]interface{}{
			"id":          a.ID.Hex(),
			"state":       a.State,
			"authorId":    a.AuthorID.Hex(),
			"createdTime": neo4j.LocalDateTimeOf(a.CreatedTime),
		})
		if err != nil {
			return nil, err
		}
		_, err = result.Consume()
		if err != nil {
			return nil, err
		}
		if len(linkFromIDs) != 0 {
			linkFrom := make([]string, len(linkFromIDs))
			for i, id := range linkFromIDs {
				linkFrom[i] = id.Hex()
			}

			createLinkCypher := "MATCH (from:Asteroid), (cur:Asteroid) " +
				"WHERE from.id IN $fromIds AND cur.id = $curId " +
				"CREATE (from)-[r:REFER]->(cur)"
			result, err := tx.Run(createLinkCypher, map[string]interface{}{
				"fromIds": linkFrom,
				"curId":   a.ID.Hex(),
			})
			if err != nil {
				return nil, err
			}
			_, err = result.Consume()
			if err != nil {
				return nil, err
			}
		}
		if len(linkToIDs) != 0 {
			linkTo := make([]string, len(linkToIDs))
			for i, id := range linkToIDs {
				linkTo[i] = id.Hex()
			}

			createLinkCypher := "MATCH (to:Asteroid), (cur:Asteroid) " +
				"WHERE to.id IN $toIds AND cur.id = $curId " +
				"CREATE (cur)-[r:REFER]->(to)"
			result, err := tx.Run(createLinkCypher, map[string]interface{}{
				"toIds": linkTo,
				"curId": a.ID.Hex(),
			})
			if err != nil {
				return nil, err
			}
			_, err = result.Consume()
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	}

	_, err = neo4jSession.WriteTransaction(neo4jCallback)
	return err
}

func (x *AsteroidRepo) createWithTx(ctx context.Context, a *asteroid.Asteroid) error {
	mongoSession, err := x._mongo.Client().StartSession()
	if err != nil {
		return err
	}
	defer mongoSession.EndSession(ctx)
	neo4jSession := x._neo4j.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})

	neo4jInsertionCallback := func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run("CREATE (a: Asteroid {id: $id, title: $title, authorId: $authorId})", map[string]interface{}{
			"id":       a.ID.Hex(),
			"title":    a.Title,
			"authorId": a.AuthorID.Hex(),
		})
		if err != nil {
			return nil, err
		}
		return result.Consume()
	}

	mongoInsertionCallback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		_, err := x._mongo.Collection("asteroid").InsertOne(sessCtx, a)
		if err != nil {
			return nil, err
		}
		return neo4jSession.WriteTransaction(neo4jInsertionCallback)
	}

	_, err = mongoSession.WithTransaction(ctx, mongoInsertionCallback)
	return err
}

func (x *AsteroidRepo) UpdateContent(ctx context.Context, a *asteroid.Asteroid) error {
	_, err := x._mongo.Collection("asteroid").UpdateByID(ctx, a.ID, bson.D{{
		"$set", bson.D{{
			"content", a.Content,
		}},
	}})
	return err
}

func (x *AsteroidRepo) Get(ctx context.Context, id primitive.ObjectID) (*asteroid.Asteroid, error) {
	a := new(asteroid.Asteroid)
	err := x._mongo.Collection("asteroid").FindOne(ctx, bson.D{
		{"_id", id},
		{"state", true},
	}).Decode(&a)
	return a, err
}

func (x *AsteroidRepo) ListHub(ctx context.Context, authorID primitive.ObjectID) ([]*asteroid.Asteroid, error) {
	cursor, err := x._mongo.Collection("asteroid").Find(ctx, bson.D{
		{"author_id", authorID},
		{"state", true},
		{"hub", true},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	as := make([]*asteroid.Asteroid, 0)
	for cursor.Next(ctx) {
		var a asteroid.Asteroid
		err := cursor.Decode(&a)
		if err != nil {
			return nil, err
		}
		as = append(as, &a)
	}
	return as, nil
}

func (x *AsteroidRepo) List(ctx context.Context, aIDs []primitive.ObjectID) ([]*asteroid.Asteroid, error) {
	result, err := x._mongo.Collection("asteroid").Find(ctx, bson.D{
		{"_id", bson.D{
			{"$in", aIDs},
		}},
	})
	if err != nil {
		return nil, err
	}
	as := make([]*asteroid.Asteroid, 0)
	for result.Next(ctx) {
		var a asteroid.Asteroid
		err := result.Decode(&a)
		if err != nil {
			return nil, err
		}
		as = append(as, &a)
	}
	return as, err
}

func (x *AsteroidRepo) BatchExist(ctx context.Context, aIDs []primitive.ObjectID) (bool, error) {
	result, err := x._mongo.Collection("asteroid").CountDocuments(ctx, bson.D{
		{"_id", bson.D{
			{"$in", aIDs},
		}},
	})
	if err != nil {
		return false, err
	}
	return result == int64(len(aIDs)), nil
}

func (x *AsteroidRepo) ListLinkedFrom(ctx context.Context, id primitive.ObjectID) ([]*asteroid.Asteroid, error) {
	session := x._neo4j.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close()

	cypher := "MATCH (a:Asteroid)-[:REFER]->(:Asteroid {id: $id}) RETURN a"

	result, err := session.Run(cypher, map[string]interface{}{"id": id.Hex()})
	if err != nil {
		return nil, err
	}

	// collect ids
	ids := make([]primitive.ObjectID, 0)
	for result.Next() {
		nodeI, _ := result.Record().Get("a")
		node := nodeI.(neo4j.Node)

		id, _ := primitive.ObjectIDFromHex(node.Props["id"].(string))
		ids = append(ids, id)
	}

	mongoResult, err := x._mongo.Collection(_AsteroidCollection).Find(ctx, bson.D{{
		"_id", bson.D{{
			"$in", ids,
		}},
	}}, options.Find().SetProjection(bson.D{
		{"content", 0},
	}))
	as := make([]*asteroid.Asteroid, 0, len(ids))
	for mongoResult.Next(ctx) {
		var a asteroid.Asteroid
		err := mongoResult.Decode(&a)
		if err != nil {
			return nil, err
		}
		as = append(as, &a)
	}
	return as, nil
}

func (x *AsteroidRepo) ListLinkedTo(ctx context.Context, id primitive.ObjectID) ([]*asteroid.Asteroid, error) {
	session := x._neo4j.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close()

	cypher := "MATCH (:Asteroid {id: $id})-[:REFER]->(a:Asteroid) RETURN a"

	result, err := session.Run(cypher, map[string]interface{}{"id": id.Hex()})
	if err != nil {
		return nil, err
	}
	ids := make([]primitive.ObjectID, 0)
	for result.Next() {
		nodeI, _ := result.Record().Get("a")
		node := nodeI.(neo4j.Node)

		id, _ := primitive.ObjectIDFromHex(node.Props["id"].(string))
		ids = append(ids, id)
	}

	mongoResult, err := x._mongo.Collection(_AsteroidCollection).Find(ctx, bson.D{{
		"_id", bson.D{{
			"$in", ids,
		}},
	}}, options.Find().SetProjection(bson.D{
		{"content", 0},
	}))
	as := make([]*asteroid.Asteroid, 0, len(ids))
	for mongoResult.Next(ctx) {
		var a asteroid.Asteroid
		err := mongoResult.Decode(&a)
		if err != nil {
			return nil, err
		}
		as = append(as, &a)
	}

	return as, nil
}
