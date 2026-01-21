package valkey

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/rueidis"
)

type Client struct {
	R rueidis.Client
}

func NewClient(addrs []string, username, password string) (*Client, error) {
	opt := rueidis.ClientOption{
		InitAddress: addrs,
	}
	if username != "" {
		opt.Username = username
	}
	if password != "" {
		opt.Password = password
	}
	r, err := rueidis.NewClient(opt)
	if err != nil {
		return nil, err
	}
	return &Client{R: r}, nil
}

func (c *Client) Close() { c.R.Close() }

func EnsurePlacesIndex(ctx context.Context, r rueidis.Client, index, prefix string) error {
	// FT.INFO to check existence
	if err := r.Do(ctx, r.B().FtInfo().Index(index).Build()).Error(); err == nil {
		return nil
	}
	create := r.B().FtCreate().
		Index(index).
		OnHash().
		Prefix(1).Prefix(prefix).
Schema().
		FieldName("category_ids").Tag().
		FieldName("country").Tag().
		FieldName("lat").Numeric().
		FieldName("lon").Numeric().
		FieldName("location").Vector(
			"FLAT",
			6,
			[]string{"TYPE", "FLOAT32", "DIM", "3", "DISTANCE_METRIC", "L2"}...,
		).Build()
	if err := r.Do(ctx, create).Error(); err != nil {
		if strings.Contains(err.Error(), "Index already exists") {
			return nil
		}
		return fmt.Errorf("FT.CREATE %s failed: %w", index, err)
	}
	return nil
}
