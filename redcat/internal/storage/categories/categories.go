package categories

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"redcat/internal/model"

	"github.com/redis/rueidis"
)

const categoriesIndexKey = "categories:index"

type CategoryStorage struct {
	client rueidis.Client
}

func New(c rueidis.Client) *CategoryStorage {
	return &CategoryStorage{
		client: c,
	}
}

func (s *CategoryStorage) LooksAlike(ctx context.Context, vec []float32, k int64) ([]model.Category, error) {
	cmd := s.client.B().Vsim().Key(categoriesIndexKey).Fp32().Vector(encodeFP32LE(vec)).Withattribs().Count(k).Build()
	resp, err := s.client.Do(ctx, cmd).ToArray()
	if err != nil {
		return nil, err
	}
	fmt.Println(resp)

	return nil, nil
}

func encodeFP32LE(vec []float32) string {
	b := make([]byte, 4*len(vec))
	for i, v := range vec {
		binary.LittleEndian.PutUint32(b[i*4:], math.Float32bits(v))
	}
	return string(b)
}
