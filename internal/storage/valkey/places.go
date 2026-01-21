package valkey

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"redcat/internal/domain/geo"
	"redcat/internal/domain/model"

	"github.com/redis/rueidis"
)

type PlacesStorage struct {
	cli       rueidis.Client
	index     string
	keyPrefix string
}

func NewPlacesStorage(cli rueidis.Client, index, prefix string) *PlacesStorage {
	return &PlacesStorage{cli: cli, index: index, keyPrefix: prefix}
}

func (s *PlacesStorage) key(id string) string { return s.keyPrefix + "{" + id + "}" }

func joinCats(ids []string) string {
	clean := make([]string, 0, len(ids))
	for _, v := range ids {
		v = strings.TrimSpace(v)
		if v != "" {
			clean = append(clean, v)
		}
	}
	return strings.Join(clean, ",")
}

func (s *PlacesStorage) Upsert(ctx context.Context, p model.Place) error {
	if p.ID == "" {
		return errors.New("empty id")
	}
	vec := geo.ToECEF(p.Lat, p.Lon)
	cmd := s.cli.B().
		Hset().
		Key(s.key(p.ID)).
		FieldValue().
		FieldValue("id", p.ID).
		FieldValue("name", p.Name).
		FieldValue("lat", fmt.Sprintf("%f", p.Lat)).
		FieldValue("lon", fmt.Sprintf("%f", p.Lon)).
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
		FieldValue("category_ids", joinCats(p.CategoryIDs)).
		FieldValue("placemaker_url", p.PlacemakerURL).
		FieldValue("bbox_xmin", fmt.Sprintf("%f", p.BBox.XMin)).
		FieldValue("bbox_ymin", fmt.Sprintf("%f", p.BBox.YMin)).
		FieldValue("bbox_xmax", fmt.Sprintf("%f", p.BBox.XMax)).
		FieldValue("bbox_ymax", fmt.Sprintf("%f", p.BBox.YMax)).
		FieldValue("dt", p.Dt).
		FieldValue("location", rueidis.VectorString32(vec[:])).
		Build()
	return s.cli.Do(ctx, cmd).Error()
}

func (s *PlacesStorage) Get(ctx context.Context, id string) (model.Place, error) {
	if id == "" { return model.Place{}, errors.New("empty id") }
	m, err := s.cli.Do(ctx, s.cli.B().Hgetall().Key(s.key(id)).Build()).AsStrMap()
	if err != nil { return model.Place{}, err }
p := model.Place{
		ID:      m["id"],
		Name:    m["name"],
		Address: m["address"],
		Locality: m["locality"],
		Region: m["region"],
		Postcode: m["postcode"],
		AdminRegion: m["admin_region"],
		PostTown: m["post_town"],
		PoBox: m["po_box"],
		Country: m["country"],
		DateCreated: m["date_created"],
		DateRefreshed: m["date_refreshed"],
		DateClosed: m["date_closed"],
		Tel: m["tel"],
		Website: m["website"],
		Email: m["email"],
		FacebookID: m["facebook_id"],
		Instagram: m["instagram"],
		Twitter: m["twitter"],
		PlacemakerURL: m["placemaker_url"],
		Dt: m["dt"],
	}
fmt.Sscanf(m["lat"], "%f", &p.Lat)
fmt.Sscanf(m["lon"], "%f", &p.Lon)
if cats := strings.TrimSpace(m["category_ids"]); cats != "" {
		p.CategoryIDs = strings.Split(cats, ",")
}
// bbox (optional fields)
fmt.Sscanf(m["bbox_xmin"], "%f", &p.BBox.XMin)
fmt.Sscanf(m["bbox_ymin"], "%f", &p.BBox.YMin)
fmt.Sscanf(m["bbox_xmax"], "%f", &p.BBox.XMax)
fmt.Sscanf(m["bbox_ymax"], "%f", &p.BBox.YMax)
return p, nil
}

func (s *PlacesStorage) Delete(ctx context.Context, id string) error {
	if id == "" { return errors.New("empty id") }
	return s.cli.Do(ctx, s.cli.B().Del().Key(s.key(id)).Build()).Error()
}

type SearchParams struct {
	Lat, Lon float64
	Limit    int64
	CategoryIDs []string
}

type SearchResult struct {
	Place      model.Place
	DistanceM  float64
}

func knnQuery(limit int64, cats []string) string {
	filter := "*"
	if len(cats) > 0 {
		or := strings.Join(cats, "|")
		filter = fmt.Sprintf("@category_ids:{%s}", or)
	}
	return fmt.Sprintf("%s=>[KNN %d @location $vec]", filter, limit)
}

func (s *PlacesStorage) SearchNearest(ctx context.Context, sp SearchParams) ([]SearchResult, error) {
	if sp.Limit <= 0 || sp.Limit > 200 { sp.Limit = 100 }
	vec := geo.ToECEF(sp.Lat, sp.Lon)
	query := knnQuery(sp.Limit, sp.CategoryIDs)

	cmd := s.cli.B().FtSearch().
		Index(s.index).
		Query(query).
		Return("4").Identifier("id").Identifier("name").Identifier("lat").Identifier("lon").
		Limit().OffsetNum(0, sp.Limit).
		Params().Nargs(2).NameValue().NameValue("vec", rueidis.VectorString32(vec[:])).
		Dialect(2).
		Build()
	arr, err := s.cli.Do(ctx, cmd).ToArray()
	if err != nil { return nil, err }
	if len(arr) == 0 { return nil, nil }
	total, _ := arr[0].AsInt64()
	_ = total

	res := make([]SearchResult, 0, (len(arr)-1)/2)
	for i := 1; i+1 < len(arr); i += 2 {
		m, err := arr[i+1].AsStrMap(); if err != nil { continue }
		var p model.Place
		p.ID = m["id"]
		p.Name = m["name"]
		fmt.Sscanf(m["lat"], "%f", &p.Lat)
		fmt.Sscanf(m["lon"], "%f", &p.Lon)
		d := haversineMeters(sp.Lat, sp.Lon, p.Lat, p.Lon)
		res = append(res, SearchResult{Place: p, DistanceM: d})
	}
	// results already sorted by KNN ASC; ensure stable by distance
	sort.SliceStable(res, func(i,j int) bool { return res[i].DistanceM < res[j].DistanceM })
	if int64(len(res)) > sp.Limit { res = res[:sp.Limit] }
	return res, nil
}

func haversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000.0
	toRad := func(d float64) float64 { return d * 3.141592653589793 / 180 }
	dlat := toRad(lat2-lat1)
	dlon := toRad(lon2-lon1)
	alat := toRad(lat1)
	blat := toRad(lat2)
	a := (sin2(dlat/2) + mathCos(alat)*mathCos(blat)*sin2(dlon/2))
	c := 2 * mathAtan2Sqrt(a)
	return R * c
}

func sin2(x float64) float64 { s := math.Sin(x); return s * s }
func mathCos(x float64) float64 { return math.Cos(x) }
func mathAtan2Sqrt(a float64) float64 { return math.Asin(math.Min(1, math.Sqrt(a))) }
