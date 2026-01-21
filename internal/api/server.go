package api

import (
	"github.com/gofiber/fiber/v2"
)

type Server struct {
	app *fiber.App
}

func New() *Server {
	app := fiber.New()
	return &Server{app: app}
}

func (s *Server) App() *fiber.App { return s.app }
