package routes

import (
	"go-fiber-api/pkg/httpcache"

	"github.com/gofiber/fiber/v2"
)

func useGetCache(r fiber.Router) {
	r.Use(httpcache.Middleware())
}
