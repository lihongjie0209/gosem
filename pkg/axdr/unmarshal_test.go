package axdr

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalData(t *testing.T) {
	type TestData struct {
		Time1  time.Time
		Value1 uint16
		Value2 int
		Value3 uint
	}

	src := decodeHexString("01020204090C07D00106050F0030FF003C01121234050ABBCCDD11010204090C07D00106050F0030FFFF880112567805000000001102")

	dec := NewDataDecoder(&src)
	data, err := dec.Decode(&src)
	assert.NoError(t, err)

	var result []TestData
	err = UnmarshalData(data, &result)
	assert.NoError(t, err)

	require.Len(t, result, 2)
	assert.Equal(t, time.Date(2000, time.January, 6, 15, 0, 48, 0, time.FixedZone("UTC-1", -3600)).Unix(), result[0].Time1.Unix())
	assert.Equal(t, time.Date(2000, time.January, 6, 15, 0, 48, 0, time.FixedZone("UTC+2", +7200)).Unix(), result[1].Time1.Unix())
	assert.Equal(t, uint16(0x1234), result[0].Value1)
	assert.Equal(t, int(0xABBCCDD), result[0].Value2)
	assert.Equal(t, uint(0x02), result[1].Value3)
}

func TestUnmarshalDataInUnexportedField(t *testing.T) {
	type TestData struct {
		Time1  time.Time
		Value1 uint16
		Value2 int
		value3 uint //nolint:unused
	}

	src := decodeHexString("01020204090C07D00106050F0030FF003C01121234050ABBCCDD11010204090C07D00106050F0030FFFF880112567805000000001102")

	dec := NewDataDecoder(&src)
	data, err := dec.Decode(&src)
	assert.NoError(t, err)

	var result []TestData
	err = UnmarshalData(data, &result)
	assert.Error(t, err)
}

func TestUnmarshalDataWithNull(t *testing.T) {
	type TestData struct {
		Value1 *uint16
		Value2 int
		Value3 *uint8
	}

	src := decodeHexString("020300050ABBCCDD1101")

	dec := NewDataDecoder(&src)
	data, err := dec.Decode(&src)
	assert.NoError(t, err)

	var result TestData

	value1 := uint16(23)
	result.Value1 = &value1

	err = UnmarshalData(data, &result)
	assert.NoError(t, err)
	assert.Empty(t, result.Value1)
	assert.Equal(t, int(0xABBCCDD), result.Value2)
	assert.NotEmpty(t, result.Value3)
	assert.Equal(t, uint8(1), *result.Value3)
}

func TestUnmarshalPartial(t *testing.T) {
	src := decodeHexString("01020204090C07D00106050F0030FF003C01121234050ABBCCDD11010204090C07D00106050F0030FF003C0112567805000000001102")
	var result [][]DlmsData

	dec := NewDataDecoder(&src)
	data, err := dec.Decode(&src)
	assert.NoError(t, err)

	err = UnmarshalData(data, &result)
	assert.NoError(t, err)

	require.Len(t, result, 2)
	assert.Len(t, result[0], 4)
}

func TestUnmarshalDataFail(t *testing.T) {
	// nil data
	src := decodeHexString("0102020312123405000000001101020312567805000000001102")

	dec := NewDataDecoder(&src)
	data, err := dec.Decode(&src)
	assert.NoError(t, err)

	err = UnmarshalData(data, nil)
	assert.Error(t, err)

	// Invalid variable kind
	type invalidTestData struct {
		Value1 uint16
		Value2 int32
		Value3 uint16
	}

	var invalidResult []invalidTestData

	err = UnmarshalData(data, &invalidResult)
	assert.Error(t, err)

	// Missing struct element
	type anotherInvalidTestData struct {
		Value1 uint16
		Value2 int32
	}

	var anotherInvalidResult []anotherInvalidTestData

	err = UnmarshalData(data, &anotherInvalidResult)
	assert.Error(t, err)

	// Invalid time format
	src = decodeHexString("090B07D00106050F0030FF003C")
	dec = NewDataDecoder(&src)
	data, _ = dec.Decode(&src)

	var time1 time.Time

	err = UnmarshalData(data, &time1)
	assert.Error(t, err)
}

func decodeHexString(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
