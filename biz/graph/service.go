package graph

import (
	"context"
	"github.com/ProjectOort/oort-server/api/middleware/auth"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type Service struct {
	logger *zap.Logger
	repo   Repo
}

type Repo interface {
	GetGraphByAsteroidID(ctx context.Context, astID primitive.ObjectID, depth int) (*Graph, error)
	GetFullGraph(ctx context.Context, accID primitive.ObjectID) (*Graph, error)
}

func NewService(logger *zap.Logger, repo Repo) *Service {
	return &Service{
		logger: logger,
		repo:   repo,
	}
}

func (s *Service) GetByAsteroidID(ctx context.Context, astID primitive.ObjectID, depth int) (*Graph, error) {
	if depth <= 0 || depth > 20 {
		depth = 20
	}
	gph, err := s.repo.GetGraphByAsteroidID(ctx, astID, depth)
	return gph, errors.WithStack(err)
}

func (s *Service) GetFull(ctx context.Context) (*Graph, error) {
	gph, err := s.repo.GetFullGraph(ctx, auth.FromContext(ctx).ID)
	return gph, errors.WithStack(err)
}
