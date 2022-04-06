package graph

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type Service struct {
	logger *zap.Logger
	repo   Repo
}

type Repo interface {
	GetGraphByAsteroidID(ctx context.Context, astID primitive.ObjectID) (*Graph, error)
}

func (x *Service) GetByAsteroidID(ctx context.Context, astID primitive.ObjectID) (*Graph, error) {
	return x.repo.GetGraphByAsteroidID(ctx, astID)
}
