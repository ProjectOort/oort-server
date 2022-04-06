package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/ProjectOort/oort-server/api/middleware/requestid"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

const (
	_AccountIDKey      = "_ACCID_"
	_BearerTokenPrefix = "Bearer "
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type tokenValidator interface {
	ValidateToken(_ context.Context, token string) (primitive.ObjectID, error)
}

type Info struct {
	ID primitive.ObjectID
}

func New(logger *zap.Logger, tokenValidator tokenValidator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := logger.With(zap.String("request_id", requestid.FromCtx(c)))

		// get authorization information
		auth := c.Get(fiber.HeaderAuthorization, "")
		if auth == "" || !strings.HasPrefix(auth, _BearerTokenPrefix) {
			log.Info("[M-Auth] Auth failed, invaild Token")
			return ErrInvalidToken
		}

		// remove prefix
		token := auth[len(_BearerTokenPrefix):]

		accID, err := tokenValidator.ValidateToken(c.Context(), token)
		if err != nil {
			log.Info("[M-Auth] Auth failed, validation failed", zap.Error(err))
			return ErrInvalidToken
		}
		log.Debug("[M-Auth] Auth success, authorized account", zap.String("account_id", accID.Hex()))

		c.Locals(_AccountIDKey, Info{accID})
		return c.Next()
	}

}

func FromCtx(c *fiber.Ctx) Info {
	return c.Locals(_AccountIDKey).(Info)
}

func FromContext(ctx context.Context) Info {
	return ctx.Value(_AccountIDKey).(Info)
}
