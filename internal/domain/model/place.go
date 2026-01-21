package model

type Place struct {
	ID          string   `json:"id"`               // fsq_place_id
	Name        string   `json:"name"`
	Lat         float64  `json:"lat"`              // latitude
	Lon         float64  `json:"lon"`              // longitude
	Address     string   `json:"address,omitempty"`
	Locality    string   `json:"locality,omitempty"`
	Region      string   `json:"region,omitempty"`
	Postcode    string   `json:"postcode,omitempty"`
	AdminRegion string   `json:"admin_region,omitempty"`
	PostTown    string   `json:"post_town,omitempty"`
	PoBox       string   `json:"po_box,omitempty"`
	Country     string   `json:"country,omitempty"`
	DateCreated   string  `json:"date_created,omitempty"`
	DateRefreshed string  `json:"date_refreshed,omitempty"`
	DateClosed    string  `json:"date_closed,omitempty"`
	Tel         string   `json:"tel,omitempty"`
	Website     string   `json:"website,omitempty"`
	Email       string   `json:"email,omitempty"`
	FacebookID  string   `json:"facebook_id,omitempty"`
	Instagram   string   `json:"instagram,omitempty"`
	Twitter     string   `json:"twitter,omitempty"`
	CategoryIDs []string `json:"category_ids"`         // derived from fsq_category_ids
	CategoryLabels []string `json:"category_labels,omitempty"`
	PlacemakerURL string  `json:"placemaker_url,omitempty"`
	// bbox kept flat for simplicity; geom omitted
	BBox struct {
		XMin float64 `json:"xmin,omitempty"`
		YMin float64 `json:"ymin,omitempty"`
		XMax float64 `json:"xmax,omitempty"`
		YMax float64 `json:"ymax,omitempty"`
	} `json:"bbox,omitempty"`
	Dt string `json:"dt,omitempty"`
}

// PlaceDoc: storage representation for HSET; vector stored separately in `location`.
type PlaceDoc struct {
	ID          string
	Name        string
	Lat         float64
	Lon         float64
	Address     string
	Locality    string
	Region      string
	Postcode    string
	AdminRegion string
	PostTown    string
	PoBox       string
	Country     string
	DateCreated   string
	DateRefreshed string
	DateClosed    string
	Tel         string
	Website     string
	Email       string
	FacebookID  string
	Instagram   string
	Twitter     string
	CategoryIDs string // comma-separated for TAG
	CategoryLabels string
	PlacemakerURL string
	BBoxXMin float64
	BBoxYMin float64
	BBoxXMax float64
	BBoxYMax float64
	Dt string
}
