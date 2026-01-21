package geo

import "math"

// ToECEF returns unit-sphere ECEF vector as 3 float32 values.
func ToECEF(latDeg, lonDeg float64) [3]float32 {
	lat := latDeg * math.Pi / 180
	lon := lonDeg * math.Pi / 180
	x := math.Cos(lat) * math.Cos(lon)
	y := math.Cos(lat) * math.Sin(lon)
	z := math.Sin(lat)
	return [3]float32{float32(x), float32(y), float32(z)}
}

// PackFloat32LE packs the 3 float32 values into little-endian bytes.
func PackFloat32LE(v [3]float32) []byte {
	b := make([]byte, 12)
	u := math.Float32bits(v[0])
	b[0], b[1], b[2], b[3] = byte(u), byte(u>>8), byte(u>>16), byte(u>>24)
	u = math.Float32bits(v[1])
	b[4], b[5], b[6], b[7] = byte(u), byte(u>>8), byte(u>>16), byte(u>>24)
	u = math.Float32bits(v[2])
	b[8], b[9], b[10], b[11] = byte(u), byte(u>>8), byte(u>>16), byte(u>>24)
	return b
}
