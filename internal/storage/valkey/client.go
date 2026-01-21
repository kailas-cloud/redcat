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
		// Connection pooling for high throughput
		BlockingPoolSize: 0, // Use pipeline mode (multiplexing) by default
		PipelineMultiplex: 128, // Max concurrent pipeline requests per connection
		// Ring buffer for batching commands
		RingScaleEachConn: 10, // 2^10 = 1024 ring buffer size
		// Disable client-side caching for write-heavy workload
		DisableCache: true,
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
