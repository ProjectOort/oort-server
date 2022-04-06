package requestid

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

const _RequestIDKey = "_REQID_"

func New() fiber.Handler {
	m := requestid.New(requestid.Config{
		ContextKey: _RequestIDKey,
	})
	return m
}

func FromCtx(c *fiber.Ctx) string {
	return c.Locals(_RequestIDKey).(string)
}

func FromContxt(ctx context.Context) string {
	return ctx.Value(_RequestIDKey).(string)
}
