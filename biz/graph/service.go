package graph

import (
	"context"
	"github.com/pkg/errors"
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

func NewService(logger *zap.Logger, repo Repo) *Service {
	return &Service{
		logger: logger,
		repo:   repo,
	}
}

func (s *Service) GetByAsteroidID(ctx context.Context, astID primitive.ObjectID) (*Graph, error) {
	gph, err := s.repo.GetGraphByAsteroidID(ctx, astID)
	return gph, errors.WithStack(err)
}
