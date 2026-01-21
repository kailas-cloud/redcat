package api

import (
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Server struct {
	app    *fiber.App
	logger *slog.Logger
}

func New() *Server {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
	})

	// Request logging middleware
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		logger.Info("request",
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("latency", latency),
			slog.String("ip", c.IP()),
		)
		return err
	})

	return &Server{app: app, logger: logger}
}

func (s *Server) App() *fiber.App    { return s.app }
func (s *Server) Logger() *slog.Logger { return s.logger }
