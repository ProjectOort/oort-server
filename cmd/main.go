package main

import (
	"github.com/ProjectOort/oort-server/api/handler/index"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	r := app.Group("/api")
	index.MakeHandlers(r)

	app.Listen(":8080")
}
