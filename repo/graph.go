package repo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/ProjectOort/oort-server/biz/asteroid"
	"github.com/ProjectOort/oort-server/biz/graph"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type GraphRepo struct {
	_mongo *mongo.Database
	_neo4j neo4j.Driver
}

func NewGraphRepo(_mongo *mongo.Database, _neo4j neo4j.Driver) *GraphRepo {
	return &GraphRepo{
		_mongo: _mongo,
		_neo4j: _neo4j,
	}
}

func (x *GraphRepo) GetGraphByAsteroidID(ctx context.Context, astID primitive.ObjectID, depth int) (*graph.Graph, error) {
	session := x._neo4j.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close()

	cypher := fmt.Sprintf("MATCH p = (:Asteroid {id: $id})-[:REFER*1..%d]-(:Asteroid) RETURN p", depth)

	result, err := session.Run(cypher, map[string]interface{}{
		"id":    astID.Hex(),
		"depth": depth,
	})
	if err != nil {
		return nil, err
	}

	var g graph.Graph
	nodeSet := make(map[int64]primitive.ObjectID)
	type linkPair struct {
		source int64
		target int64
	}
	linkSet := make(map[linkPair]struct{})

	for result.Next() {
		_p, _ := result.Record().Get("p")
		p := _p.(neo4j.Path)
		for _, node := range p.Nodes {
			idHex, err := primitive.ObjectIDFromHex(node.Props["id"].(string))
			if err != nil {
				return nil, err
			}
			nodeSet[node.Id] = idHex
		}
		for _, rel := range p.Relationships {
			linkSet[linkPair{source: rel.StartId, target: rel.EndId}] = struct{}{}
		}
	}

	for link := range linkSet {
		g.Links = append(g.Links,
			graph.Link{
				Source: nodeSet[link.source].Hex(),
				Target: nodeSet[link.target].Hex(),
			},
		)
	}

	ids := make([]primitive.ObjectID, 0, len(nodeSet))
	for _, id := range nodeSet {
		ids = append(ids, id)
	}

	mongoResult, err := x._mongo.Collection(_AsteroidCollection).Find(ctx, bson.D{{
		"_id", bson.D{{
			"$in", ids,
		}},
	}})
	if err != nil {
		return nil, err
	}

	for mongoResult.Next(ctx) {
		var ast asteroid.Asteroid
		err := mongoResult.Decode(&ast)
		if err != nil {
			return nil, err
		}
		g.Nodes = append(g.Nodes, graph.Node{
			ID:    ast.ID.Hex(),
			Hub:   ast.Hub,
			Title: ast.Title,
		})
	}

	if len(g.Nodes) == 0 {
		ast := new(asteroid.Asteroid)
		err := x._mongo.Collection("asteroid").FindOne(ctx, bson.D{
			{"_id", astID},
			{"state", true},
		}).Decode(&ast)
		if err != nil {
			return nil, err
		}
		g.Nodes = append(g.Nodes, graph.Node{
			ID:    ast.ID.Hex(),
			Hub:   ast.Hub,
			Title: ast.Title,
		})
	}
	return &g, nil
}

func (x *GraphRepo) GetFullGraph(ctx context.Context, accID primitive.ObjectID) (*graph.Graph, error) {
	nodeResult, err := x._mongo.Collection(_AsteroidCollection).Find(ctx, bson.D{
		{"author_id", accID},
		{"state", true},
	}, options.Find().SetProjection(bson.D{
		{"content", 0},
	}))
	if err != nil {
		return nil, err
	}
	var g graph.Graph
	for nodeResult.Next(ctx) {
		var ast asteroid.Asteroid
		err := nodeResult.Decode(&ast)
		if err != nil {
			return nil, err
		}
		g.Nodes = append(g.Nodes, graph.Node{
			ID:    ast.ID.Hex(),
			Hub:   ast.Hub,
			Title: ast.Title,
		})
	}

	session := x._neo4j.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close()

	cypher := "MATCH (a1:Asteroid)-[:REFER]->(a2:Asteroid) " +
		"WHERE a1.authorId=$authorId AND a2.authorId=$authorId " +
		"RETURN a1, a2"

	linkResult, err := session.Run(cypher, map[string]interface{}{
		"authorId": accID.Hex(),
	})
	if err != nil {
		return nil, err
	}
	for linkResult.Next() {
		_a1_, _ := linkResult.Record().Get("a1")
		_a2_, _ := linkResult.Record().Get("a2")
		a1 := _a1_.(neo4j.Node)
		a2 := _a2_.(neo4j.Node)
		a1ID := a1.Props["id"].(string)
		a2ID := a2.Props["id"].(string)
		g.Links = append(g.Links, graph.Link{
			Source: a1ID,
			Target: a2ID,
		})
	}
	return &g, nil
}
