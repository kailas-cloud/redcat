package api

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"redcat/internal/domain/model"
	svc "redcat/internal/service/places"
)

type Handlers struct {
	Places *svc.Service
}

func Register(app *fiber.App, h Handlers) {
	app.Get("/healthz", func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) })

	app.Post("/api/v1/places/search", func(c *fiber.Ctx) error {
		var req struct {
			Location    struct{ Lat, Lon float64 } `json:"location"`
			CategoryIDs []string                   `json:"category_ids"`
			Limit       int64                      `json:"limit"`
		}
		if err := c.BodyParser(&req); err != nil {
			slog.Warn("search: invalid JSON", slog.String("error", err.Error()))
			return fiber.NewError(http.StatusBadRequest, "invalid JSON")
		}

		slog.Info("search",
			slog.Float64("lat", req.Location.Lat),
			slog.Float64("lon", req.Location.Lon),
			slog.Int64("limit", req.Limit),
			slog.Int("categories", len(req.CategoryIDs)),
		)

		res, err := h.Places.SearchNearest(c.Context(), svc.SearchParams{
			Lat: req.Location.Lat, Lon: req.Location.Lon,
			Limit: req.Limit, CategoryIDs: req.CategoryIDs,
		})
		if err != nil {
			slog.Error("search failed", slog.String("error", err.Error()))
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}

		slog.Info("search completed", slog.Int("results", len(res)))

		items := make([]map[string]any, 0, len(res))
		for _, r := range res {
			p := r.Place
			item := map[string]any{
				"id":           p.ID,
				"name":         p.Name,
				"location":     map[string]any{"lat": p.Lat, "lon": p.Lon},
				"address":      p.Address,
				"locality":     p.Locality,
				"region":       p.Region,
				"postcode":     p.Postcode,
				"admin_region": p.AdminRegion,
				"post_town":    p.PostTown,
				"po_box":       p.PoBox,
				"country":      p.Country,
				"category_ids": p.CategoryIDs,
				"distance_m":   r.DistanceM,
			}
			items = append(items, item)
		}
		return c.JSON(fiber.Map{"places": items, "total": len(items), "query": fiber.Map{"location": req.Location, "limit": req.Limit}})
	})

	app.Post("/api/v1/places", func(c *fiber.Ctx) error {
		var p model.Place
		if err := c.BodyParser(&p); err != nil {
			slog.Warn("create place: invalid JSON", slog.String("error", err.Error()))
			return fiber.NewError(http.StatusBadRequest, "invalid JSON")
		}
		if strings.TrimSpace(p.ID) == "" || strings.TrimSpace(p.Name) == "" {
			slog.Warn("create place: missing required fields", slog.String("id", p.ID), slog.String("name", p.Name))
			return fiber.NewError(http.StatusBadRequest, "id and name required")
		}

		slog.Info("creating place", slog.String("id", p.ID), slog.String("name", p.Name))

		if err := h.Places.Add(c.Context(), p); err != nil {
			slog.Error("create place failed", slog.String("id", p.ID), slog.String("error", err.Error()))
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}

		slog.Info("place created", slog.String("id", p.ID))
		return c.Status(http.StatusCreated).JSON(fiber.Map{"item": p})
	})

	app.Get("/api/v1/places/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		slog.Info("getting place", slog.String("id", id))

		p, err := h.Places.Get(c.Context(), id)
		if err != nil {
			slog.Warn("place not found", slog.String("id", id))
			return fiber.NewError(http.StatusNotFound, "not found")
		}
		return c.JSON(p)
	})

	app.Delete("/api/v1/places/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		slog.Info("deleting place", slog.String("id", id))

		if err := h.Places.Delete(c.Context(), id); err != nil {
			slog.Error("delete place failed", slog.String("id", id), slog.String("error", err.Error()))
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}

		slog.Info("place deleted", slog.String("id", id))
		return c.SendStatus(http.StatusNoContent)
	})
}
