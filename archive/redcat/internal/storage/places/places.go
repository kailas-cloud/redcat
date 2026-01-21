package places

import (
	"context"
	"fmt"

	"redcat/internal/model"

	"github.com/redis/rueidis"
)

const placeKey = "place:{%s}"

func makeKey(placeId string) string {
	return fmt.Sprintf(placeKey, placeId)
}

type Storage struct {
	rdb rueidis.Client
}

func New(rdb rueidis.Client) *Storage {
	return &Storage{
		rdb: rdb,
	}
}

func (s *Storage) Upsert(ctx context.Context, p model.Place) error {
	if p.FsqPlaceID == "" {
		return fmt.Errorf("empty FsqPlaceID")
	}

	cmd := s.rdb.B().
		Hset().
		Key(makeKey(p.FsqPlaceID)).
		FieldValue().
		FieldValue("fsq_place_id", p.FsqPlaceID).
		FieldValue("name", p.Name).
		FieldValue("latitude", p.Latitude).
		FieldValue("longitude", p.Longitude).
		FieldValue("address", p.Address).
		FieldValue("locality", p.Locality).
		FieldValue("region", p.Region).
		FieldValue("postcode", p.Postcode).
		FieldValue("admin_region", p.AdminRegion).
		FieldValue("post_town", p.PostTown).
		FieldValue("po_box", p.PoBox).
		FieldValue("country", p.Country).
		FieldValue("date_created", p.DateCreated).
		FieldValue("date_refreshed", p.DateRefreshed).
		FieldValue("date_closed", p.DateClosed).
		FieldValue("tel", p.Tel).
		FieldValue("website", p.Website).
		FieldValue("email", p.Email).
		FieldValue("facebook_id", p.FacebookID).
		FieldValue("instagram", p.Instagram).
		FieldValue("twitter", p.Twitter).
		FieldValue("fsq_category_ids", p.FsqCategoryIDs).
		FieldValue("fsq_category_labels", p.FsqCategoryLabels).
		FieldValue("placemaker_url", p.PlacemakerURL).
		FieldValue("geom", p.Geom).
		FieldValue("bbox", p.BBox).
		FieldValue("dt", p.Dt).
		Build()

	return s.rdb.Do(ctx, cmd).Error()
}
