// Copyright (c) 2023 Circutor S.A. All rights reserved.

package axdr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAsnEncode(t *testing.T) {
	dateTime1 := time.Date(2006, time.January, 2, 0, 0, 0, 0, time.UTC)
	timeTime := time.Date(0, time.January, 1, 15, 4, 5, 0, time.UTC)

	tests := []struct {
		name    string
		v       string
		want    *DlmsData
		wantErr bool
	}{
		{
			name:    "null data",
			v:       "null_data{}",
			want:    CreateAxdrNull(),
			wantErr: false,
		},
		{
			name:    "null array",
			v:       "array{}",
			want:    CreateAxdrArray([]*DlmsData{}),
			wantErr: false,
		},
		{
			name:    "simple array",
			v:       "array{long_unsigned{2}long_unsigned{4}}",
			want:    CreateAxdrArray([]*DlmsData{CreateAxdrLongUnsigned(2), CreateAxdrLongUnsigned(4)}),
			wantErr: false,
		},
		{
			name:    "complex array",
			v:       "array{long_unsigned{2}long_unsigned{4}structure{long_unsigned{2}long_unsigned{4}}}",
			want:    CreateAxdrArray([]*DlmsData{CreateAxdrLongUnsigned(2), CreateAxdrLongUnsigned(4), CreateAxdrStructure([]*DlmsData{CreateAxdrLongUnsigned(2), CreateAxdrLongUnsigned(4)})}),
			wantErr: false,
		},
		{
			name:    "simple structure",
			v:       "structure{long_unsigned{8}octet_string{00 00 01 00 00 ff}integer{2}long_unsigned{0}}",
			want:    CreateAxdrStructure([]*DlmsData{CreateAxdrLongUnsigned(8), CreateAxdrOctetString("00 00 01 00 00 ff"), CreateAxdrInteger(2), CreateAxdrLongUnsigned(0)}),
			wantErr: false,
		},
		{
			name:    "boolean",
			v:       "boolean{true}",
			want:    CreateAxdrBoolean(true),
			wantErr: false,
		},
		{
			name:    "bit_string",
			v:       "bit_string{1010000010}",
			want:    CreateAxdrBitString("1010000010"),
			wantErr: false,
		},
		{
			name:    "double_long",
			v:       "double_long{-123456789}",
			want:    CreateAxdrDoubleLong(-123456789),
			wantErr: false,
		},
		{
			name:    "double_long_unsigned",
			v:       "double_long_unsigned{123456789}",
			want:    CreateAxdrDoubleLongUnsigned(123456789),
			wantErr: false,
		},
		{
			name:    "floating_point",
			v:       "floating_point{4.59}",
			want:    CreateAxdrFloatingPoint(4.59),
			wantErr: false,
		},
		{
			name:    "octet_string",
			v:       "octet_string{test_string}",
			want:    CreateAxdrOctetString("test_string"),
			wantErr: false,
		},
		{
			name:    "visible_string",
			v:       "visible_string{123}",
			want:    CreateAxdrVisibleString("123"),
			wantErr: false,
		},
		{
			name:    "bcd",
			v:       "bcd{25}",
			want:    CreateAxdrBCD(25),
			wantErr: false,
		},
		{
			name:    "integer",
			v:       "integer{2}",
			want:    CreateAxdrInteger(2),
			wantErr: false,
		},
		{
			name:    "long",
			v:       "long{-34}",
			want:    CreateAxdrLong(-34),
			wantErr: false,
		},
		{
			name:    "unsigned",
			v:       "unsigned{2}",
			want:    CreateAxdrUnsigned(2),
			wantErr: false,
		},
		{
			name:    "long_unsigned",
			v:       "long_unsigned{2}",
			want:    CreateAxdrLongUnsigned(2),
			wantErr: false,
		},
		{
			name:    "long64",
			v:       "long64{-1234567890123456789}",
			want:    CreateAxdrLong64(-1234567890123456789),
			wantErr: false,
		},
		{
			name:    "long64_unsigned",
			v:       "long64_unsigned{1234567890123456789}",
			want:    CreateAxdrLong64Unsigned(1234567890123456789),
			wantErr: false,
		},
		{
			name:    "enum",
			v:       "enum{8}",
			want:    CreateAxdrEnum(8),
			wantErr: false,
		},
		{
			name:    "float_32",
			v:       "float_32{1.25}",
			want:    CreateAxdrFloat32(1.25),
			wantErr: false,
		},
		{
			name:    "float_64",
			v:       "float_64{1.23456789}",
			want:    CreateAxdrFloat64(1.23456789),
			wantErr: false,
		},
		{
			name:    "date_time",
			v:       "date_time{07 E8 01 11 03 0A 00 00 FF 80 00 00}",
			want:    CreateAxdrOctetString("07 E8 01 11 03 0A 00 00 FF 80 00 00"),
			wantErr: false,
		},
		{
			name:    "date",
			v:       "date{2006/01/02}",
			want:    CreateAxdrDate(dateTime1),
			wantErr: false,
		},
		{
			name:    "time",
			v:       "time{15:04:05}",
			want:    CreateAxdrTime(timeTime),
			wantErr: false,
		},
		{
			name:    "complex structure",
			v:       "structure{structure{long_unsigned{8}octet_string{00 00 01 00 00 ff}integer{2}long_unsigned{0}}date_time{07 E8 01 11 03 0A 00 00 FF 80 00 00}date_time{07 E8 01 12 04 0A 00 00 FF 80 00 00}array{}}",
			want:    CreateAxdrStructure([]*DlmsData{CreateAxdrStructure([]*DlmsData{CreateAxdrLongUnsigned(8), CreateAxdrOctetString("00 00 01 00 00 ff"), CreateAxdrInteger(2), CreateAxdrLongUnsigned(0)}), CreateAxdrOctetString("07 E8 01 11 03 0A 00 00 FF 80 00 00"), CreateAxdrOctetString("07 E8 01 12 04 0A 00 00 FF 80 00 00"), CreateAxdrArray([]*DlmsData{})}),
			wantErr: false,
		},
		{
			name:    "wrong array",
			v:       "array{long_unsigned{-4}long_unsigned{4}}",
			wantErr: true,
		},
		{
			name:    "wrong structure",
			v:       "structure{long_unsigned{8}octet_string{00 00 01 00 00 ff}integer{eee}long_unsigned{0}}",
			wantErr: true,
		},
		{
			name:    "wrong boolean",
			v:       "boolean{hello}",
			wantErr: true,
		},
		{
			name:    "double_long",
			v:       "double_long{hello}",
			wantErr: true,
		},
		{
			name:    "double_long_unsigned",
			v:       "double_long_unsigned{-123456789}",
			wantErr: true,
		},
		{
			name:    "floating_point",
			v:       "floating_point{hello}",
			wantErr: true,
		},
		{
			name:    "bcd",
			v:       "bcd{325}",
			wantErr: true,
		},
		{
			name:    "integer",
			v:       "integer{hello}",
			wantErr: true,
		},
		{
			name:    "long",
			v:       "long{hello}",
			wantErr: true,
		},
		{
			name:    "unsigned",
			v:       "unsigned{-2}",
			wantErr: true,
		},
		{
			name:    "long_unsigned",
			v:       "long_unsigned{-2}",
			wantErr: true,
		},
		{
			name:    "long64",
			v:       "long64{hello}",
			wantErr: true,
		},
		{
			name:    "long64_unsigned",
			v:       "long64_unsigned{-1234567890123456789}",
			wantErr: true,
		},
		{
			name:    "enum",
			v:       "enum{a}",
			wantErr: true,
		},
		{
			name:    "float_32",
			v:       "float_32{hello}",
			wantErr: true,
		},
		{
			name:    "float_64",
			v:       "float_64{hello}}",
			wantErr: true,
		},
		{
			name:    "date",
			v:       "date{2006/01/32}",
			wantErr: true,
		},
		{
			name:    "time",
			v:       "time{15:60:05}",
			wantErr: true,
		},
		{
			name:    "no exist",
			v:       "no_exist{}",
			wantErr: true,
		},
		{
			name:    "empty",
			v:       "",
			wantErr: true,
		},
		{
			name:    "wrong",
			v:       "bli blo",
			wantErr: true,
		},
		{
			name:    "wrong structure",
			v:       "structure{long_unsigned{}}",
			wantErr: true,
		},
		{
			name: "raw",
			v: `raw{02 04090C		a31cfc8d8e10d9a78ba9f8470
			    918d 3bf12b805219580d0c5fb8908d65ba40bc9a2dbf2466a670941a95e323c090c16ecd66199dbba69814dee55a01b1afaec319ec5ccb1cf43e8fd6c132954cded79caf2242ef5a93c57d53e51f763b6cad591dba5c6430e2b5f544f090C5e129ea9722e
				8685cd8 	 019f4}`,
			want:    CreateAxdrStructure([]*DlmsData{CreateAxdrOctetString("a31cfc8d8e10d9a78ba9f847"), CreateAxdrOctetString("d3bf12b805219580d0c5fb8908d65ba40bc9a2dbf2466a67"), CreateAxdrOctetString("a95e323c090c16ecd66199dbba69814dee55a01b1afaec319ec5ccb1cf43e8fd6c132954cded79caf2242ef5a93c57d53e51f763b6cad591dba5c6430e2b5f544f"), CreateAxdrOctetString("5e129ea9722e8685cd8019f4")}),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := AsnEncode(tt.v)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
