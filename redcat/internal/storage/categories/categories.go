package categories

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"log"
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
	blob := encodeFP32LE(vec)
	cmd := s.client.B().Vsim().Key(categoriesIndexKey).
		Fp32().
		Vector(rueidis.BinaryString(blob)).
		Withattribs().
		Count(k).
		Build()
	resp, err := s.client.Do(ctx, cmd).AsStrMap()
	if err != nil {
		return nil, err
	}

	var attrs struct {
		Name  string `json:"name"`
		Label string `json:"label"`
	}

	results := make([]model.Category, 0, len(resp))
	for id, msg := range resp {
		if err := json.Unmarshal([]byte(msg), &attrs); err != nil {
			log.Printf("Error unmarshalling attrs: %s", msg)
			continue
		}

		results = append(results, model.Category{
			ID:    id,
			Name:  attrs.Name,
			Label: attrs.Label,
		})
	}

	return results, nil
}

func encodeFP32LE(vec []float32) []byte {
	b := make([]byte, 4*len(vec))
	for i, v := range vec {
		binary.LittleEndian.PutUint32(b[i*4:], math.Float32bits(v))
	}
	return b
}
