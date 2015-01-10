package fpc

import (
	"log"
	"testing"
)

func TestLongToByteArray(t *testing.T) {
	comp := NewCompressor(16)

	for i := 0; i < 10000; i++ {
		if comp.ToLong(comp.ToByteArray(int64(i))) != int64(i) {
			t.Fatalf("Bad round trip: %d", i)
		}
	}
}

func TestRoundtripWithTwoValues(t *testing.T) {
	var tests = []struct {
		Vals []float64
	}{
		{[]float64{1.0, 0.0}},
	}

	comp := NewCompressor(16)

	for _, tt := range tests {
		vals := comp.Decompress(comp.Compress(tt.Vals))
		log.Println(vals)
	}
}
