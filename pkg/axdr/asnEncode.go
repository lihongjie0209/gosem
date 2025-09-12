// Copyright (c) 2023 Circutor S.A. All rights reserved.

package axdr

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	strNull               = "null_data"
	strArray              = "array"
	strStructure          = "structure"
	strBoolean            = "boolean"
	strBitString          = "bit_string"
	strDoubleLong         = "double_long"
	strDoubleLongUnsigned = "double_long_unsigned"
	strFloatingPoint      = "floating_point"
	strOctetString        = "octet_string"
	strVisibleString      = "visible_string"
	strBCD                = "bcd"
	strInteger            = "integer"
	strLong               = "long"
	strUnsigned           = "unsigned"
	strLongUnsigned       = "long_unsigned"
	strCompactArray       = "compact_array"
	strLong64             = "long64"
	strLong64Unsigned     = "long64_unsigned"
	strEnum               = "enum"
	strFloat32            = "float_32"
	strFloat64            = "float_64"
	strDateTime           = "date_time"
	strDate               = "date"
	strTime               = "time"
	strDontCare           = "dont_care"
	strRaw                = "raw"
)

const (
	openBracket       = '{'
	closeBracket      = '}'
	nonEncodableError = "data is non-encodable: "
	filter            = `^([^\{]+)\{(.*?)\}$`
	dateTimeLayout    = "2006/01/02 15:04:05"
	dateLayout        = "2006/01/02"
	timeLayout        = "15:04:05"
)

func AsnEncode(value string) (data *DlmsData, err error) {
	// Remove all new lines and carriage returns
	re := regexp.MustCompile(`[\r\n]+`)
	value = re.ReplaceAllString(value, "")

	re = regexp.MustCompile(filter)
	valueSplit := re.FindStringSubmatch(value)

	if len(valueSplit) != 3 {
		return nil, fmt.Errorf(nonEncodableError+"%s", value)
	}

	switch valueSplit[1] {
	case strNull:
		data = CreateAxdrNull()
	case strArray:
		axdrArray, err := setArrayStructure(valueSplit[2])
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}
		data = CreateAxdrArray(axdrArray)
	case strStructure:
		axdrStruct, err := setArrayStructure(valueSplit[2])
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}

		data = CreateAxdrStructure(axdrStruct)
	case strBoolean:
		tmp, err := strconv.ParseBool(valueSplit[2])
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}

		data = CreateAxdrBoolean(tmp)
	case strBitString:
		data = CreateAxdrBitString(valueSplit[2])
	case strDoubleLong:
		tmp, err := strconv.ParseInt(valueSplit[2], 10, 32)
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}
		data = CreateAxdrDoubleLong(int32(tmp))
	case strDoubleLongUnsigned:
		tmp, err := strconv.ParseUint(valueSplit[2], 10, 32)
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}
		data = CreateAxdrDoubleLongUnsigned(uint32(tmp))
	case strFloatingPoint:
		tmp, err := strconv.ParseFloat(valueSplit[2], 32)
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}
		data = CreateAxdrFloatingPoint(float32(tmp))
	case strOctetString:
		data = CreateAxdrOctetString(valueSplit[2])
	case strVisibleString:
		data = CreateAxdrVisibleString(valueSplit[2])
	case strBCD:
		tmp, err := strconv.ParseInt(valueSplit[2], 10, 8)
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}
		data = CreateAxdrBCD(int8(tmp))
	case strInteger:
		tmp, err := strconv.ParseInt(valueSplit[2], 10, 8)
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}
		data = CreateAxdrInteger(int8(tmp))
	case strLong:
		tmp, err := strconv.ParseInt(valueSplit[2], 10, 16)
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}
		data = CreateAxdrLong(int16(tmp))
	case strUnsigned:
		tmp, err := strconv.ParseUint(valueSplit[2], 10, 8)
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}
		data = CreateAxdrUnsigned(uint8(tmp))
	case strLongUnsigned:
		tmp, err := strconv.ParseUint(valueSplit[2], 10, 16)
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}
		data = CreateAxdrLongUnsigned(uint16(tmp))
	case strLong64:
		tmp, err := strconv.ParseInt(valueSplit[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}
		data = CreateAxdrLong64(tmp)
	case strLong64Unsigned:
		tmp, err := strconv.ParseUint(valueSplit[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}
		data = CreateAxdrLong64Unsigned(tmp)
	case strEnum:
		tmp, err := strconv.ParseInt(valueSplit[2], 10, 8)
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}
		data = CreateAxdrEnum(uint8(tmp))
	case strFloat32:
		tmp, err := strconv.ParseFloat(valueSplit[2], 32)
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}
		data = CreateAxdrFloat32(float32(tmp))
	case strFloat64:
		tmp, err := strconv.ParseFloat(valueSplit[2], 64)
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}
		data = CreateAxdrFloat64(tmp)
	case strDateTime:
		data = CreateAxdrOctetString(valueSplit[2])
		// data = CreateAxdrDateTime(tmp)
	case strDate:
		tmp, err := time.Parse(dateLayout, valueSplit[2])
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}
		data = CreateAxdrDate(tmp)
	case strTime:
		tmp, err := time.Parse(timeLayout, valueSplit[2])
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}
		data = CreateAxdrTime(tmp)
	case strRaw:
		// Remove all spaces, tabs and new lines
		re := regexp.MustCompile(`\s+`)
		cleaned := re.ReplaceAllString(valueSplit[2], "")

		src, err := hex.DecodeString(cleaned)
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}

		dec := NewDataDecoder(&src)
		t1, err := dec.Decode(&src)
		if err != nil {
			return nil, fmt.Errorf(nonEncodableError+"%w", err)
		}

		data = &t1
	default:
		return nil, fmt.Errorf("unsupported type: %s", valueSplit[2])
	}

	return
}

func getElementsData(data string) []string {
	dataElements := make([]string, 0, 1)
	length := len(data)
	posIni := 0
	lastPost := 0
	chrIni := data[posIni]

	for i, chr := range data {
		if chr == openBracket && (chrIni == 'a' || chrIni == 's') {
			tmpData := data
			for _, v := range dataElements {
				tmpData = strings.TrimPrefix(tmpData, v)
			}
			lastPost = getLastPositionArrayStruct(tmpData) + posIni
			str := data[posIni:lastPost]
			dataElements = append(dataElements, str)
			posIni = lastPost
		} else if chr == closeBracket && i > posIni {
			if chrIni != 'a' && chrIni != 's' {
				lastPost = i + 1
				dataElements = append(dataElements, data[posIni:lastPost])
				posIni = lastPost
			}
		}

		if posIni >= length {
			break
		}
		chrIni = data[posIni]
	}

	return dataElements
}

func getLastPositionArrayStruct(data string) int {
	counter := 0
	position := 0

	for i := 0; i < len(data); i++ {
		if data[i] == openBracket {
			counter++
		} else if data[i] == closeBracket {
			counter--
			if counter <= 0 {
				position = i + 1
				break
			}
		}
	}

	return position
}

func setArrayStructure(data string) (axdrValue []*DlmsData, err error) {
	if data == "" {
		return make([]*DlmsData, 0), nil
	}
	str := getElementsData(data)
	axdrValue = make([]*DlmsData, len(str))

	for i, v := range str {
		axdrValue[i], err = AsnEncode(v)
		if err != nil {
			return nil, fmt.Errorf("value \"%s\": %w", v, err)
		}
	}
	return axdrValue, nil
}
