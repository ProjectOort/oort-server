package account

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/ProjectOort/oort-server/conf"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Service struct {
	logger      *zap.Logger
	repo        Repo
	tokenHelper *tokenHelper
	giteeHelper *giteeHelper
}

type Repo interface {
	Create(ctx context.Context, account *Account) error
	GetByGiteeID(ctx context.Context, id int) (*Account, error)
	GetByUserName(ctx context.Context, uname string) (*Account, error)
	GetByEmail(ctx context.Context, email string) (*Account, error)
	GetByMobile(ctx context.Context, mobile string) (*Account, error)
}

func NewService(logger *zap.Logger, cfg *conf.Account, repo Repo) *Service {
	return &Service{
		logger:      logger,
		repo:        repo,
		tokenHelper: newTokenHelper(cfg),
		giteeHelper: newGiteeHelper(cfg),
	}
}

var (
	UserNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{4,20}$`)
	EmailPattern    = regexp.MustCompile(`^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`)
	MobilePattern   = regexp.MustCompile(`^(13[0-9]|14[5|7]|15[0|1|2|3|4|5|6|7|8|9]|18[0|1|2|3|5|6|7|8|9])\d{8}$`)
)

func (s *Service) Token(_ context.Context, acc *Account) (string, error) {
	return acc.Token(s.tokenHelper)
}

func (s *Service) ValidateToken(_ context.Context, token string) (primitive.ObjectID, error) {
	return ValidateToken(token, s.tokenHelper)
}

func (s *Service) Login(ctx context.Context, identifier, password string) (*Account, error) {
	if identifier == "" || password == "" {
		return nil, errors.New("wrong identifier or password")
	}
	switch {
	case UserNamePattern.MatchString(identifier):
		return s.loginByUserName(ctx, identifier, password)
	case EmailPattern.MatchString(identifier):
		return s.loginByEmail(ctx, identifier, password)
	case MobilePattern.MatchString(identifier):
		return s.loginByMobile(ctx, identifier, password)
	default:
		return nil, errors.New("wrong identifier or password")
	}
}

func (s *Service) loginByUserName(ctx context.Context, uname, passwd string) (*Account, error) {
	s.logger.Debug("loginByUserName received: ", zap.String("uname", uname), zap.String("passwd", passwd))
	account, err := s.repo.GetByUserName(ctx, uname)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			s.logger.Debug("uname doesn't exist", zap.String("uname", uname))
			return nil, errors.New("user doesn't exist")
		}

		return nil, errors.New("unknown")
	}
	if !account.PasswdEqual(passwd) {
		return nil, errors.New("wrong password")
	}
	return account, nil
}

func (s *Service) loginByEmail(ctx context.Context, email, passwd string) (*Account, error) {
	account, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user doesn't exist")
		}
		return nil, errors.New("unknown")
	}
	if !account.PasswdEqual(passwd) {
		return nil, errors.New("wrong password")
	}
	return account, nil
}

func (s *Service) loginByMobile(ctx context.Context, mobile, passwd string) (*Account, error) {
	account, err := s.repo.GetByMobile(ctx, mobile)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user doesn't exist")
		}
		return nil, errors.New("unknown")
	}
	if !account.PasswdEqual(passwd) {
		return nil, errors.New("wrong password")
	}
	return account, nil
}

func (s *Service) Register(ctx context.Context, acc *Account) error {
	if _, err := s.repo.GetByUserName(ctx, acc.UserName); !errors.Is(err, mongo.ErrNoDocuments) {
		return errors.New("user_name already exist")
	}
	if err := acc.HashPasswd(); err != nil {
		return errors.New("invaild password")
	}
	acc.ID = primitive.NewObjectID()
	acc.CreatedTime = time.Now()
	acc.UpdatedTime = time.Now()
	acc.State = true
	return s.repo.Create(ctx, acc)
}

func (s *Service) OAuthGitee(ctx context.Context, code string) (*Account, error) {
	if code == "" {
		return nil, errors.New("invalid code")
	}

	// uses code to exchange access_token from gitee server.
	oauthResult, err := s.giteeHelper.OAuth(code)
	if err != nil {
		return nil, errors.New("unknown")
	}

	// uses access_token to fetch user info.
	userInfoResult, err := s.giteeHelper.UserInfo(oauthResult.AccessToken)
	if err != nil {
		return nil, errors.New("unknown")
	}

	account, err := s.repo.GetByGiteeID(ctx, userInfoResult.ID)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, errors.New("unknown")
	}

	// if the account doesn't exist. creates a new one.
	if err == mongo.ErrNoDocuments {
		account = &Account{
			ID:    primitive.NewObjectID(),
			State: true,
			BindStatus: BindStatus{
				Gitee: true,
			},
			AvatarURL:   userInfoResult.AvatarURL,
			NickName:    userInfoResult.Name,
			GiteeID:     userInfoResult.ID,
			CreatedTime: time.Now(),
			UpdatedTime: time.Now(),
		}
		err := s.repo.Create(ctx, account)
		if err != nil {
			return nil, errors.New("unknown")
		}
	}

	return account, nil
}
