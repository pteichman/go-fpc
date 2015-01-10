package fpc

import (
	"bytes"
	"math"
)

type Compressor struct {
	logOfTableSize int
	pred1          *FcmPredictor
	pred2          *DfcmPredictor
}

func NewCompressor(logOfTableSize int) *Compressor {
	return &Compressor{
		logOfTableSize: logOfTableSize,
		pred1:          NewFcmPredictor(logOfTableSize),
		pred2:          NewDfcmPredictor(logOfTableSize),
	}
}

func (c *Compressor) Compress(vals []float64) *bytes.Buffer {
	var buf bytes.Buffer

	for i := 0; i < len(vals); i += 2 {
		if i == len(vals)-1 {
			c.encodeAndPad(&buf, vals[i])
		} else {
			c.encode(&buf, vals[i], vals[i+1])
		}
	}

	return &buf
}

func (c *Compressor) encodeAndPad(buf *bytes.Buffer, v float64) {
	dbits := int64(math.Float64bits(v))
	diff1d := c.pred1.Prediction() ^ dbits
	diff2d := c.pred2.Prediction() ^ dbits

	pred1better := countLeadingZeros(uint64(diff1d)) >= countLeadingZeros(uint64(diff2d))

	c.pred1.Update(dbits)
	c.pred2.Update(dbits)

	var code byte
	if pred1better {
		zb := encodeZeroBytes(diff1d)
		code |= byte(zb << 4)
	} else {
		zb := encodeZeroBytes(diff2d)
		code |= 0x80
		code |= byte(zb << 4)
	}

	code |= 0x06

	// FIXME: ignoring errors
	buf.WriteByte(code)
	if pred1better {
		buf.Write(c.ToByteArray(diff1d))
	} else {
		buf.Write(c.ToByteArray(diff2d))
	}

	buf.WriteByte(0)
}

func (c *Compressor) encode(buf *bytes.Buffer, d, e float64) {
	dbits := int64(math.Float64bits(d))
	diff1d := c.pred1.Prediction() ^ dbits
	diff2d := c.pred2.Prediction() ^ dbits

	pred1BetterForD := countLeadingZeros(uint64(diff1d)) >= countLeadingZeros(uint64(diff2d))

	c.pred1.Update(dbits)
	c.pred2.Update(dbits)

	ebits := int64(math.Float64bits(e))
	diff1e := c.pred1.Prediction() ^ ebits
	diff2e := c.pred2.Prediction() ^ ebits

	pred1BetterForE := countLeadingZeros(uint64(diff1e)) >= countLeadingZeros(uint64(diff2e))

	c.pred1.Update(ebits)
	c.pred2.Update(ebits)

	var code byte
	if pred1BetterForD {
		zb := encodeZeroBytes(diff1d)
		code |= byte(zb << 4)
	} else {
		zb := encodeZeroBytes(diff2d)
		code |= 0x80
		code |= byte(zb << 4)
	}

	if pred1BetterForE {
		zb := encodeZeroBytes(diff1e)
		code |= byte(zb)
	} else {
		zb := encodeZeroBytes(diff2e)
		code |= 0x08
		code |= byte(zb)
	}

	// FIXME: ignoring errors
	buf.WriteByte(code)
	if pred1BetterForD {
		buf.Write(c.ToByteArray(diff1d))
	} else {
		buf.Write(c.ToByteArray(diff2d))
	}

	if pred1BetterForE {
		buf.Write(c.ToByteArray(diff1e))
	} else {
		buf.Write(c.ToByteArray(diff2e))
	}
}

func (c *Compressor) Decompress(buf *bytes.Buffer) []float64 {
	var ret []float64

	for i := 0; i < buf.Len(); i += 2 {
		ret = append(ret, c.decode(buf)...)
	}

	return ret
}

func (c *Compressor) decode(buf *bytes.Buffer) []float64 {
	head, _ := buf.ReadByte()

	var pred int64
	if (head & 0x80) != 0 {
		pred = c.pred2.Prediction()
	} else {
		pred = c.pred1.Prediction()
	}

	nzb := (head & 0x70) >> 4
	if nzb > 3 {
		nzb++
	}

	dst := make([]byte, 8-nzb)

	// FIXME: errors
	buf.Read(dst)

	diff := c.ToLong(dst)
	actual := pred ^ diff

	c.pred1.Update(actual)
	c.pred2.Update(actual)

	var ret []float64
	ret = append(ret, math.Float64frombits(uint64(actual)))

	if (head & 0x08) != 0 {
		pred = c.pred2.Prediction()
	} else {
		pred = c.pred1.Prediction()
	}

	nzb = head & 0x07
	if nzb > 3 {
		nzb++
	}

	dst = make([]byte, 8-nzb)

	// FIXME: errors
	buf.Read(dst)

	diff = c.ToLong(dst)

	if nzb == 7 && diff == 0 {
		return ret
	}

	actual = pred ^ diff

	c.pred1.Update(actual)
	c.pred2.Update(actual)

	return append(ret, math.Float64frombits(uint64(actual)))
}

func (c *Compressor) ToByteArray(diff int64) []byte {
	ezb := encodeZeroBytes(diff)

	a := make([]byte, 8-ezb)
	for i := 0; i < len(a); i++ {
		a[i] = byte(diff) & 0xFF
		diff = diff >> 8
	}

	return a
}

func (c *Compressor) ToLong(buf []byte) int64 {
	var result int64
	for i := len(buf); i > 0; i-- {
		result = result << 8
		result |= int64(buf[i-1] & 0xFF)
	}
	return result
}

func encodeZeroBytes(diff1d int64) int32 {
	lzb := int32(countLeadingZeros(uint64(diff1d)) / 8)
	if lzb >= 4 {
		lzb--
	}
	return lzb
}

func countLeadingZeros(n uint64) int {
	// There are plenty of better ways to do this, but use the sign of n
	// to check its high bit.
	for i := 0; i < 64; i++ {
		if n&0x8000000000000000 != 0 {
			return i
		}
		n = n << 1
	}

	return 64
}
