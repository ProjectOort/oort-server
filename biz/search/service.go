package search

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
	SearchAsteroid(ctx context.Context, text string, authorID primitive.ObjectID) ([]*Item, error)
}

func NewService(logger *zap.Logger, repo Repo) *Service {
	return &Service{
		logger: logger,
		repo:   repo,
	}
}

func (s *Service) Asteroid(ctx context.Context, text string) ([]*Item, error) {
	items, err := s.repo.SearchAsteroid(ctx, text, auth.FromContext(ctx).ID)
	return items, errors.WithStack(err)
}
