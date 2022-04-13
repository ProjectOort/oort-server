package graph

import (
	"github.com/ProjectOort/oort-server/api/middleware/gerrors"
	"github.com/ProjectOort/oort-server/api/middleware/requestid"
	"github.com/ProjectOort/oort-server/biz/graph"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

func RegisterHandlers(r fiber.Router, logger *zap.Logger, validate *validator.Validate, graphService *graph.Service) {
	h := handler{logger, validate, graphService}

	r.Get("/graph/asteroid", h.getByAsteroidID)
	r.Get("/graph/full", h.getFull)
}

type handler struct {
	logger       *zap.Logger
	validate     *validator.Validate
	graphService *graph.Service
}

func (h *handler) getByAsteroidID(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()

	var input struct {
		ID    string `json:"id"`
		Depth int    `json:"depth"`
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
	gph, err := h.graphService.GetByAsteroidID(c.Context(), astID, input.Depth)
	if err != nil {
		return err
	}
	toJ := MakeGraphPresenter(gph)
	return c.JSON(toJ)
}

func (h *handler) getFull(c *fiber.Ctx) error {
	_ = h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()
	gph, err := h.graphService.GetFull(c.Context())
	if err != nil {
		return err
	}
	toJ := MakeGraphPresenter(gph)
	return c.JSON(toJ)
}
