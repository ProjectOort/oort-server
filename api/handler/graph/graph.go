package graph

import (
	"github.com/ProjectOort/oort-server/api/middleware/requestid"
	"github.com/ProjectOort/oort-server/biz/graph"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

func RegisterHandlers(r fiber.Router, logger *zap.Logger, graphService *graph.Service) {
	h := handler{logger: logger, graphService: graphService}

	r.Get("/graph", h.getByAsteroidID)
}

type handler struct {
	logger       *zap.Logger
	graphService *graph.Service
}

func (h *handler) getByAsteroidID(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()

	var input struct {
		ID string `json:"id"`
	}
	if err := c.QueryParser(&input); err != nil {
		return err
	}
	log.Debugf("parsed params, input = %+v", input)

	astID, err := primitive.ObjectIDFromHex(input.ID)
	if err != nil {
		return err
	}
	gph, err := h.graphService.GetByAsteroidID(c.Context(), astID)
	if err != nil {
		return err
	}
	toJ := MakeGraphPresenter(gph)
	return c.JSON(toJ)
}
