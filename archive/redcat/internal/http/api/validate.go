package api

import (
	"errors"
	"redcat/internal/model"
	"strings"
)

func validatePlace(p *model.Place) error {
	p.FsqPlaceID = strings.TrimSpace(p.FsqPlaceID)
	p.Latitude = strings.TrimSpace(p.Latitude)
	p.Longitude = strings.TrimSpace(p.Longitude)
	p.Name = strings.TrimSpace(p.Name)
	p.Country = strings.TrimSpace(p.Country)

	if p.FsqPlaceID == "" {
		return errors.New("fsq_place_id is required")
	}
	if p.Latitude == "" || p.Longitude == "" {
		return errors.New("latitude and longitude are required")
	}
	if p.Name == "" {
		return errors.New("name is required")
	}
	if p.Country == "" {
		return errors.New("country is required")
	}

	return nil
}
