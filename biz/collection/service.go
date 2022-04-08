package collection

import (
	"context"
	"github.com/ProjectOort/oort-server/api/middleware/auth"
	bizerr "github.com/ProjectOort/oort-server/biz/errors"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Service struct {
	logger *zap.Logger
	repo   Repo
}

type Repo interface {
	Create(ctx context.Context, col *Collection) error
	Update(ctx context.Context, col *Collection) (*Collection, error)
	Delete(ctx context.Context, colID primitive.ObjectID) error
	Get(ctx context.Context, colID primitive.ObjectID) (*Collection, error)
	List(ctx context.Context, OwnerID primitive.ObjectID) ([]*Collection, error)

	PushItem(ctx context.Context, colID primitive.ObjectID, itemID primitive.ObjectID) error
	PopItem(ctx context.Context, colID primitive.ObjectID, itemID primitive.ObjectID) error
	ListItems(ctx context.Context, colID primitive.ObjectID) ([]*Item, error)
}

func NewService(logger *zap.Logger, repo Repo) *Service {
	return &Service{
		logger: logger,
		repo:   repo,
	}
}

func (s *Service) Create(ctx context.Context, col *Collection) error {
	col.ID = primitive.NewObjectID()
	col.State = true
	col.OwnerID = auth.FromContext(ctx).ID
	col.CreatedTime = time.Now()
	col.UpdatedTime = time.Now()
	return errors.WithStack(s.repo.Create(ctx, col))
}

func (s *Service) Update(ctx context.Context, col *Collection) error {
	if err := s.checkIfCollectionBelongToUser(ctx, auth.FromContext(ctx).ID, col.ID); err != nil {
		return err
	}
	col.UpdatedTime = time.Now()
	_, err := s.repo.Update(ctx, col)
	return errors.WithStack(err)
}

func (s *Service) checkIfCollectionBelongToUser(ctx context.Context, accID, colID primitive.ObjectID) error {
	existedCol, err := s.repo.Get(ctx, colID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return bizerr.New().StatusCode(http.StatusNotFound).Msg("收藏夹不存在").WrapSelf()
		}
		return errors.WithStack(err)
	}
	if existedCol.OwnerID != accID {
		return bizerr.New().StatusCode(http.StatusForbidden).Msg("你无权访问不属于你的收藏夹").WrapSelf()
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, colID primitive.ObjectID) error {
	if err := s.checkIfCollectionBelongToUser(ctx, auth.FromContext(ctx).ID, colID); err != nil {
		return err
	}
	return errors.WithStack(s.repo.Delete(ctx, colID))
}

func (s *Service) List(ctx context.Context) ([]*Collection, error) {
	cols, err := s.repo.List(ctx, auth.FromContext(ctx).ID)
	return cols, errors.WithStack(err)
}

func (s *Service) PushItem(ctx context.Context, colID primitive.ObjectID, itemID primitive.ObjectID) error {
	if err := s.checkIfCollectionBelongToUser(ctx, auth.FromContext(ctx).ID, colID); err != nil {
		return err
	}
	return errors.WithStack(s.repo.PushItem(ctx, colID, itemID))
}

func (s *Service) PopItem(ctx context.Context, colID primitive.ObjectID, itemID primitive.ObjectID) error {
	if err := s.checkIfCollectionBelongToUser(ctx, auth.FromContext(ctx).ID, colID); err != nil {
		return err
	}
	return errors.WithStack(s.repo.PopItem(ctx, colID, itemID))
}

func (s *Service) ListItems(ctx context.Context, colID primitive.ObjectID) ([]*Item, error) {
	if err := s.checkIfCollectionBelongToUser(ctx, auth.FromContext(ctx).ID, colID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListItems(ctx, colID)
	return items, errors.WithStack(err)
}
