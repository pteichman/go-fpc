package fpc

import (
	"math"
	"testing"
)

func TestLongToByteArray(t *testing.T) {
	comp := NewCompressor(16)

	for i := 0; i < 10000; i++ {
		if comp.ToLong(comp.ToByteArray(int64(i))) != int64(i) {
			t.Fatalf("Failed value round trip: %d", i)
		}
	}
}

func TestRoundtripWithTwoValues(t *testing.T) {
	var tests = []struct {
		Vals []float64
	}{
		{[]float64{1.0, 0.0}},
		{[]float64{3.0, 0.0}},
	}

	for _, tt := range tests {
		comp := NewCompressor(16)
		vals := comp.Decompress(comp.Compress(tt.Vals))

		if !almostEqualSlice(tt.Vals, vals) {
			t.Fatalf("Failed round trip: %v (got %v)", tt.Vals, vals)
		}
	}
}

func almostEqualSlice(v1, v2 []float64) bool {
	if len(v1) != len(v2) {
		return false
	}

	for i := 0; i < len(v1); i++ {
		if !almostEqual(v1[i], v2[i]) {
			return false
		}
	}

	return true
}

func almostEqual(v1, v2 float64) bool {
	return math.Abs(v1-v2) < 0.00000001
}
