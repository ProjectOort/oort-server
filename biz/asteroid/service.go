package asteroid

import (
	"context"
	bizerr "github.com/ProjectOort/oort-server/biz/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"

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
	Create(context.Context, *Asteroid, []primitive.ObjectID, []primitive.ObjectID) error
	UpdateContent(context.Context, *Asteroid) error
	Get(context.Context, primitive.ObjectID) (*Asteroid, error)
	List(context.Context, []primitive.ObjectID) ([]*Asteroid, error)
	ListHub(context.Context, primitive.ObjectID) ([]*Asteroid, error)
	ListLinkedFrom(context.Context, primitive.ObjectID) ([]*Asteroid, error)
	ListLinkedTo(context.Context, primitive.ObjectID) ([]*Asteroid, error)
}

func NewService(logger *zap.Logger, repo Repo) *Service {
	return &Service{
		logger: logger,
		repo:   repo,
	}
}

func (s *Service) Create(ctx context.Context, ast *Asteroid, linkFromIDs []primitive.ObjectID, linkToIDs []primitive.ObjectID) (*Asteroid, error) {
	accID := auth.FromContext(ctx).ID

	ast.ID = primitive.NewObjectID()
	ast.State = true
	ast.CreatedTime = time.Now()
	ast.UpdatedTime = time.Now()

	if err := s.checkIfTargetAsteroidBelongToUser(ctx, accID, mergeIDSlices(linkFromIDs, linkToIDs)...); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, ast, linkFromIDs, linkToIDs); err != nil {
		return nil, errors.WithStack(err)
	}
	return ast, nil
}

func mergeIDSlices(s1 []primitive.ObjectID, s2 []primitive.ObjectID) []primitive.ObjectID {
	size := len(s1) + len(s2)
	res := make([]primitive.ObjectID, 0, size)
	res = append(res, s1...)
	res = append(res, s2...)
	return res
}

func (s *Service) checkIfTargetAsteroidBelongToUser(ctx context.Context, accID primitive.ObjectID, astIDs ...primitive.ObjectID) error {
	if len(astIDs) == 0 {
		return nil
	}
	asts, err := s.repo.List(ctx, astIDs)
	if err != nil {
		return errors.WithStack(err)
	}
	if len(asts) != len(astIDs) {
		return bizerr.New().StatusCode(http.StatusForbidden).Msg("你要连接的某些节点不存在").WrapSelf()
	}
	for _, ast := range asts {
		if ast.AuthorID != accID {
			return bizerr.New().StatusCode(http.StatusForbidden).Msg("你没有权限连接不属于你的节点").WrapSelf()
		}
	}
	return nil
}

func (s *Service) Sync(ctx context.Context, ast *Asteroid) error {
	existedAsteroid, err := s.repo.Get(ctx, ast.ID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return bizerr.New().StatusCode(http.StatusNotFound).Msg("你要同步的节点不存在").WrapSelf()
		}
		return errors.WithStack(err)
	}
	accID := auth.FromContext(ctx).ID
	if existedAsteroid.AuthorID != accID {
		return bizerr.New().StatusCode(http.StatusForbidden).Msg("你无权同步不属于你的节点").WrapSelf()
	}
	return s.repo.UpdateContent(ctx, ast)
}

func (s *Service) List(ctx context.Context) ([]*Asteroid, error) {
	// TODO Add other type of asteroid query support
	return s.repo.ListHub(ctx, auth.FromContext(ctx).ID)
}

func (s *Service) Get(ctx context.Context, astID primitive.ObjectID) (*Asteroid, error) {
	ast, err := s.repo.Get(ctx, astID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, bizerr.New().StatusCode(http.StatusNotFound).Msg("你要查看的节点不存在").WrapSelf()
		}
		return nil, errors.WithStack(err)
	}
	if ast.AuthorID != auth.FromContext(ctx).ID {
		return nil, bizerr.New().StatusCode(http.StatusForbidden).Msg("你无权查看不属于你的节点").WrapSelf()
	}
	return ast, nil
}

func (s *Service) ListLinkedFrom(ctx context.Context, astID primitive.ObjectID) ([]*Asteroid, error) {
	ast, err := s.repo.Get(ctx, astID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, bizerr.New().StatusCode(http.StatusNotFound).Msg("你要查看的节点不存在").WrapSelf()
		}
		return nil, err
	}
	if ast.AuthorID != auth.FromContext(ctx).ID {
		return nil, bizerr.New().StatusCode(http.StatusForbidden).Msg("你无权查看不属于你的节点").WrapSelf()
	}
	asts, err := s.repo.ListLinkedFrom(ctx, astID)
	return asts, errors.WithStack(err)
}

func (s *Service) ListLinkedTo(ctx context.Context, astID primitive.ObjectID) ([]*Asteroid, error) {
	ast, err := s.repo.Get(ctx, astID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, bizerr.New().StatusCode(http.StatusNotFound).Msg("你要查看的节点不存在").WrapSelf()
		}
		return nil, err
	}
	if ast.AuthorID != auth.FromContext(ctx).ID {
		return nil, bizerr.New().StatusCode(http.StatusForbidden).Msg("你无权查看不属于你的节点").WrapSelf()
	}
	asts, err := s.repo.ListLinkedTo(ctx, astID)
	return asts, errors.WithStack(err)
}
