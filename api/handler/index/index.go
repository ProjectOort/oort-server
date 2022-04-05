package index

import "github.com/gofiber/fiber/v2"

func MakeHandlers(r fiber.Router) {
	var h handler
	r.Get("/", h.Index)
}

type handler struct{}

func (x handler) Index(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"msg":     "Oort Server is running",
		"status":  "UP",
		"version": "1.0",
	})
}
