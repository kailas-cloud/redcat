package model

type Place struct {
	FsqPlaceID        string `json:"fsq_place_id"`
	Name              string `json:"name"`
	Latitude          string `json:"latitude"`
	Longitude         string `json:"longitude"`
	Address           string `json:"address"`
	Locality          string `json:"locality"`
	Region            string `json:"region"`
	Postcode          string `json:"postcode"`
	AdminRegion       string `json:"admin_region"`
	PostTown          string `json:"post_town"`
	PoBox             string `json:"po_box"`
	Country           string `json:"country"`
	DateCreated       string `json:"date_created"`
	DateRefreshed     string `json:"date_refreshed"`
	DateClosed        string `json:"date_closed"`
	Tel               string `json:"tel"`
	Website           string `json:"website"`
	Email             string `json:"email"`
	FacebookID        string `json:"facebook_id"`
	Instagram         string `json:"instagram"`
	Twitter           string `json:"twitter"`
	FsqCategoryIDs    string `json:"fsq_category_ids"`
	FsqCategoryLabels string `json:"fsq_category_labels"`
	PlacemakerURL     string `json:"placemaker_url"`
	Geom              string `json:"geom"`
	BBox              string `json:"bbox"`
	Dt                string `json:"dt"`
}
