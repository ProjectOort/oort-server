package gerrors

import (
	"errors"
	"fmt"
	"github.com/ProjectOort/oort-server/api/middleware/requestid"
	bizerr "github.com/ProjectOort/oort-server/biz/errors"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"net/http"
)

var (
	ErrParamsParsingFailed = errors.New("params parsing failed")
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func New(logger *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		log := logger.With(zap.String("request_id", requestid.FromCtx(c))).Named("[MIDDLEWARE]")
		if err == nil {
			return nil
		}
		if errors.Is(err, ErrParamsParsingFailed) {
			log.Sugar().Debugf("params parsing failed, error:\n%+v", err)
			return c.JSON(&ErrorResponse{Message: "参数解析失败"})
		}
		if berr, ok := bizerr.As(err); ok {
			if berr.GetStatusCode() == http.StatusInternalServerError {
				log.Error(fmt.Sprintf("unknown error:\n%+v\n", err), zap.Error(err))
			} else {
				log.Debug(fmt.Sprintf("error:\n%+v\n", err), zap.Error(err))
			}
			return c.Status(berr.GetStatusCode()).JSON(&ErrorResponse{Message: berr.GetMsg()})
		}
		log.Error(fmt.Sprintf("unknown error:\n%+v\n", err), zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(&ErrorResponse{
			Message: "服务器内部错误",
		})
	}
}
