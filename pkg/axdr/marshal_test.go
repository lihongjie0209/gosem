package axdr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMarshalData(t *testing.T) {
	var data int16 = 123

	tests := []struct {
		name    string
		v       interface{}
		want    *DlmsData
		wantErr bool
	}{
		{
			name:    "int8",
			v:       int8(1),
			want:    CreateAxdrInteger(1),
			wantErr: false,
		},
		{
			name:    "int16",
			v:       int16(-23),
			want:    CreateAxdrLong(-23),
			wantErr: false,
		},
		{
			name:    "int32",
			v:       int32(-123456789),
			want:    CreateAxdrDoubleLong(-123456789),
			wantErr: false,
		},
		{
			name:    "int64",
			v:       int64(-1234567890123456789),
			want:    CreateAxdrLong64(-1234567890123456789),
			wantErr: false,
		},
		{
			name:    "uint8",
			v:       uint8(1),
			want:    CreateAxdrUnsigned(1),
			wantErr: false,
		},
		{
			name:    "uint16",
			v:       uint16(23),
			want:    CreateAxdrLongUnsigned(23),
			wantErr: false,
		},
		{
			name:    "uint32",
			v:       uint32(123456789),
			want:    CreateAxdrDoubleLongUnsigned(123456789),
			wantErr: false,
		},
		{
			name:    "uint64",
			v:       uint64(1234567890123456789),
			want:    CreateAxdrLong64Unsigned(1234567890123456789),
			wantErr: false,
		},
		{
			name:    "float32",
			v:       float32(1.23),
			want:    CreateAxdrFloat32(1.23),
			wantErr: false,
		},
		{
			name:    "float64",
			v:       float64(1.23456789),
			want:    CreateAxdrFloat64(1.23456789),
			wantErr: false,
		},
		{
			name:    "string",
			v:       "test",
			want:    CreateAxdrOctetString("test"),
			wantErr: false,
		},
		{
			name:    "bool",
			v:       true,
			want:    CreateAxdrBoolean(true),
			wantErr: false,
		},
		{
			name:    "int16 ptr",
			v:       &data,
			want:    CreateAxdrLong(123),
			wantErr: false,
		},
		{
			name:    "slice",
			v:       []uint16{2, 4},
			want:    CreateAxdrArray([]*DlmsData{CreateAxdrLongUnsigned(2), CreateAxdrLongUnsigned(4)}),
			wantErr: false,
		},
		{
			name: "struct",
			v: struct {
				A int16
				B int32
			}{A: 1, B: 2},
			want:    CreateAxdrStructure([]*DlmsData{CreateAxdrLong(1), CreateAxdrDoubleLong(2)}),
			wantErr: false,
		},
		{
			name: "struct with int16 ptr",
			v: struct {
				A *int16
				B int32
			}{A: &data, B: 2},
			want:    CreateAxdrStructure([]*DlmsData{CreateAxdrLong(123), CreateAxdrDoubleLong(2)}),
			wantErr: false,
		},
		{
			name:    "date time",
			v:       time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC),
			want:    CreateAxdrOctetString(time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)),
			wantErr: false,
		},
		{
			name:    "invalid int",
			v:       int(1),
			wantErr: true,
		},
		{
			name:    "slice with invalid int",
			v:       []int{1, 2},
			wantErr: true,
		},
		{
			name: "struct with invalid int",
			v: struct {
				A int16
				B int
			}{A: 1, B: 2},
			wantErr: true,
		},
		{
			name:    "nil",
			v:       nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarshalData(tt.v)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
