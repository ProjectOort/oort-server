package index

import (
	"github.com/ProjectOort/oort-server/conf"
	"github.com/gofiber/fiber/v2"
)

func RegisterHandlers(r fiber.Router, cfg *conf.App) {
	var h = handler{cfg: cfg}
	r.Get("/", h.Index)
}

type handler struct {
	cfg *conf.App
}

func (x handler) Index(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"name":    x.cfg.Name,
		"version": x.cfg.Version,
		"status":  "UP",
	})
}
