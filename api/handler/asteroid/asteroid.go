package asteroid

import (
	"github.com/ProjectOort/oort-server/api/middleware/auth"
	"github.com/ProjectOort/oort-server/api/middleware/requestid"
	"github.com/ProjectOort/oort-server/biz/asteroid"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

func MakeHandlers(r fiber.Router, logger *zap.Logger, asteroidService *asteroid.Service) {
	h := &handler{logger: logger, asteroidService: asteroidService}

	r.Post("/asteroid", h.create)
}

type handler struct {
	logger          *zap.Logger
	asteroidService *asteroid.Service
}

func (h *handler) create(c *fiber.Ctx) error {
	log := h.logger.With(zap.String("request_id", requestid.FromCtx(c))).Sugar()

	var input struct {
		Hub      bool     `json:"hub"`
		Title    string   `json:"title"`
		Content  string   `json:"content"`
		LinkFrom []string `json:"link_from"`
		LinkTo   []string `json:"link_to"`
	}
	if err := c.BodyParser(&input); err != nil {
		return err
	}
	log.Debugf("[H] parsed params, input = %+v", input)

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
		Hub:      false,
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
