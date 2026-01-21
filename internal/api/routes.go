package api

import (
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
			Location struct{ Lat, Lon float64 } `json:"location"`
			CategoryIDs []string `json:"category_ids"`
			Limit int64 `json:"limit"`
		}
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid JSON")
		}
		res, err := h.Places.SearchNearest(c.Context(), svc.SearchParams{
			Lat: req.Location.Lat, Lon: req.Location.Lon,
			Limit: req.Limit, CategoryIDs: req.CategoryIDs,
		})
		if err != nil { return fiber.NewError(http.StatusInternalServerError, err.Error()) }
		// flatten to PlaceWithDistance to match OpenAPI
		items := make([]map[string]any, 0, len(res))
		for _, r := range res {
			p := r.Place
			item := map[string]any{
				"id":   p.ID,
				"name": p.Name,
				"location": map[string]any{"lat": p.Lat, "lon": p.Lon},
				"address": p.Address,
				"locality": p.Locality,
				"region": p.Region,
				"postcode": p.Postcode,
				"admin_region": p.AdminRegion,
				"post_town": p.PostTown,
				"po_box": p.PoBox,
				"country": p.Country,
				"category_ids": p.CategoryIDs,
				"distance_m": r.DistanceM,
			}
			items = append(items, item)
		}
		return c.JSON(fiber.Map{"places": items, "total": len(items), "query": fiber.Map{"location": req.Location, "limit": req.Limit}})
	})

	app.Post("/api/v1/places", func(c *fiber.Ctx) error {
		var p model.Place
		if err := c.BodyParser(&p); err != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid JSON")
		}
		if strings.TrimSpace(p.ID) == "" || strings.TrimSpace(p.Name) == "" {
			return fiber.NewError(http.StatusBadRequest, "id and name required")
		}
		if err := h.Places.Add(c.Context(), p); err != nil {
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}
		return c.Status(http.StatusCreated).JSON(fiber.Map{"item": p})
	})

	app.Get("/api/v1/places/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		p, err := h.Places.Get(c.Context(), id)
		if err != nil { return fiber.NewError(http.StatusNotFound, "not found") }
		return c.JSON(p)
	})

	app.Delete("/api/v1/places/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if err := h.Places.Delete(c.Context(), id); err != nil {
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}
		return c.SendStatus(http.StatusNoContent)
	})
}
