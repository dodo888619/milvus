package indexparamcheck

import (
	"strconv"
	"testing"

	"github.com/milvus-io/milvus-proto/go-api/v2/schemapb"
	"github.com/milvus-io/milvus/pkg/util/metric"

	"github.com/stretchr/testify/assert"
)

func Test_raftIVFPQChecker_CheckTrain(t *testing.T) {

	validParams := map[string]string{
		DIM:    strconv.Itoa(128),
		NLIST:  strconv.Itoa(1024),
		IVFM:   strconv.Itoa(4),
		NBITS:  strconv.Itoa(8),
		Metric: metric.L2,
	}

	validParamsWithoutNbits := map[string]string{
		DIM:    strconv.Itoa(128),
		NLIST:  strconv.Itoa(1024),
		IVFM:   strconv.Itoa(4),
		Metric: metric.L2,
	}

	validParamsWithoutDim := map[string]string{
		NLIST:  strconv.Itoa(1024),
		IVFM:   strconv.Itoa(4),
		NBITS:  strconv.Itoa(8),
		Metric: metric.L2,
	}

	invalidParamsDim := copyParams(validParams)
	invalidParamsDim[DIM] = "NAN"

	invalidParamsNbits := copyParams(validParams)
	invalidParamsNbits[NBITS] = "NAN"

	invalidParamsWithoutIVF := map[string]string{
		DIM:    strconv.Itoa(128),
		NLIST:  strconv.Itoa(1024),
		NBITS:  strconv.Itoa(8),
		Metric: metric.L2,
	}

	invalidParamsIVF := copyParams(validParams)
	invalidParamsIVF[IVFM] = "NAN"

	invalidParamsM := copyParams(validParams)
	invalidParamsM[DIM] = strconv.Itoa(65536)

	validParamsMzero := copyParams(validParams)
	validParamsMzero[IVFM] = "0"

	p1 := map[string]string{
		DIM:    strconv.Itoa(128),
		NLIST:  strconv.Itoa(1024),
		IVFM:   strconv.Itoa(4),
		NBITS:  strconv.Itoa(8),
		Metric: metric.L2,
	}
	p2 := map[string]string{
		DIM:    strconv.Itoa(128),
		NLIST:  strconv.Itoa(1024),
		IVFM:   strconv.Itoa(4),
		NBITS:  strconv.Itoa(8),
		Metric: metric.IP,
	}
	p3 := map[string]string{
		DIM:    strconv.Itoa(128),
		NLIST:  strconv.Itoa(1024),
		IVFM:   strconv.Itoa(4),
		NBITS:  strconv.Itoa(8),
		Metric: metric.COSINE,
	}
	p4 := map[string]string{
		DIM:    strconv.Itoa(128),
		NLIST:  strconv.Itoa(1024),
		IVFM:   strconv.Itoa(4),
		NBITS:  strconv.Itoa(8),
		Metric: metric.HAMMING,
	}
	p5 := map[string]string{
		DIM:    strconv.Itoa(128),
		NLIST:  strconv.Itoa(1024),
		IVFM:   strconv.Itoa(4),
		NBITS:  strconv.Itoa(8),
		Metric: metric.JACCARD,
	}

	cases := []struct {
		params   map[string]string
		errIsNil bool
	}{
		{validParams, true},
		{validParamsWithoutNbits, true},
		{invalidIVFParamsMin(), false},
		{invalidIVFParamsMax(), false},
		{validParamsWithoutDim, false},
		{invalidParamsDim, false},
		{invalidParamsNbits, false},
		{invalidParamsWithoutIVF, false},
		{invalidParamsIVF, false},
		{invalidParamsM, false},
		{validParamsMzero, true},
		{p1, true},
		{p2, true},
		{p3, true},
		{p4, false},
		{p5, false},
	}

	c := newRaftIVFPQChecker()
	for _, test := range cases {
		err := c.CheckTrain(test.params)
		if test.errIsNil {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func Test_raftIVFPQChecker_CheckValidDataType(t *testing.T) {
	cases := []struct {
		dType    schemapb.DataType
		errIsNil bool
	}{
		{
			dType:    schemapb.DataType_Bool,
			errIsNil: false,
		},
		{
			dType:    schemapb.DataType_Int8,
			errIsNil: false,
		},
		{
			dType:    schemapb.DataType_Int16,
			errIsNil: false,
		},
		{
			dType:    schemapb.DataType_Int32,
			errIsNil: false,
		},
		{
			dType:    schemapb.DataType_Int64,
			errIsNil: false,
		},
		{
			dType:    schemapb.DataType_Float,
			errIsNil: false,
		},
		{
			dType:    schemapb.DataType_Double,
			errIsNil: false,
		},
		{
			dType:    schemapb.DataType_String,
			errIsNil: false,
		},
		{
			dType:    schemapb.DataType_VarChar,
			errIsNil: false,
		},
		{
			dType:    schemapb.DataType_Array,
			errIsNil: false,
		},
		{
			dType:    schemapb.DataType_JSON,
			errIsNil: false,
		},
		{
			dType:    schemapb.DataType_FloatVector,
			errIsNil: true,
		},
		{
			dType:    schemapb.DataType_BinaryVector,
			errIsNil: false,
		},
	}

	c := newRaftIVFPQChecker()
	for _, test := range cases {
		err := c.CheckValidDataType(test.dType)
		if test.errIsNil {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}
