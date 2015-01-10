package fpc

type DfcmPredictor struct {
	Table     []int64
	DfcmHash  int32
	LastValue int64
}

func NewDfcmPredictor(logOfTableSize int) *DfcmPredictor {
	return &DfcmPredictor{
		Table: make([]int64, logOfTableSize),
	}
}

func (dp *DfcmPredictor) Prediction() int64 {
	return dp.Table[dp.DfcmHash] + dp.LastValue
}

func (dp *DfcmPredictor) Update(trueValue int64) {
	dp.Table[dp.DfcmHash] = trueValue - dp.LastValue
	dp.DfcmHash = int32(((dp.DfcmHash << 2) ^ int32((trueValue-dp.LastValue)>>40)) & int32(len(dp.Table)-1))
	dp.LastValue = trueValue
}

type FcmPredictor struct {
	Table   []int64
	FcmHash int32
}

func NewFcmPredictor(logOfTableSize int) *FcmPredictor {
	return &FcmPredictor{
		Table: make([]int64, logOfTableSize),
	}
}

func (dp *FcmPredictor) Prediction() int64 {
	return dp.Table[dp.FcmHash]
}

func (dp *FcmPredictor) Update(trueValue int64) {
	dp.Table[dp.FcmHash] = trueValue
	dp.FcmHash = int32(((dp.FcmHash << 6) ^ int32(trueValue>>48)) & int32(len(dp.Table)-1))
}
