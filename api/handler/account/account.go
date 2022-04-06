package account

import (
	"github.com/ProjectOort/oort-server/api/middleware/requestid"
	"github.com/ProjectOort/oort-server/biz/account"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func MakeHandlers(r fiber.Router, logger *zap.Logger, accountService *account.Service) {
	h := &handler{logger: logger, accountService: accountService}

	r.Post("/account/login", h.login)
	r.Post("/account/oauth/gitee", h.oAuthGitee)
}

type handler struct {
	logger         *zap.Logger
	accountService *account.Service
}

func (h *handler) login(c *fiber.Ctx) error {
	log := h.logger.With(zap.String("request_id", requestid.FromCtx(c))).Sugar()

	var input struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}
	if err := c.BodyParser(&input); err != nil {
		return err
	}
	log.Debugf("[H-login] parsed params, input = %+v", input)

	acc, err := h.accountService.Login(c.Context(), input.Identifier, input.Password)
	if err != nil {
		return err
	}

	token, err := h.accountService.Token(c.Context(), acc)
	if err != nil {
		return err
	}

	toJ := MakeAccountPresenter(acc, token)
	return c.JSON(toJ)
}

func (h *handler) oAuthGitee(c *fiber.Ctx) error {
	log := h.logger.With(zap.String("request_id", requestid.FromCtx(c))).Sugar()

	var input struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&input); err != nil {
		return err
	}
	log.Debugf("[H-oAuthGitee] parsed params, input = %+v", input)

	acc, err := h.accountService.OAuthGitee(c.Context(), input.Code)
	if err != nil {
		return err
	}

	token, err := h.accountService.Token(c.Context(), acc)
	if err != nil {
		return err
	}

	toJ := MakeAccountPresenter(acc, token)
	return c.JSON(toJ)
}
