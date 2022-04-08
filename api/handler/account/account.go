package account

import (
	"github.com/ProjectOort/oort-server/api/middleware/gerrors"
	"github.com/ProjectOort/oort-server/api/middleware/requestid"
	"github.com/ProjectOort/oort-server/biz/account"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func RegisterHandlers(r fiber.Router, logger *zap.Logger, validate *validator.Validate, accountService *account.Service) {
	h := &handler{logger, validate, accountService}

	r.Post("/account!login", h.login)
	r.Post("/account!register", h.register)
	r.Post("/account!oauth/gitee", h.oAuthGitee)
	r.Put("/account/password", h.updatePassword)
}

type handler struct {
	logger         *zap.Logger
	validate       *validator.Validate
	accountService *account.Service
}

func (h *handler) login(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()

	var input struct {
		Identifier string `json:"identifier" validate:"required"`
		Password   string `json:"password" validate:"required"`
	}
	if err := c.BodyParser(&input); err != nil {
		return errors.WithStack(gerrors.ErrParamsParsingFailed)
	}
	log.Debugw("parsed params", "body", input)
	if err := h.validate.Struct(input); err != nil {
		return err
	}

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

func (h *handler) register(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()

	var input struct {
		AvatarURL   string `json:"avatar_url"`
		UserName    string `json:"user_name" validate:"required"`
		Password    string `json:"password" validate:"required"`
		NickName    string `json:"nick_name" validate:"required"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&input); err != nil {
		return errors.WithStack(gerrors.ErrParamsParsingFailed)
	}
	log.Debugw("parsed params", "body", input)
	if err := h.validate.Struct(input); err != nil {
		return err
	}

	err := h.accountService.Register(c.Context(), &account.Account{
		NickName:    input.NickName,
		AvatarURL:   input.AvatarURL,
		Description: input.Description,
		UserName:    input.UserName,
		Password:    input.Password,
	})
	return err
}

func (h *handler) oAuthGitee(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()

	var input struct {
		Code string `json:"code" validate:"required"`
	}
	if err := c.BodyParser(&input); err != nil {
		return errors.WithStack(gerrors.ErrParamsParsingFailed)
	}
	log.Debugw("parsed params", "body", input)
	if err := h.validate.Struct(input); err != nil {
		return err
	}

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

func (h *handler) updatePassword(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()

	var input struct {
		OldPassword string `json:"old_password" validate:"required"`
		NewPassword string `json:"new_password" validate:"required"`
	}
	if err := c.BodyParser(&input); err != nil {
		return errors.WithStack(gerrors.ErrParamsParsingFailed)
	}
	log.Debugw("parsed params", "body", input)
	if err := h.validate.Struct(input); err != nil {
		return err
	}

	return h.accountService.UpdatePassword(c.Context(), input.NewPassword, input.OldPassword)
}
