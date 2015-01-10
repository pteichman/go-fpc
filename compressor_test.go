package fpc

import (
	"math"
	"reflect"
	"testing"
)

func TestCountLeadingZeros(t *testing.T) {
	var tests = []struct {
		N        uint64
		Expected int
	}{
		{0, 64},
		{0xFFFFFFFF, 32},
		{0xFFFFFFFFFFFFFFFF, 0},
	}

	for _, tt := range tests {
		v := countLeadingZeros(tt.N)
		if v != tt.Expected {
			t.Errorf("Leading zeros for %d == %d (expected %d)", tt.N, v, tt.Expected)
		}
	}
}

func TestLongToByteArray(t *testing.T) {
	comp := NewCompressor(16)

	for i := 0; i < 10000; i++ {
		if comp.ToLong(comp.ToByteArray(int64(i))) != int64(i) {
			t.Fatalf("Failed value round trip: %d", i)
		}
	}
}

func TestEncode(t *testing.T) {
	var tests = []struct {
		Vals     []float64
		Expected []byte
	}{
		{[]float64{0.0}, []byte{0x76, 0x0}},
		{[]float64{1.0}, []byte{0x06, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0, 0x3f, 0x0}},
		{[]float64{2.0}, []byte{0x06, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x0}},
		{[]float64{1.0, 2.0}, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0, 0x3f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40}},
	}

	for _, tt := range tests {
		comp := NewCompressor(16)
		b := comp.Compress(tt.Vals).Bytes()

		if !reflect.DeepEqual(b, tt.Expected) {
			t.Errorf("Failed encode: %x (expected %x)", b, tt.Expected)
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
			t.Errorf("Failed round trip: %v (got %v)", tt.Vals, vals)
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
