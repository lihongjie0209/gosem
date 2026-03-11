// Copyright (c) 2023 Circutor S.A. All rights reserved.

package axdr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAsnDecode(t *testing.T) {
	time1 := time.Date(2016, time.April, 1, 10, 0, 0, 0, time.UTC)
	dateTime1 := time.Date(2006, time.January, 2, 0, 0, 0, 0, time.UTC)
	timeTime := time.Date(0, time.January, 1, 15, 4, 5, 0, time.UTC)

	tests := []struct {
		name    string
		v       *DlmsData
		want    string
		wantErr bool
	}{
		{
			name:    "null data",
			v:       CreateAxdrNull(),
			want:    "null_data{}",
			wantErr: false,
		},
		{
			name:    "null array",
			v:       CreateAxdrArray([]*DlmsData{}),
			want:    "array{}",
			wantErr: false,
		},
		{
			name:    "simple array",
			v:       CreateAxdrArray(([]*DlmsData{CreateAxdrLongUnsigned(2), CreateAxdrLongUnsigned(4)})),
			want:    "array{long_unsigned{2}long_unsigned{4}}",
			wantErr: false,
		},
		{
			name:    "complex array",
			v:       CreateAxdrArray([]*DlmsData{CreateAxdrLongUnsigned(2), CreateAxdrLongUnsigned(4), CreateAxdrStructure([]*DlmsData{CreateAxdrLongUnsigned(2), CreateAxdrLongUnsigned(4)})}),
			want:    "array{long_unsigned{2}long_unsigned{4}structure{long_unsigned{2}long_unsigned{4}}}",
			wantErr: false,
		},
		{
			name:    "simple structure",
			v:       CreateAxdrStructure([]*DlmsData{CreateAxdrLongUnsigned(8), CreateAxdrOctetString("00 00 01 00 00 ff"), CreateAxdrInteger(2), CreateAxdrLongUnsigned(0)}),
			want:    "structure{long_unsigned{8}octet_string{00 00 01 00 00 ff}integer{2}long_unsigned{0}}",
			wantErr: false,
		},
		{
			name:    "boolean",
			v:       CreateAxdrBoolean(true),
			want:    "boolean{true}",
			wantErr: false,
		},
		{
			name:    "bit_string",
			v:       CreateAxdrBitString("1010000010"),
			want:    "bit_string{1010000010}",
			wantErr: false,
		},
		{
			name:    "double_long",
			v:       CreateAxdrDoubleLong(-123456789),
			want:    "double_long{-123456789}",
			wantErr: false,
		},
		{
			name:    "double_long_unsigned",
			v:       CreateAxdrDoubleLongUnsigned(123456789),
			want:    "double_long_unsigned{123456789}",
			wantErr: false,
		},
		{
			name:    "floating_point",
			v:       CreateAxdrFloatingPoint(4.59),
			want:    "floating_point{4.59}",
			wantErr: false,
		},
		{
			name:    "octet_string",
			v:       CreateAxdrOctetString("test_string"),
			want:    "octet_string{test_string}",
			wantErr: false,
		},
		{
			name:    "octet_string_time",
			v:       CreateAxdrOctetString("07d0020501000d24ff800001"),
			want:    "octet_string{07d0020501000d24ff800001}",
			wantErr: false,
		},
		{
			name:    "visible_string",
			v:       CreateAxdrVisibleString("123"),
			want:    "visible_string{123}",
			wantErr: false,
		},
		{
			name:    "bcd",
			v:       CreateAxdrBCD(25),
			want:    "bcd{25}",
			wantErr: false,
		},
		{
			name:    "integer",
			v:       CreateAxdrInteger(2),
			want:    "integer{2}",
			wantErr: false,
		},
		{
			name:    "long",
			v:       CreateAxdrLong(-34),
			want:    "long{-34}",
			wantErr: false,
		},
		{
			name:    "unsigned",
			v:       CreateAxdrUnsigned(2),
			want:    "unsigned{2}",
			wantErr: false,
		},
		{
			name:    "long_unsigned",
			v:       CreateAxdrLongUnsigned(2),
			want:    "long_unsigned{2}",
			wantErr: false,
		},
		{
			name:    "compact_array_simple",
			v:       CreateAxdrCompactArray([]*DlmsData{CreateAxdrLongUnsigned(2), CreateAxdrLongUnsigned(4), CreateAxdrLongUnsigned(6), CreateAxdrLongUnsigned(8)}),
			want:    "compact_array{long_unsigned{2}long_unsigned{4}long_unsigned{6}long_unsigned{8}}",
			wantErr: false,
		},
		{
			name:    "long64",
			v:       CreateAxdrLong64(-1234567890123456789),
			want:    "long64{-1234567890123456789}",
			wantErr: false,
		},
		{
			name:    "long64_unsigned",
			v:       CreateAxdrLong64Unsigned(1234567890123456789),
			want:    "long64_unsigned{1234567890123456789}",
			wantErr: false,
		},
		{
			name:    "enum",
			v:       CreateAxdrEnum(8),
			want:    "enum{8}",
			wantErr: false,
		},
		{
			name:    "float_32",
			v:       CreateAxdrFloat32(1.25),
			want:    "float_32{1.25}",
			wantErr: false,
		},
		{
			name:    "float_64",
			v:       CreateAxdrFloat64(1.23456789),
			want:    "float_64{1.23456789}",
			wantErr: false,
		},
		{
			name:    "date_time",
			v:       CreateAxdrDateTime(time1),
			want:    "date_time{2016/04/01 10:00:00}",
			wantErr: false,
		},
		{
			name:    "date",
			v:       CreateAxdrDate(dateTime1),
			want:    "date{2006/01/02}",
			wantErr: false,
		},
		{
			name:    "time",
			v:       CreateAxdrTime(timeTime),
			want:    "time{15:04:05}",
			wantErr: false,
		},
		{
			name:    "no exist",
			v:       nil,
			wantErr: true,
		},
		{
			name:    "wrong array",
			v:       CreateAxdrArray(([]*DlmsData{CreateAxdrLongUnsigned(2), nil})),
			wantErr: true,
		},
		{
			name:    "wrong structure",
			v:       CreateAxdrStructure(([]*DlmsData{CreateAxdrLongUnsigned(2), nil})),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AsnDecode(tt.v)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
