package axdr

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeLength(t *testing.T) {
	encoded, err := EncodeLength(125)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("7D"), encoded)

	encoded, err = EncodeLength(128)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("8180"), encoded)

	encoded, err = EncodeLength(255)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("81FF"), encoded)

	encoded, err = EncodeLength(256)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("820100"), encoded)

	encoded, err = EncodeLength(65535)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("82FFFF"), encoded)

	encoded, err = EncodeLength(65536)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("83010000"), encoded)

	encoded, err = EncodeLength(uint(18446744073709551615))
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("88FFFFFFFFFFFFFFFF"), encoded)

	encoded, err = EncodeLength(uint64(18446744073709551615))
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("88FFFFFFFFFFFFFFFF"), encoded)

	_, err = EncodeLength("123")
	assert.Error(t, err)

	_, err = EncodeLength(3.14)
	assert.Error(t, err)

	_, err = EncodeLength(-500)
	assert.Error(t, err)

	_, err = EncodeLength(int64(-500000000))
	assert.Error(t, err)

	encoded, err = EncodeLength(0)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("00"), encoded)
}

func TestEncodeBoolean(t *testing.T) {
	encoded, err := EncodeBoolean(true)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("01"), encoded)

	encoded, err = EncodeBoolean(false)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("00"), encoded)
}

func TestEncodeBitString(t *testing.T) {
	encoded, err := EncodeBitString("11111000")
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("F8"), encoded)

	encoded, err = EncodeBitString("111100000001")
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("f010"), encoded)

	encoded, err = EncodeBitString("0000111111110000111111110000000101010101")
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("0ff0ff0155"), encoded)

	encoded, err = EncodeBitString("00001111 11110000 11111111 00000001 01010101 1")
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("0ff0ff015580"), encoded)
}

func TestEncodeDoubleLong(t *testing.T) {
	encoded, err := EncodeDoubleLong(0)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("00000000"), encoded)

	encoded, err = EncodeDoubleLong(255)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("000000FF"), encoded)

	encoded, err = EncodeDoubleLong(-25)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("FFFFFFE7"), encoded)

	encoded, err = EncodeDoubleLong(65535)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("0000FFFF"), encoded)

	encoded, err = EncodeDoubleLong(2147483647)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("7FFFFFFF"), encoded)

	encoded, err = EncodeDoubleLong(-2147483647)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("80000001"), encoded)
}

func TestEncodeDoubleLongUnsigned(t *testing.T) {
	encoded, err := EncodeDoubleLongUnsigned(0)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("00000000"), encoded)

	encoded, err = EncodeDoubleLongUnsigned(255)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("000000FF"), encoded)

	encoded, err = EncodeDoubleLongUnsigned(65535)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("0000FFFF"), encoded)

	encoded, err = EncodeDoubleLongUnsigned(4294967295)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("FFFFFFFF"), encoded)
}

func TestEncodeOctetString(t *testing.T) {
	encoded, err := EncodeOctetString("07D20C04030A060BFF007800")
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("07D20C04030A060BFF007800"), encoded)

	encoded, err = EncodeOctetString("1.0.0.3.0.255")
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("0100000300FF"), encoded)

	encoded, err = EncodeOctetString("07 D2 0C    04")
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("07D20C04"), encoded)
}

func TestEncodeVisibleString(t *testing.T) {
	encoded, err := EncodeVisibleString("ABCD")
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("41424344"), encoded)

	encoded, err = EncodeVisibleString("a1 -")
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("6131202d"), encoded)

	encoded, err = EncodeVisibleString("{}[]()!;")
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("7b7d5b5d2829213b"), encoded)

	_, err = EncodeVisibleString("ÆÁÉÍÓÚ")
	assert.Error(t, err)
}

func TestEncodeUTF8String(t *testing.T) {
	encoded, err := EncodeUTF8String("ABCD")
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("41424344"), encoded)

	encoded, err = EncodeUTF8String("aфᐃ𝕫")
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("61d184e19083f09d95ab"), encoded)

	encoded, err = EncodeUTF8String("我愛你")
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("e68891e6849be4bda0"), encoded)
}

func TestEncodeBCDAndBCDs(t *testing.T) {
	encoded, err := EncodeBCD(int8(127))
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("7F"), encoded)

	encoded, err = EncodeBCD(int8(-1))
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("FF"), encoded)

	encoded, err = EncodeBCDs("1234")
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("1234"), encoded)

	encoded, err = EncodeBCDs("12345")
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("123450"), encoded)
}

func TestEncodeInteger(t *testing.T) {
	encoded, err := EncodeInteger(-128)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("80"), encoded)

	encoded, err = EncodeInteger(0)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("00"), encoded)

	encoded, err = EncodeInteger(127)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("7F"), encoded)

	encoded, err = EncodeInteger(-1)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("FF"), encoded)
}

func TestEncodeLong(t *testing.T) {
	encoded, err := EncodeLong(0)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("0000"), encoded)

	encoded, err = EncodeLong(256)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("0100"), encoded)

	encoded, err = EncodeLong(1<<15 - 1)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("7FFF"), encoded)

	encoded, err = EncodeLong(-1 << 15)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("8000"), encoded)
}

func TestEncodeUnsigned(t *testing.T) {
	encoded, err := EncodeUnsigned(0)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("00"), encoded)

	encoded, err = EncodeUnsigned(255)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("FF"), encoded)
}

func TestEncodeLongUnsigned(t *testing.T) {
	encoded, err := EncodeLongUnsigned(0)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("0000"), encoded)

	encoded, err = EncodeLongUnsigned(1<<16 - 1)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("FFFF"), encoded)
}

func TestEncodeLong64(t *testing.T) {
	encoded, err := EncodeLong64(0)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("0000000000000000"), encoded)

	encoded, err = EncodeLong64(1<<63 - 1)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("7fffffffffffffff"), encoded)

	encoded, err = EncodeLong64(-1 << 63)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("8000000000000000"), encoded)

	encoded, err = EncodeLong64(-1)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("ffffffffffffffff"), encoded)
}

func TestEncodeLong64Unsigned(t *testing.T) {
	encoded, err := EncodeLong64Unsigned(0)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("0000000000000000"), encoded)

	encoded, err = EncodeLong64Unsigned(1<<64 - 1)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("FFFFFFFFFFFFFFFF"), encoded)
}

func TestEncodeFloat32(t *testing.T) {
	encoded, err := EncodeFloat32(0)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("00000000"), encoded)

	encoded, err = EncodeFloat32(float32(3.14))
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("4048f5c3"), encoded)

	encoded, err = EncodeFloat32(float32(-3.14))
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("c048f5c3"), encoded)

	encoded, err = EncodeFloat32(4294967295)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("4f800000"), encoded)
}

func TestEncodeFloat64(t *testing.T) {
	encoded, err := EncodeFloat64(0)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("0000000000000000"), encoded)

	encoded, err = EncodeFloat64(float64(3.14))
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("40091EB851EB851F"), encoded)

	encoded, err = EncodeFloat64(float64(-3.14))
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("C0091EB851EB851F"), encoded)

	encoded, err = EncodeFloat64(4294967295)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("41EFFFFFFFE00000"), encoded)

	encoded, err = EncodeFloat64(3.1415926535)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("400921FB54411744"), encoded)
}

func TestEncodeDate(t *testing.T) {
	dt := time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC)
	encoded, err := EncodeDate(dt)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("07D90B0A02"), encoded)

	dt = time.Date(1500, time.January, 1, 0, 0, 0, 0, time.UTC)
	encoded, err = EncodeDate(dt)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("05DC010101"), encoded)
}

func TestEncodeTime(t *testing.T) {
	dt := time.Date(2020, time.January, 1, 10, 0, 0, 0, time.UTC)
	encoded, err := EncodeTime(dt)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("0A000000"), encoded)

	dt = time.Date(2020, time.January, 1, 23, 59, 59, 990000000, time.UTC)
	encoded, err = EncodeTime(dt)
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("173B3B63"), encoded)
}

func TestEncodeDateTime(t *testing.T) {
	local, _ := time.LoadLocation("Europe/Madrid")

	tests := []struct {
		name              string
		timeZoneDeviation TimeZone
		expected          string
		val               time.Time
	}{
		{"Future time", TimeZoneStandard, "4E200C1E06173B3B63000000", time.Date(20000, time.December, 30, 23, 59, 59, 990000000, time.UTC)},
		{"Past time", TimeZoneStandard, "05DC01010100000000000000", time.Date(1500, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{"Local time", TimeZoneStandard, "07E403100100000000FFC400", time.Date(2020, time.March, 16, 0, 0, 0, 0, local)},
		{"Summer local time", TimeZoneStandard, "07E40701030A000000FF8880", time.Date(2020, time.July, 1, 10, 0, 0, 0, local)},
		{"Local time reversed", TimeZoneReversed, "07E403100100000000003C00", time.Date(2020, time.March, 16, 0, 0, 0, 0, local)},
		{"Local time ignored", TimeZoneIgnored, "07E403100100000000800000", time.Date(2020, time.March, 16, 0, 0, 0, 0, local)},
		{"Sunday", TimeZoneStandard, "07E7010F0700000000000000", time.Date(2023, time.January, 15, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TimeZoneDeviation = tt.timeZoneDeviation
			encoded, err := EncodeDateTime(tt.val)
			TimeZoneDeviation = TimeZoneStandard
			assert.NoError(t, err)
			assert.Equal(t, decodeHexString(tt.expected), encoded)
		})
	}
}

func TestDlmsData(t *testing.T) {
	tDD := DlmsData{Tag: TagBoolean, Value: true}
	encoded, err := tDD.Encode()
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("0301"), encoded)

	tDD = DlmsData{Tag: TagBitString, Value: "0000111111110000111111110000000101010101"}
	encoded, err = tDD.Encode()
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("04280FF0FF0155"), encoded)
}

func TestDlmsData_NilValue(t *testing.T) {
	tDD := DlmsData{Tag: TagBoolean, Value: nil}
	_, err := tDD.Encode()
	assert.Error(t, err)
}

func TestDlmsData_WrongBoolValue(t *testing.T) {
	tDD := DlmsData{Tag: TagBoolean, Value: 1234}
	_, err := tDD.Encode()
	assert.Error(t, err)
}

func TestDlmsData_WrongBitStringValue(t *testing.T) {
	tDD := DlmsData{Tag: TagBitString, Value: "ABCDEFG"}
	_, err := tDD.Encode()
	assert.Error(t, err)
}

func TestDlmsData_DateTime(t *testing.T) {
	tDD := DlmsData{Tag: TagDateTime, Value: "9999-12-30 23:59:59"}
	encoded, err := tDD.Encode()
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("19270f0c1e04173b3b00000000"), encoded)

	dt := time.Date(20000, time.December, 30, 23, 59, 59, 0, time.UTC)
	tDD.Value = dt
	encoded, err = tDD.Encode()
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("194e200c1e06173b3b00000000"), encoded)
}

func TestArray(t *testing.T) {
	d1 := DlmsData{Tag: TagBoolean, Value: true}
	d2 := DlmsData{Tag: TagBitString, Value: "111"}
	d3 := DlmsData{Tag: TagDateTime, Value: "2020-03-11 18:00:00"}

	ls := []*DlmsData{&d1, &d2, &d3}
	ts, err := EncodeArray(ls)
	res := bytes.Compare(ts, []byte{byte(TagBoolean), 1, byte(TagBitString), 3, 224, byte(TagDateTime), 7, 228, 3, 11, 3, 18, 0, 0, 0, 0, 0, 0})
	if res != 0 || err != nil {
		t.Errorf("t1 failed. val: %d, err:%v", ts, err)
	}

	tables := []struct {
		x DlmsData
		y DlmsData
		z DlmsData
		r []byte
	}{
		{d1, d2, d3, []byte{byte(TagBoolean), 1, byte(TagBitString), 3, 224, byte(TagDateTime), 7, 228, 3, 11, 3, 18, 0, 0, 0, 0, 0, 0}},
		{d2, d1, d3, []byte{byte(TagBitString), 3, 224, byte(TagBoolean), 1, byte(TagDateTime), 7, 228, 3, 11, 3, 18, 0, 0, 0, 0, 0, 0}},
		{d3, d2, d1, []byte{byte(TagDateTime), 7, 228, 3, 11, 3, 18, 0, 0, 0, 0, 0, 0, byte(TagBitString), 3, 224, byte(TagBoolean), 1}},
	}
	for idx, table := range tables {
		ts, err = EncodeArray([]*DlmsData{&tables[idx].x, &tables[idx].y, &tables[idx].z})
		assert.NoError(t, err)
		assert.Equal(t, table.r, ts)
	}
}

func TestDlmsData_Array(t *testing.T) {
	d1 := DlmsData{Tag: TagBoolean, Value: true}
	d2 := DlmsData{Tag: TagBitString, Value: "111"}
	d3 := DlmsData{Tag: TagDateTime, Value: "2020-03-11 18:00:00"}
	tDD := DlmsData{Tag: TagArray, Value: []*DlmsData{&d1, &d2, &d3}}
	encoded, err := tDD.Encode()
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("010303010403E01907E4030B0312000000000000"), encoded)

	tDD = DlmsData{Tag: TagArray, Value: []*DlmsData{{Tag: TagBoolean, Value: true}, {Tag: TagBoolean, Value: false}}}
	encoded, err = tDD.Encode()
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("010203010300"), encoded)

	tDD = DlmsData{Tag: TagArray, Value: []*DlmsData{}}
	encoded, err = tDD.Encode()
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("0100"), encoded)
}

// ---------- decoding tests

func TestDecodeLength(t *testing.T) {
	tables := []struct {
		src []byte
		bt  []byte
		val uint64
	}{
		{[]byte{2, 1, 2, 3}, []byte{2}, 2},
		{[]byte{131, 1, 0, 0, 1, 2, 3}, []byte{131, 1, 0, 0}, 65536},
		{[]byte{136, 255, 255, 255, 255, 255, 255, 255, 255, 1, 2, 3}, []byte{136, 255, 255, 255, 255, 255, 255, 255, 255}, 18446744073709551615},
	}
	for idx, table := range tables {
		bt, val, err := DecodeLength(&tables[idx].src)
		if err != nil {
			t.Errorf("combination %v failed. got an error:%v", idx, err)
		}
		// compare length byte
		sameByte := bytes.Compare(table.bt, bt)
		if sameByte != 0 {
			t.Errorf("combination %v failed. Byte get: %d, should:%v", idx, bt, table.bt)
		}
		// compare length value
		sameValue := (table.val == val)
		if !sameValue {
			t.Errorf("combination %v failed. Value get: %d, should:%v", idx, val, table.val)
		}
		// compare remainder bytes of src
		sameReminder := bytes.Compare(tables[idx].src, []byte{1, 2, 3})
		if sameReminder != 0 {
			t.Errorf("combination %v failed. Reminder get: %d, should:[1, 2, 3]", idx, table.src)
		}
	}
}

func TestDecodeBoolean(t *testing.T) {
	src := []byte{255, 1, 2, 3}
	bt, val, err := DecodeBoolean(&src)
	if err != nil {
		t.Errorf("t1 failed. got an error:%v", err)
	}
	sameByte := bytes.Compare(bt, []byte{255})
	if sameByte != 0 {
		t.Errorf("t1 failed. val: %d", sameByte)
	}
	sameValue := (val == true)
	if !sameValue {
		t.Errorf("t1 failed. Value get: %v", val)
	}
	sameReminder := bytes.Compare(src, []byte{1, 2, 3})
	if sameReminder != 0 {
		t.Errorf("t1 failed. Reminder get: %d, should:[1, 2, 3]", src)
	}
}

func TestDecodeBitString(t *testing.T) {
	tables := []struct {
		src []byte
		bt  []byte
		val string
	}{
		{[]byte{248, 1, 2, 3}, []byte{248}, "11111000"},
		{[]byte{15, 240, 255, 1, 85, 1, 2, 3}, []byte{15, 240, 255, 1, 85}, "0000111111110000111111110000000101010101"},
		{[]byte{15, 240, 255, 1, 85, 128, 1, 2, 3}, []byte{15, 240, 255, 1, 85, 128}, "00001111111100001111111100000001010101011"},
	}
	for idx, table := range tables {
		bt, val, err := DecodeBitString(&tables[idx].src, uint64(len(table.val)))
		if err != nil {
			t.Errorf("combination %v failed. got an error:%v", idx, err)
		}
		// compare length byte
		sameByte := bytes.Compare(table.bt, bt)
		if sameByte != 0 {
			t.Errorf("combination %v failed. Byte get: %v, should:%v", idx, bt, table.bt)
		}
		// compare length value
		sameValue := (table.val == val)
		if !sameValue {
			t.Errorf("combination %v failed. Value get: %s, should:%v", idx, val, table.val)
		}
		// compare remainder bytes of src
		sameReminder := bytes.Compare(tables[idx].src, []byte{1, 2, 3})
		if sameReminder != 0 {
			t.Errorf("combination %v failed. Reminder get: %v, should:[1, 2, 3]", idx, table.src)
		}
	}
}

func TestDecodeDoubleLong(t *testing.T) {
	tables := []struct {
		src []byte
		bt  []byte
		val int32
	}{
		{[]byte{255, 255, 255, 231, 1, 2, 3}, []byte{255, 255, 255, 231}, -25},
		{[]byte{127, 255, 255, 255, 1, 2, 3}, []byte{127, 255, 255, 255}, 2147483647},
		{[]byte{128, 0, 0, 1, 1, 2, 3}, []byte{128, 0, 0, 1}, -2147483647},
	}
	for idx, table := range tables {
		bt, val, err := DecodeDoubleLong(&tables[idx].src)
		if err != nil {
			t.Errorf("combination %v failed. got an error:%v", idx, err)
		}
		// compare length byte
		sameByte := bytes.Compare(table.bt, bt)
		if sameByte != 0 {
			t.Errorf("combination %v failed. Byte get: %v, should:%v", idx, bt, table.bt)
		}
		// compare length value
		sameValue := (table.val == val)
		if !sameValue {
			t.Errorf("combination %v failed. Value get: %v, should:%v", idx, val, table.val)
		}
		// compare remainder bytes of src
		sameReminder := bytes.Compare(tables[idx].src, []byte{1, 2, 3})
		if sameReminder != 0 {
			t.Errorf("combination %v failed. Reminder get: %v, should:[1, 2, 3]", idx, table.src)
		}
	}
}

func TestDecodeDoubleLongUnsigned(t *testing.T) {
	tables := []struct {
		src []byte
		bt  []byte
		val uint32
	}{
		{[]byte{0, 0, 0, 255, 1, 2, 3}, []byte{0, 0, 0, 255}, 255},
		{[]byte{0, 0, 255, 255, 1, 2, 3}, []byte{0, 0, 255, 255}, 65535},
		{[]byte{255, 255, 255, 255, 1, 2, 3}, []byte{255, 255, 255, 255}, 4294967295},
	}
	for idx, table := range tables {
		bt, val, err := DecodeDoubleLongUnsigned(&tables[idx].src)
		if err != nil {
			t.Errorf("combination %v failed. got an error:%v", idx, err)
		}
		// compare length byte
		sameByte := bytes.Compare(table.bt, bt)
		if sameByte != 0 {
			t.Errorf("combination %v failed. Byte get: %v, should:%v", idx, bt, table.bt)
		}
		// compare length value
		sameValue := (table.val == val)
		if !sameValue {
			t.Errorf("combination %v failed. Value get: %v, should:%v", idx, val, table.val)
		}
		// compare remainder bytes of src
		sameReminder := bytes.Compare(tables[idx].src, []byte{1, 2, 3})
		if sameReminder != 0 {
			t.Errorf("combination %v failed. Reminder get: %v, should:[1, 2, 3]", idx, table.src)
		}
	}
}

func TestDecodeOctetString(t *testing.T) {
	tables := []struct {
		src []byte
		lt  uint64
		val string
	}{
		{[]byte{7, 210, 12, 4, 3, 10, 6, 11, 255, 0, 120, 0, 1, 2, 3}, 12, "07D20C04030A060BFF007800"},
		{[]byte{1, 0, 0, 3, 0, 255, 1, 2, 3}, 6, "0100000300FF"},
	}
	for idx, table := range tables {
		answer := table.src[:table.lt]
		bt, val, err := DecodeOctetString(&tables[idx].src, table.lt)
		if err != nil {
			t.Errorf("combination %v failed. got an error:%v", idx, err)
		}
		// compare length byte
		sameByte := bytes.Compare(answer, bt)
		if sameByte != 0 {
			t.Errorf("combination %v failed. Byte get: %v, should:%v", idx, bt, answer)
		}
		// compare length value
		sameValue := (table.val == strings.ToUpper(val))
		if !sameValue {
			t.Errorf("combination %v failed. Value get: %s, should:%v", idx, strings.ToUpper(val), table.val)
		}
		// compare remainder bytes of src
		sameReminder := bytes.Compare(tables[idx].src, []byte{1, 2, 3})
		if sameReminder != 0 {
			t.Errorf("combination %v failed. Reminder get: %v, should:[1, 2, 3]", idx, table.src)
		}
	}
}

func TestDecodeVisibleString(t *testing.T) {
	tables := []struct {
		src []byte
		bt  []byte
		val string
	}{
		{[]byte{65, 66, 67, 68, 1, 2, 3}, []byte{65, 66, 67, 68}, "ABCD"},
		{[]byte{123, 125, 91, 93, 40, 41, 33, 59, 1, 2, 3}, []byte{123, 125, 91, 93, 40, 41, 33, 59}, "{}[]()!;"},
	}
	for idx, table := range tables {
		bt, val, err := DecodeVisibleString(&tables[idx].src, uint64(len(table.val)))
		if err != nil {
			t.Errorf("combination %v failed. got an error:%v", idx, err)
		}
		// compare length byte
		sameByte := bytes.Compare(table.bt, bt)
		if sameByte != 0 {
			t.Errorf("combination %v failed. Byte get: %v, should:%v", idx, bt, table.bt)
		}
		// compare length value
		sameValue := (table.val == val)
		if !sameValue {
			t.Errorf("combination %v failed. Value get: %s, should:%v", idx, val, table.val)
		}
		// compare remainder bytes of src
		sameReminder := bytes.Compare(tables[idx].src, []byte{1, 2, 3})
		if sameReminder != 0 {
			t.Errorf("combination %v failed. Reminder get: %v, should:[1, 2, 3]", idx, table.src)
		}
	}
}

func TestDecodeUTF8String(t *testing.T) {
	tables := []struct {
		src []byte
		bt  []byte
		val string
	}{
		{[]byte{65, 66, 67, 68, 1, 2, 3}, []byte{65, 66, 67, 68}, "ABCD"},
		{[]byte{97, 209, 132, 225, 144, 131, 240, 157, 149, 171, 1, 2, 3}, []byte{97, 209, 132, 225, 144, 131, 240, 157, 149, 171}, "aфᐃ𝕫"},
		{[]byte{230, 136, 145, 230, 132, 155, 228, 189, 160, 1, 2, 3}, []byte{230, 136, 145, 230, 132, 155, 228, 189, 160}, "我愛你"},
	}
	for idx, table := range tables {
		bt, val, err := DecodeUTF8String(&tables[idx].src, uint64(len(table.val)))
		if err != nil {
			t.Errorf("combination %v failed. got an error:%v", idx, err)
		}
		// compare length byte
		sameByte := bytes.Compare(table.bt, bt)
		if sameByte != 0 {
			t.Errorf("combination %v failed. Byte get: %v, should:%v", idx, bt, table.bt)
		}
		// compare length value
		sameValue := (table.val == val)
		if !sameValue {
			t.Errorf("combination %v failed. Value get: %s, should:%v", idx, val, table.val)
		}
		// compare remainder bytes of src
		sameReminder := bytes.Compare(tables[idx].src, []byte{1, 2, 3})
		if sameReminder != 0 {
			t.Errorf("combination %v failed. Reminder get: %v, should:[1, 2, 3]", idx, table.src)
		}
	}
}

func TestDecodeBCD(t *testing.T) {
	tables := []struct {
		src []byte
		bt  []byte
		val int8
	}{
		{[]byte{127, 1, 2, 3}, []byte{127}, 127},
		{[]byte{255, 1, 2, 3}, []byte{255}, -1},
	}
	for idx, table := range tables {
		bt, val, err := DecodeBCD(&tables[idx].src)
		if err != nil {
			t.Errorf("combination %v failed. got an error:%v", idx, err)
		}
		// compare length byte
		sameByte := bytes.Compare(table.bt, bt)
		if sameByte != 0 {
			t.Errorf("combination %v failed. Byte get: %v, should:%v", idx, bt, table.bt)
		}
		// compare length value
		sameValue := (table.val == val)
		if !sameValue {
			t.Errorf("combination %v failed. Value get: %v, should:%v", idx, val, table.val)
		}
		// compare remainder bytes of src
		sameReminder := bytes.Compare(tables[idx].src, []byte{1, 2, 3})
		if sameReminder != 0 {
			t.Errorf("combination %v failed. Reminder get: %v, should:[1, 2, 3]", idx, table.src)
		}
	}
}

// DecodeInteger == DecodeBCD == DecodeEnum

func TestDecodeLong(t *testing.T) {
	tables := []struct {
		src []byte
		bt  []byte
		val int16
	}{
		{[]byte{127, 255, 1, 2, 3}, []byte{127, 255}, 1<<15 - 1},
		{[]byte{128, 0, 1, 2, 3}, []byte{128, 0}, -1 << 15},
	}
	for idx, table := range tables {
		bt, val, err := DecodeLong(&tables[idx].src)
		if err != nil {
			t.Errorf("combination %v failed. got an error:%v", idx, err)
		}
		// compare length byte
		sameByte := bytes.Compare(table.bt, bt)
		if sameByte != 0 {
			t.Errorf("combination %v failed. Byte get: %v, should:%v", idx, bt, table.bt)
		}
		// compare length value
		sameValue := (table.val == val)
		if !sameValue {
			t.Errorf("combination %v failed. Value get: %v, should:%v", idx, val, table.val)
		}
		// compare remainder bytes of src
		sameReminder := bytes.Compare(tables[idx].src, []byte{1, 2, 3})
		if sameReminder != 0 {
			t.Errorf("combination %v failed. Reminder get: %v, should:[1, 2, 3]", idx, table.src)
		}
	}
}

func TestDecodeUnsigned(t *testing.T) {
	tables := []struct {
		src []byte
		bt  []byte
		val uint8
	}{
		{[]byte{255, 1, 2, 3}, []byte{255}, 255},
		{[]byte{0, 1, 2, 3}, []byte{0}, 0},
	}
	for idx, table := range tables {
		bt, val, err := DecodeUnsigned(&tables[idx].src)
		if err != nil {
			t.Errorf("combination %v failed. got an error:%v", idx, err)
		}
		// compare length byte
		sameByte := bytes.Compare(table.bt, bt)
		if sameByte != 0 {
			t.Errorf("combination %v failed. Byte get: %v, should:%v", idx, bt, table.bt)
		}
		// compare length value
		sameValue := (table.val == val)
		if !sameValue {
			t.Errorf("combination %v failed. Value get: %v, should:%v", idx, val, table.val)
		}
		// compare remainder bytes of src
		sameReminder := bytes.Compare(tables[idx].src, []byte{1, 2, 3})
		if sameReminder != 0 {
			t.Errorf("combination %v failed. Reminder get: %v, should:[1, 2, 3]", idx, table.src)
		}
	}
}

func TestDecodeLongUnsigned(t *testing.T) {
	tables := []struct {
		src []byte
		bt  []byte
		val uint16
	}{
		{[]byte{255, 255, 1, 2, 3}, []byte{255, 255}, 1<<16 - 1},
		{[]byte{0, 0, 1, 2, 3}, []byte{0, 0}, 0},
	}
	for idx, table := range tables {
		bt, val, err := DecodeLongUnsigned(&tables[idx].src)
		if err != nil {
			t.Errorf("combination %v failed. got an error:%v", idx, err)
		}
		// compare length byte
		sameByte := bytes.Compare(table.bt, bt)
		if sameByte != 0 {
			t.Errorf("combination %v failed. Byte get: %v, should:%v", idx, bt, table.bt)
		}
		// compare length value
		sameValue := (table.val == val)
		if !sameValue {
			t.Errorf("combination %v failed. Value get: %v, should:%v", idx, val, table.val)
		}
		// compare remainder bytes of src
		sameReminder := bytes.Compare(tables[idx].src, []byte{1, 2, 3})
		if sameReminder != 0 {
			t.Errorf("combination %v failed. Reminder get: %v, should:[1, 2, 3]", idx, table.src)
		}
	}
}

func TestDecodeLong64(t *testing.T) {
	tables := []struct {
		src []byte
		bt  []byte
		val int64
	}{
		{[]byte{127, 255, 255, 255, 255, 255, 255, 255, 1, 2, 3}, []byte{127, 255, 255, 255, 255, 255, 255, 255}, 1<<63 - 1},
		{[]byte{128, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3}, []byte{128, 0, 0, 0, 0, 0, 0, 0}, -1 << 63},
		{[]byte{255, 255, 255, 255, 255, 255, 255, 255, 1, 2, 3}, []byte{255, 255, 255, 255, 255, 255, 255, 255}, -1},
	}
	for idx, table := range tables {
		bt, val, err := DecodeLong64(&tables[idx].src)
		if err != nil {
			t.Errorf("combination %v failed. got an error:%v", idx, err)
		}
		// compare length byte
		sameByte := bytes.Compare(table.bt, bt)
		if sameByte != 0 {
			t.Errorf("combination %v failed. Byte get: %v, should:%v", idx, bt, table.bt)
		}
		// compare length value
		sameValue := (table.val == val)
		if !sameValue {
			t.Errorf("combination %v failed. Value get: %v, should:%v", idx, val, table.val)
		}
		// compare remainder bytes of src
		sameReminder := bytes.Compare(tables[idx].src, []byte{1, 2, 3})
		if sameReminder != 0 {
			t.Errorf("combination %v failed. Reminder get: %v, should:[1, 2, 3]", idx, table.src)
		}
	}
}

func TestDecodeLong64Unsigned(t *testing.T) {
	tables := []struct {
		src []byte
		bt  []byte
		val uint64
	}{
		{[]byte{255, 255, 255, 255, 255, 255, 255, 255, 1, 2, 3}, []byte{255, 255, 255, 255, 255, 255, 255, 255}, 1<<64 - 1},
		{[]byte{0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3}, []byte{0, 0, 0, 0, 0, 0, 0, 0}, 0},
	}
	for idx, table := range tables {
		bt, val, err := DecodeLong64Unsigned(&tables[idx].src)
		if err != nil {
			t.Errorf("combination %v failed. got an error:%v", idx, err)
		}
		// compare length byte
		sameByte := bytes.Compare(table.bt, bt)
		if sameByte != 0 {
			t.Errorf("combination %v failed. Byte get: %v, should:%v", idx, bt, table.bt)
		}
		// compare length value
		sameValue := (table.val == val)
		if !sameValue {
			t.Errorf("combination %v failed. Value get: %v, should:%v", idx, val, table.val)
		}
		// compare remainder bytes of src
		sameReminder := bytes.Compare(tables[idx].src, []byte{1, 2, 3})
		if sameReminder != 0 {
			t.Errorf("combination %v failed. Reminder get: %v, should:[1, 2, 3]", idx, table.src)
		}
	}
}

func TestDecodeFloat32(t *testing.T) {
	tables := []struct {
		src []byte
		bt  []byte
		val float32
	}{
		{[]byte{64, 72, 245, 195, 1, 2, 3}, []byte{64, 72, 245, 195}, 3.14},
		{[]byte{79, 128, 0, 0, 1, 2, 3}, []byte{79, 128, 0, 0}, 4294967295},
		{[]byte{192, 72, 245, 195, 1, 2, 3}, []byte{192, 72, 245, 195}, -3.14},
	}
	for idx, table := range tables {
		bt, val, err := DecodeFloat32(&tables[idx].src)
		if err != nil {
			t.Errorf("combination %v failed. got an error:%v", idx, err)
		}
		// compare length byte
		sameByte := bytes.Compare(table.bt, bt)
		if sameByte != 0 {
			t.Errorf("combination %v failed. Byte get: %v, should:%v", idx, bt, table.bt)
		}
		// compare length value
		sameValue := (table.val == val)
		if !sameValue {
			t.Errorf("combination %v failed. Value get: %v, should:%v", idx, val, table.val)
		}
		// compare remainder bytes of src
		sameReminder := bytes.Compare(tables[idx].src, []byte{1, 2, 3})
		if sameReminder != 0 {
			t.Errorf("combination %v failed. Reminder get: %v, should:[1, 2, 3]", idx, table.src)
		}
	}
}

func TestDecodeFloat64(t *testing.T) {
	tables := []struct {
		src []byte
		bt  []byte
		val float64
	}{
		{[]byte{64, 9, 30, 184, 81, 235, 133, 31, 1, 2, 3}, []byte{64, 9, 30, 184, 81, 235, 133, 31}, 3.14},
		{[]byte{64, 9, 33, 251, 84, 65, 23, 68, 1, 2, 3}, []byte{64, 9, 33, 251, 84, 65, 23, 68}, 3.1415926535},
		{[]byte{65, 239, 255, 255, 255, 224, 0, 0, 1, 2, 3}, []byte{65, 239, 255, 255, 255, 224, 0, 0}, 4294967295},
	}
	for idx, table := range tables {
		bt, val, err := DecodeFloat64(&tables[idx].src)
		if err != nil {
			t.Errorf("combination %v failed. got an error:%v", idx, err)
		}
		// compare length byte
		sameByte := bytes.Compare(table.bt, bt)
		if sameByte != 0 {
			t.Errorf("combination %v failed. Byte get: %v, should:%v", idx, bt, table.bt)
		}
		// compare length value
		sameValue := (table.val == val)
		if !sameValue {
			t.Errorf("combination %v failed. Value get: %v, should:%v", idx, val, table.val)
		}
		// compare remainder bytes of src
		sameReminder := bytes.Compare(tables[idx].src, []byte{1, 2, 3})
		if sameReminder != 0 {
			t.Errorf("combination %v failed. Reminder get: %v, should:[1, 2, 3]", idx, table.src)
		}
	}
}

func TestDecodeDate(t *testing.T) {
	tables := []struct {
		src []byte
		bt  []byte
		val time.Time
	}{
		{[]byte{7, 217, 11, 10, 2, 1, 2, 3}, []byte{7, 217, 11, 10, 2}, time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC)},
		{[]byte{5, 220, 1, 1, 1, 1, 2, 3}, []byte{5, 220, 1, 1, 1}, time.Date(1500, time.January, 1, 0, 0, 0, 0, time.UTC)},
	}
	for idx, table := range tables {
		bt, val, err := DecodeDate(&tables[idx].src)
		if err != nil {
			t.Errorf("combination %v failed. got an error:%v", idx, err)
		}
		// compare byte value
		sameByte := bytes.Compare(table.bt, bt)
		if sameByte != 0 {
			t.Errorf("combination %v failed. Byte get: %v, should:%v", idx, bt, table.bt)
		}
		// compare time value
		if !table.val.Equal(val) {
			t.Errorf("combination %v failed. Value get: %v, should:%v", idx, val, table.val)
		}
		// compare remainder bytes of src
		sameReminder := bytes.Compare(tables[idx].src, []byte{1, 2, 3})
		if sameReminder != 0 {
			t.Errorf("combination %v failed. Reminder get: %v, should:[1, 2, 3]", idx, table.src)
		}
	}
}

func TestDecodeTime(t *testing.T) {
	tables := []struct {
		src []byte
		bt  []byte
		val time.Time
	}{
		{[]byte{10, 0, 0, 255, 1, 2, 3}, []byte{10, 0, 0, 255}, time.Date(0, time.January, 1, 10, 0, 0, 255, time.UTC)},
		{[]byte{23, 59, 59, 255, 1, 2, 3}, []byte{23, 59, 59, 255}, time.Date(0, time.January, 1, 23, 59, 59, 255, time.UTC)},
	}
	for idx, table := range tables {
		bt, val, err := DecodeTime(&tables[idx].src)
		if err != nil {
			t.Errorf("combination %v failed. got an error:%v", idx, err)
		}
		// compare byte value
		sameByte := bytes.Compare(table.bt, bt)
		if sameByte != 0 {
			t.Errorf("combination %v failed. Byte get: %v, should:%v", idx, bt, table.bt)
		}
		// compare time value
		if !table.val.Equal(val) {
			t.Errorf("combination %v failed. Value get: %v, should:%v", idx, val, table.val)
		}
		// compare remainder bytes of src
		sameReminder := bytes.Compare(tables[idx].src, []byte{1, 2, 3})
		if sameReminder != 0 {
			t.Errorf("combination %v failed. Reminder get: %v, should:[1, 2, 3]", idx, table.src)
		}
	}
}

func TestDecodeDateTime(t *testing.T) {
	tests := []struct {
		name              string
		src               string
		timeZoneDeviation TimeZone
		val               time.Time
	}{
		{"Current time", "07D00C1E06173B3BFF000000", TimeZoneStandard, time.Date(2000, time.December, 30, 23, 59, 59, 0, time.UTC)},
		{"Past time", "05DC010101000000FF000000", TimeZoneStandard, time.Date(1500, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{"Local time", "07E40310FF000000FF800000", TimeZoneStandard, time.Date(2020, time.March, 16, 0, 0, 0, 0, time.Local)},
		{"UTC positive", "07D00106050F0030FF003C01", TimeZoneStandard, time.Date(2000, time.January, 6, 15, 0, 48, 0, time.FixedZone("UTC-1", -3600))},
		{"UTC negative", "07D00106050F0030FFFF8801", TimeZoneStandard, time.Date(2000, time.January, 6, 15, 0, 48, 0, time.FixedZone("UTC+2", 7200))},
		{"UTC positive with reversed", "07D00106050F0030FF003C01", TimeZoneReversed, time.Date(2000, time.January, 6, 15, 0, 48, 0, time.FixedZone("UTC+1", 3600))},
		{"UTC negative with reversed", "07D00106050F0030FFFF8801", TimeZoneReversed, time.Date(2000, time.January, 6, 15, 0, 48, 0, time.FixedZone("UTC-2", -7200))},
		{"UTC positive with ignored", "07D00106050F0030FF003C01", TimeZoneIgnored, time.Date(2000, time.January, 6, 15, 0, 48, 0, time.Local)},
		{"UTC negative with ignored", "07D00106050F0030FFFF8801", TimeZoneIgnored, time.Date(2000, time.January, 6, 15, 0, 48, 0, time.Local)},
		{"Empty date", "000000000000000000000000", TimeZoneStandard, time.Time{}},
		{"Invalid date", "00B43A190210380AFF0078FF", TimeZoneStandard, time.Time{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := decodeHexString(tt.src)

			TimeZoneDeviation = tt.timeZoneDeviation

			bt, val, err := DecodeDateTime(&src)
			assert.NoError(t, err)

			TimeZoneDeviation = TimeZoneStandard

			// Compare byte value
			assert.Equal(t, decodeHexString(tt.src), bt)

			// Compare time value
			assert.Equal(t, tt.val, val)
		})
	}
}

func TestDecoder1(t *testing.T) {
	d1 := DlmsData{Tag: TagLongUnsigned, Value: uint16(60226)}
	d2 := DlmsData{Tag: TagDateTime, Value: time.Date(2020, time.March, 16, 0, 0, 0, 0, time.Local)}
	d3 := DlmsData{Tag: TagBitString, Value: "0"}
	d4 := DlmsData{Tag: TagDoubleLongUnsigned, Value: uint32(33426304)}
	d5 := DlmsData{Tag: TagLongUnsigned, Value: uint16(3105)}

	src, _ := hex.DecodeString("0101020512EB421907E40310FF000000FF8000000401000601FE0B80120C21")

	dec := NewDataDecoder(&src)
	t1, err := dec.Decode(&src)
	assert.NoError(t, err)
	assert.Equal(t, TagArray, t1.Tag)

	t2 := t1.Value.([]*DlmsData)[0]
	assert.Equal(t, TagStructure, t2.Tag)

	t3 := t2.Value.([]*DlmsData)
	require.Len(t, t3, 5)
	assert.Equal(t, d1.Value, t3[0].Value)
	assert.Equal(t, d2.Value, t3[1].Value)
	assert.Equal(t, d3.Value, t3[2].Value)
	assert.Equal(t, d4.Value, t3[3].Value)
	assert.Equal(t, d5.Value, t3[4].Value)
}

func TestDecoderSimpleCompactArray(t *testing.T) {
	d1 := DlmsData{Tag: TagDoubleLongUnsigned, Value: uint32(305419896)}
	d2 := DlmsData{Tag: TagDoubleLongUnsigned, Value: uint32(572662306)}
	d3 := DlmsData{Tag: TagDoubleLongUnsigned, Value: uint32(50529027)}
	d4 := DlmsData{Tag: TagDoubleLongUnsigned, Value: uint32(1126253345)}
	d5 := DlmsData{Tag: TagDoubleLongUnsigned, Value: uint32(1431651105)}

	src, _ := hex.DecodeString("1306141234567822222222030303034321432155554321")

	dec := NewDataDecoder(&src)
	t1, err := dec.Decode(&src)
	assert.NoError(t, err)
	assert.Equal(t, TagCompactArray, t1.Tag)

	t2 := t1.Value.([]*DlmsData)
	require.Len(t, t2, 5)

	assert.Equal(t, d1.Value, t2[0].Value)
	assert.Equal(t, d2.Value, t2[1].Value)
	assert.Equal(t, d3.Value, t2[2].Value)
	assert.Equal(t, d4.Value, t2[3].Value)
	assert.Equal(t, d5.Value, t2[4].Value)
}

func TestDecoderComplexCompactArray(t *testing.T) {
	d1 := DlmsData{Tag: TagVisibleString, Value: "hello"}
	d2 := DlmsData{Tag: TagInteger, Value: int8(101)}
	d3 := DlmsData{Tag: TagLong, Value: int16(1234)}
	d4 := DlmsData{Tag: TagLong64, Value: int64(1311768467284833366)}
	d5 := DlmsData{Tag: TagVisibleString, Value: "joan"}
	d6 := DlmsData{Tag: TagInteger, Value: int8(7)}
	d7 := DlmsData{Tag: TagLong, Value: int16(451)}
	d8 := DlmsData{Tag: TagLong64, Value: int64(1234605616436508552)}

	src, _ := hex.DecodeString("1302040A0F1014210568656C6C6F6504D21234567890123456046A6F616E0701C31122334455667788")

	dec := NewDataDecoder(&src)
	t1, err := dec.Decode(&src)
	assert.NoError(t, err)
	assert.Equal(t, TagCompactArray, t1.Tag)

	t2 := t1.Value.([]*DlmsData)[0]
	assert.Equal(t, TagStructure, t2.Tag)

	t3 := t2.Value.([]*DlmsData)
	require.Len(t, t3, 4)
	assert.Equal(t, d1.Value, t3[0].Value)
	assert.Equal(t, d2.Value, t3[1].Value)
	assert.Equal(t, d3.Value, t3[2].Value)
	assert.Equal(t, d4.Value, t3[3].Value)

	t4 := t1.Value.([]*DlmsData)[1]
	assert.Equal(t, TagStructure, t4.Tag)

	t5 := t4.Value.([]*DlmsData)
	require.Len(t, t5, 4)
	assert.Equal(t, d5.Value, t5[0].Value)
	assert.Equal(t, d6.Value, t5[1].Value)
	assert.Equal(t, d7.Value, t5[2].Value)
	assert.Equal(t, d8.Value, t5[3].Value)
}
