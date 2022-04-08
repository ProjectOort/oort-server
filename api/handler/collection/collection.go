package collection

import (
	"github.com/ProjectOort/oort-server/api/middleware/requestid"
	"github.com/ProjectOort/oort-server/biz/collection"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

func RegisterHandlers(r fiber.Router, logger *zap.Logger, collectionService *collection.Service) {
	h := handler{logger: logger, collectionService: collectionService}

	r.Post("/collection", h.create)
	r.Put("/collection", h.update)
	r.Delete("/collection", h.delete)
	r.Get("/collections", h.update)
	r.Post("/collection/item", h.pushItem)
	r.Delete("/collection/item", h.popItem)
	r.Get("/collection/items", h.listItems)
}

type handler struct {
	logger            *zap.Logger
	collectionService *collection.Service
}

func (h *handler) create(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&input); err != nil {
		return err
	}
	log.Debugf("parsed params, input = %+v", input)

	return h.collectionService.Create(c.Context(), &collection.Collection{
		Name:        input.Name,
		Description: input.Description,
	})
}

func (h *handler) update(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()
	var input struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&input); err != nil {
		return err
	}
	log.Debugf("parsed params, input = %+v", input)

	colID, err := primitive.ObjectIDFromHex(input.ID)
	if err != nil {
		return err
	}

	return h.collectionService.Update(c.Context(), &collection.Collection{
		ID:          colID,
		Name:        input.Name,
		Description: input.Description,
	})
}

func (h *handler) delete(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()
	var input struct {
		ID string `json:"id"`
	}
	if err := c.BodyParser(&input); err != nil {
		return err
	}
	log.Debugf("parsed params, input = %+v", input)

	colID, err := primitive.ObjectIDFromHex(input.ID)
	if err != nil {
		return err
	}
	return h.collectionService.Delete(c.Context(), colID)
}

func (h *handler) list(c *fiber.Ctx) error {
	_ = h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()
	cols, err := h.collectionService.List(c.Context())
	if err != nil {
		return err
	}
	toJ := make([]*Collection, 0, len(cols))
	for _, col := range cols {
		toJ = append(toJ, MakeCollectionPresenter(col))
	}
	return c.JSON(toJ)
}

func (h *handler) pushItem(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()
	var input struct {
		CollectionID string `json:"collection_id"`
		ItemID       string `json:"item_id"`
	}
	if err := c.BodyParser(&input); err != nil {
		return err
	}
	log.Debugf("parsed params, input = %+v", input)

	colID, err := primitive.ObjectIDFromHex(input.CollectionID)
	if err != nil {
		return err
	}

	itemID, err := primitive.ObjectIDFromHex(input.ItemID)
	if err != nil {
		return err
	}

	return h.collectionService.PushItem(c.Context(), colID, itemID)
}

func (h *handler) popItem(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()
	var input struct {
		CollectionID string `json:"collection_id"`
		ItemID       string `json:"item_id"`
	}
	if err := c.BodyParser(&input); err != nil {
		return err
	}
	log.Debugf("parsed params, input = %+v", input)

	colID, err := primitive.ObjectIDFromHex(input.CollectionID)
	if err != nil {
		return err
	}

	itemID, err := primitive.ObjectIDFromHex(input.ItemID)
	if err != nil {
		return err
	}
	return h.collectionService.PopItem(c.Context(), colID, itemID)
}

func (h *handler) listItems(c *fiber.Ctx) error {
	log := h.logger.Named("[HANDLER]").With(zap.String("request_id", requestid.FromCtx(c))).Sugar()
	var input struct {
		ID string `json:"id"`
	}
	if err := c.QueryParser(&input); err != nil {
		return err
	}
	log.Debugf("parsed params, input = %+v", input)

	colID, err := primitive.ObjectIDFromHex(input.ID)
	if err != nil {
		return err
	}
	items, err := h.collectionService.ListItems(c.Context(), colID)
	if err != nil {
		return err
	}
	toJ := make([]*Item, 0, len(items))
	for _, item := range items {
		toJ = append(toJ, MakeItemPresenter(item))
	}
	return c.JSON(toJ)
}
