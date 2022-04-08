package search

import (
	"github.com/ProjectOort/oort-server/api/middleware/gerrors"
	"github.com/ProjectOort/oort-server/api/middleware/requestid"
	"github.com/ProjectOort/oort-server/biz/search"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func RegisterHandlers(r fiber.Router, logger *zap.Logger, searchService *search.Service) {
	h := handler{logger: logger, searchService: searchService}

	r.Get("search/asteroid", h.searchAsteroid)
}

type handler struct {
	logger        *zap.Logger
	searchService *search.Service
}

func (h *handler) searchAsteroid(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()

	var input struct {
		Text string `json:"q"`
	}
	if err := c.QueryParser(&input); err != nil {
		return errors.WithStack(gerrors.ErrParamsParsingFailed)
	}
	log.Debugw("parsed params", "query", input)

	items, err := h.searchService.Asteroid(c.Context(), input.Text)
	if err != nil {
		return err
	}
	toJ := make([]*Item, 0, len(items))
	for _, item := range items {
		toJ = append(toJ, MakeItemPresenter(item))
	}
	return c.JSON(toJ)
}
