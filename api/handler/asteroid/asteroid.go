package asteroid

import (
	"github.com/ProjectOort/oort-server/api/middleware/auth"
	"github.com/ProjectOort/oort-server/api/middleware/gerrors"
	"github.com/ProjectOort/oort-server/api/middleware/requestid"
	"github.com/ProjectOort/oort-server/biz/asteroid"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

func RegisterHandlers(r fiber.Router, logger *zap.Logger, validate *validator.Validate, asteroidService *asteroid.Service) {
	h := &handler{logger, validate, asteroidService}

	r.Post("/asteroid", h.create)
	r.Post("/asteroid!linkTo", h.linkTo)
	r.Post("/asteroid!linkFrom", h.linkFrom)
	r.Put("/asteroid/content", h.sync)
	r.Get("/asteroids", h.list)
	r.Get("/asteroid", h.get)
	r.Get("/linked/from/asteroid", h.listLinkedFrom)
	r.Get("/linked/to/asteroid", h.listLinkedTo)
}

type handler struct {
	logger          *zap.Logger
	validate        *validator.Validate
	asteroidService *asteroid.Service
}

func (h *handler) create(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()

	var input struct {
		Hub      *bool    `json:"hub" validate:"required"`
		Title    string   `json:"title" validate:"required"`
		Content  string   `json:"content"`
		LinkFrom []string `json:"link_from"`
		LinkTo   []string `json:"link_to"`
	}
	if err := c.BodyParser(&input); err != nil {
		return errors.WithStack(gerrors.ErrParamsParsingFailed)
	}
	log.Debugw("parsed params", "body", input)
	if err := h.validate.Struct(input); err != nil {
		return err
	}

	linkFromIDs := make([]primitive.ObjectID, 0, len(input.LinkFrom))
	linkToIDs := make([]primitive.ObjectID, 0, len(input.LinkTo))

	// convert hex strings to ObjectID for from links.
	for _, id := range input.LinkFrom {
		id, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}
		linkFromIDs = append(linkFromIDs, id)
	}

	// convert hex strings to ObjectID for to links.
	for _, id := range input.LinkTo {
		id, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}
		linkToIDs = append(linkToIDs, id)
	}

	accID := auth.FromCtx(c).ID
	ast, err := h.asteroidService.Create(c.Context(), &asteroid.Asteroid{
		Hub:      *input.Hub,
		AuthorID: accID,
		Title:    input.Title,
		Content:  input.Content,
	}, linkFromIDs, linkToIDs)
	if err != nil {
		return err
	}

	toJ := MakeAsteroidPresenter(ast)
	return c.JSON(toJ)
}

func (h *handler) linkTo(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()

	var input struct {
		ID     string   `json:"id" validate:"required"`
		LinkTo []string `json:"link_to" validate:"required"`
	}
	if err := c.BodyParser(&input); err != nil {
		return errors.WithStack(gerrors.ErrParamsParsingFailed)
	}
	log.Debugw("parsed params", "body", input)
	if err := h.validate.Struct(input); err != nil {
		return err
	}

	curAstID, err := primitive.ObjectIDFromHex(input.ID)
	if err != nil {
		return err
	}
	linkToIDs := make([]primitive.ObjectID, 0, len(input.LinkTo))
	for _, id := range input.LinkTo {
		id, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}
		linkToIDs = append(linkToIDs, id)
	}

	return h.asteroidService.LinkTo(c.Context(), curAstID, linkToIDs)
}

func (h *handler) linkFrom(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()

	var input struct {
		ID       string   `json:"id" validate:"required"`
		LinkFrom []string `json:"link_from" validate:"required"`
	}
	if err := c.BodyParser(&input); err != nil {
		return errors.WithStack(gerrors.ErrParamsParsingFailed)
	}
	log.Debugw("parsed params", "body", input)
	if err := h.validate.Struct(input); err != nil {
		return err
	}

	curAstID, err := primitive.ObjectIDFromHex(input.ID)
	if err != nil {
		return err
	}
	linkFromIDs := make([]primitive.ObjectID, 0, len(input.LinkFrom))
	for _, id := range input.LinkFrom {
		id, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}
		linkFromIDs = append(linkFromIDs, id)
	}
	return h.asteroidService.LinkFrom(c.Context(), curAstID, linkFromIDs)
}

func (h *handler) sync(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()

	var input struct {
		ID      string `json:"id" validate:"required"`
		Content string `json:"content"`
	}
	if err := c.BodyParser(&input); err != nil {
		return errors.WithStack(gerrors.ErrParamsParsingFailed)
	}
	log.Debugw("parsed params", "body", input)
	if err := h.validate.Struct(input); err != nil {
		return err
	}

	astID, err := primitive.ObjectIDFromHex(input.ID)
	if err != nil {
		return err
	}

	return h.asteroidService.Sync(c.Context(), &asteroid.Asteroid{ID: astID, Content: input.Content})
}

func (h *handler) list(c *fiber.Ctx) error {
	asts, err := h.asteroidService.List(c.Context())
	if err != nil {
		return errors.WithStack(gerrors.ErrParamsParsingFailed)
	}
	toJ := make([]*Item, 0, len(asts))
	for _, ast := range asts {
		toJ = append(toJ, MakeItemPresenter(ast))
	}
	return c.JSON(toJ)
}

func (h *handler) get(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()

	var input struct {
		ID string `json:"id" validate:"required"`
	}
	if err := c.QueryParser(&input); err != nil {
		return errors.WithStack(gerrors.ErrParamsParsingFailed)
	}
	log.Debugw("parsed params", "query", input)
	if err := h.validate.Struct(input); err != nil {
		return err
	}

	astID, err := primitive.ObjectIDFromHex(input.ID)
	if err != nil {
		return err
	}
	ast, err := h.asteroidService.Get(c.Context(), astID)
	if err != nil {
		return err
	}
	toJ := MakeAsteroidPresenter(ast)
	return c.JSON(toJ)
}

func (h *handler) listLinkedFrom(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()

	var input struct {
		ID string `json:"id" validate:"required"`
	}
	if err := c.QueryParser(&input); err != nil {
		return errors.WithStack(gerrors.ErrParamsParsingFailed)
	}
	log.Debugw("parsed params", "query", input)
	if err := h.validate.Struct(input); err != nil {
		return err
	}

	astID, err := primitive.ObjectIDFromHex(input.ID)
	if err != nil {
		return err
	}
	asts, err := h.asteroidService.ListLinkedFrom(c.Context(), astID)
	if err != nil {
		return err
	}
	toJ := make([]*Item, 0, len(asts))
	for _, ast := range asts {
		toJ = append(toJ, MakeItemPresenter(ast))
	}
	return c.JSON(toJ)
}

func (h *handler) listLinkedTo(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()

	var input struct {
		ID string `json:"id" validate:"required"`
	}
	if err := c.QueryParser(&input); err != nil {
		return errors.WithStack(gerrors.ErrParamsParsingFailed)
	}
	log.Debugw("parsed params", "query", input)
	if err := h.validate.Struct(input); err != nil {
		return err
	}

	astID, err := primitive.ObjectIDFromHex(input.ID)
	if err != nil {
		return err
	}
	asts, err := h.asteroidService.ListLinkedTo(c.Context(), astID)
	if err != nil {
		return err
	}
	toJ := make([]*Item, 0, len(asts))
	for _, ast := range asts {
		toJ = append(toJ, MakeItemPresenter(ast))
	}
	return c.JSON(toJ)
}
