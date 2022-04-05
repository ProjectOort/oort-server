package main

import (
	"github.com/ProjectOort/oort-server/api/handler/index"
	"github.com/ProjectOort/oort-server/conf"
	"github.com/gofiber/fiber/v2"
)

func main() {

	cfg := conf.Parse("conf/")

	app := fiber.New()

	r := app.Group("/api")
	index.MakeHandlers(r, cfg)

	app.Listen(cfg.Endpoint.HTTP.URL)
}
