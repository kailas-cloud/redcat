package geo

import "testing"

func almost(a, b, eps float64) bool {
	if a > b { return a-b < eps }
	return b-a < eps
}

func TestToECEF_Equator_PrimeMeridian(t *testing.T) {
	v := ToECEF(0, 0)
	if !almost(float64(v[0]), 1, 1e-6) || !almost(float64(v[1]), 0, 1e-6) || !almost(float64(v[2]), 0, 1e-6) {
		t.Fatalf("want (1,0,0) got (%f,%f,%f)", v[0], v[1], v[2])
	}
}

func TestToECEF_Equator_90E(t *testing.T) {
	v := ToECEF(0, 90)
	if !almost(float64(v[0]), 0, 1e-6) || !almost(float64(v[1]), 1, 1e-6) || !almost(float64(v[2]), 0, 1e-6) {
		t.Fatalf("want (0,1,0) got (%f,%f,%f)", v[0], v[1], v[2])
	}
}

func TestToECEF_NorthPole(t *testing.T) {
	v := ToECEF(90, 0)
	if !almost(float64(v[0]), 0, 1e-6) || !almost(float64(v[1]), 0, 1e-6) || !almost(float64(v[2]), 1, 1e-6) {
		t.Fatalf("want (0,0,1) got (%f,%f,%f)", v[0], v[1], v[2])
	}
}
