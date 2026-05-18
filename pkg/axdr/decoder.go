package axdr

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Decoder struct {
	tag dataTag
}

type TimeZone int

const (
	TimeZoneStandard = 0
	TimeZoneReversed = 1
	TimeZoneIgnored  = 2
)

//nolint:gochecknoglobals
var TimeZoneDeviation TimeZone = TimeZoneStandard
var errUnknownTag = errors.New("unknown DLMS tag")

var ErrLengthLess = errors.New("not enough byte length provided")

// Get dataTag equivalent of supplied uint8
func getDataTag(in uint8) (t dataTag, err error) {
	mapToDataTag := map[uint8]dataTag{
		0:   TagNull,
		1:   TagArray,
		2:   TagStructure,
		3:   TagBoolean,
		4:   TagBitString,
		5:   TagDoubleLong,
		6:   TagDoubleLongUnsigned,
		7:   TagFloatingPoint,
		9:   TagOctetString,
		10:  TagVisibleString,
		12:  TagUTF8String,
		13:  TagBCD,
		15:  TagInteger,
		16:  TagLong,
		17:  TagUnsigned,
		18:  TagLongUnsigned,
		19:  TagCompactArray,
		20:  TagLong64,
		21:  TagLong64Unsigned,
		22:  TagEnum,
		23:  TagFloat32,
		24:  TagFloat64,
		25:  TagDateTime,
		26:  TagDate,
		27:  TagTime,
		255: TagDontCare,
	}

	t, ok := mapToDataTag[in]
	if !ok {
		err = fmt.Errorf("unknown dataTag: %d", in)
	}

	return
}

// Create new decode from either supplied byte slice pointer. It will remove first byte from source
func NewDataDecoder(ori *[]byte) *Decoder {
	if len(*ori) < 1 {
		return &Decoder{tag: TagNull}
	}
	tag, err := getDataTag((*ori)[0])
	if err != nil {
		return &Decoder{tag: TagNull}
	}
	(*ori) = (*ori)[1:]
	return &Decoder{tag: tag}
}

// Decode expect byte second after tag byte.
func (dec *Decoder) Decode(ori *[]byte) (r DlmsData, err error) {
	lengthAfterTag := map[dataTag]bool{
		TagNull:               false,
		TagArray:              true,
		TagStructure:          true,
		TagBoolean:            false,
		TagBitString:          true,
		TagDoubleLong:         false,
		TagDoubleLongUnsigned: false,
		TagFloatingPoint:      false,
		TagOctetString:        true,
		TagVisibleString:      true,
		TagUTF8String:         true,
		TagBCD:                false,
		TagInteger:            false,
		TagLong:               false,
		TagUnsigned:           false,
		TagLongUnsigned:       false,
		TagCompactArray:       false,
		TagLong64:             false,
		TagLong64Unsigned:     false,
		TagEnum:               false,
		TagFloat32:            false,
		TagFloat64:            false,
		TagDateTime:           false,
		TagDate:               false,
		TagTime:               false,
		TagDontCare:           false,
	}

	src := *ori

	r.Tag = dec.tag
	haveLength := lengthAfterTag[dec.tag]
	var lengthByte []byte
	var lengthInt uint64
	if haveLength {
		lengthByte, lengthInt, err = DecodeLength(&src)
		if err != nil {
			return
		}
	}

	var rawValue []byte
	var value interface{}
	switch dec.tag {
	case TagNull:
		rawValue = []byte{}
		value = nil
	case TagArray:
		output := make([]*DlmsData, lengthInt)
		// make carbon copy of src to calc rawValue later
		temp := src
		for i := 0; i < int(lengthInt); i++ {
			thisDecoder := NewDataDecoder(&temp)
			thisDlmsData, thisError := thisDecoder.Decode(&temp)
			if thisError != nil {
				err = thisError
				return
			}
			output[i] = &thisDlmsData
		}
		rawValue = src[:len(src)-len(temp)]
		value = output

	case TagStructure:
		// same as array
		output := make([]*DlmsData, lengthInt)
		// make carbon copy of src to calc rawValue later
		temp := src
		for i := 0; i < int(lengthInt); i++ {
			thisDecoder := NewDataDecoder(&temp)
			thisDlmsData, thisError := thisDecoder.Decode(&temp)
			if thisError != nil {
				err = thisError
				return
			}
			output[i] = &thisDlmsData
		}
		rawValue = src[:len(src)-len(temp)]
		value = output

	case TagBoolean:
		rawValue, value, err = DecodeBoolean(&src)
	case TagBitString:
		rawValue, value, err = DecodeBitString(&src, lengthInt)
	case TagDoubleLong:
		rawValue, value, err = DecodeDoubleLong(&src)
	case TagDoubleLongUnsigned:
		rawValue, value, err = DecodeDoubleLongUnsigned(&src)
	case TagFloatingPoint:
		rawValue, value, err = DecodeFloat32(&src)
	case TagOctetString:
		rawValue, value, err = DecodeOctetString(&src, lengthInt)
	case TagVisibleString:
		rawValue, value, err = DecodeVisibleString(&src, lengthInt)
	case TagUTF8String:
		rawValue, value, err = DecodeUTF8String(&src, lengthInt)
	case TagBCD:
		rawValue, value, err = DecodeBCD(&src)
	case TagInteger:
		rawValue, value, err = DecodeInteger(&src)
	case TagLong:
		rawValue, value, err = DecodeLong(&src)
	case TagUnsigned:
		rawValue, value, err = DecodeUnsigned(&src)
	case TagLongUnsigned:
		rawValue, value, err = DecodeLongUnsigned(&src)
	case TagCompactArray:
		rawValue, value, err = DecodeCompactArray(src)
	case TagLong64:
		rawValue, value, err = DecodeLong64(&src)
	case TagLong64Unsigned:
		rawValue, value, err = DecodeLong64Unsigned(&src)
	case TagEnum:
		rawValue, value, err = DecodeEnum(&src)
	case TagFloat32:
		rawValue, value, err = DecodeFloat32(&src)
	case TagFloat64:
		rawValue, value, err = DecodeFloat64(&src)
	case TagDateTime:
		rawValue, value, err = DecodeDateTime(&src)
	case TagDate:
		rawValue, value, err = DecodeDate(&src)
	case TagTime:
		rawValue, value, err = DecodeTime(&src)
	case TagDontCare:
		err = fmt.Errorf("not yet implemented")
	}

	if err != nil {
		return
	}

	r.Value = value

	length := len(rawValue)
	if haveLength {
		length += len(lengthByte)
	}

	// remove bytes from original on success
	(*ori) = (*ori)[length:]

	return
}

func getCompactArrayDecoders(src *[]byte) (outVal []Decoder, err error) {
	if len(*src) < 1 {
		err = ErrLengthLess
		return
	}
	initType := (*src)[0]
	if initType != byte(TagStructure) {
		thisDecoder := NewDataDecoder(src)
		if thisDecoder.tag == TagNull {
			return nil, errUnknownTag
		}

		outVal = []Decoder{*thisDecoder}
		return
	}

	if len(*src) < 2 {
		err = ErrLengthLess
		return
	}
	numberOfStructParams := (*src)[1]
	outVal = make([]Decoder, numberOfStructParams)

	(*src) = (*src)[2:]
	for i := 0; i < int(numberOfStructParams); i++ {
		thisDecoder := NewDataDecoder(src)
		if thisDecoder.tag == TagNull {
			return nil, errUnknownTag
		}

		outVal[i] = *thisDecoder
	}

	return
}

func DecodeLength(src *[]byte) (outByte []byte, outVal uint64, err error) {
	if len(*src) < 1 {
		err = ErrLengthLess
		return
	}
	if (*src)[0] > byte(128) {
		lOfLength := int((*src)[0]) - 128 // L-of-length part
		if len((*src)) < lOfLength+1 {
			err = ErrLengthLess
			return
		}
		realLength := (*src)[1 : lOfLength+1] // real length part

		if len(realLength) > 8 {
			err = fmt.Errorf("length value is bigger than uint64 max value. This Decoder is limited to uint64")
			return
		}

		outByte = (*src)[0 : lOfLength+1] // L-of-length and length

		buf := []byte{0, 0, 0, 0, 0, 0, 0, 0}
		bufStart := 7
		outStart := len(realLength) - 1
		for outStart >= 0 {
			buf[bufStart] = realLength[outStart]
			outStart--
			bufStart--
		}

		outVal = binary.BigEndian.Uint64(buf)
		(*src) = (*src)[1+len(realLength):]
	} else {
		outByte = append(outByte, (*src)[0])
		outVal = uint64((*src)[0])
		(*src) = (*src)[1:]
	}

	return
}

// DecodePartialArray decodes as many complete array elements as possible from a
// potentially truncated byte slice. It is useful when a multi-block transfer has
// been interrupted: instead of failing because the declared element count cannot
// be fulfilled, it returns the elements that were fully received.
//
// data must start with the TagArray byte (0x01). The returned DlmsData has
// Tag = TagArray and Value = []*DlmsData with the successfully decoded elements,
// matching the same shape as a value produced by Decode. If at least one element
// is successfully decoded the returned error is nil; otherwise the decode error
// is returned.
func DecodePartialArray(data []byte) (DlmsData, error) {
	src := make([]byte, len(data))
	copy(src, data)

	if len(src) < 1 {
		return DlmsData{}, ErrLengthLess
	}

	tag, err := getDataTag(src[0])
	if err != nil {
		return DlmsData{}, err
	}

	if tag != TagArray {
		return DlmsData{}, fmt.Errorf("DecodePartialArray: expected TagArray (0x01), got tag %d", src[0])
	}

	src = src[1:] // consume the tag byte

	// Read the declared element count to use as capacity hint and upper bound.
	_, declaredCount, err := DecodeLength(&src)
	if err != nil {
		return DlmsData{}, err
	}

	result := make([]*DlmsData, 0, declaredCount)
	var lastErr error
	for uint64(len(result)) < declaredCount && len(src) > 0 {
		prevLen := len(src)

		temp := src
		dec := NewDataDecoder(&temp)
		d, decErr := dec.Decode(&temp)
		if decErr != nil {
			lastErr = decErr
			break
		}

		// Guard against zero-progress (e.g. unrecognised tag with no advance).
		if len(temp) == prevLen {
			break
		}

		result = append(result, &d)
		src = temp
	}

	if len(result) == 0 {
		if lastErr != nil {
			return DlmsData{}, lastErr
		}
		return DlmsData{}, fmt.Errorf("DecodePartialArray: no elements decoded")
	}

	return DlmsData{Tag: TagArray, Value: result}, nil
}

func DecodeBoolean(src *[]byte) (outByte []byte, outVal bool, err error) {
	if len(*src) < 1 {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:1]
	if outByte[0] == 0x00 {
		outVal = false
	} else {
		outVal = true
	}
	(*src) = (*src)[1:]
	return
}

func DecodeBitString(src *[]byte, length uint64) (outByte []byte, outVal string, err error) {
	byteLength := int(math.Ceil(float64(length) / 8))
	if len(*src) < byteLength {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:byteLength]

	var r strings.Builder
	for _, b := range outByte {
		r.WriteString(fmt.Sprintf("%08b", b))
	}
	outVal = (r.String())[:length]
	(*src) = (*src)[byteLength:]

	return
}

func DecodeDoubleLong(src *[]byte) (outByte []byte, outVal int32, err error) {
	if len(*src) < 4 {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:4]
	outVal |= int32(outByte[0]) << 24
	outVal |= int32(outByte[1]) << 16
	outVal |= int32(outByte[2]) << 8
	outVal |= int32(outByte[3])

	// buf := bytes.NewBuffer(outByte)
	// binary.Read(buf, binary.BigEndian, &outVal)
	(*src) = (*src)[4:]
	return
}

func DecodeDoubleLongUnsigned(src *[]byte) (outByte []byte, outVal uint32, err error) {
	if len(*src) < 4 {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:4]
	outVal |= uint32(outByte[0]) << 24
	outVal |= uint32(outByte[1]) << 16
	outVal |= uint32(outByte[2]) << 8
	outVal |= uint32(outByte[3])
	(*src) = (*src)[4:]
	return
}

func DecodeOctetString(src *[]byte, length uint64) (outByte []byte, outVal string, err error) {
	if uint64(len(*src)) < length {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:length]
	outVal = hex.EncodeToString(outByte)
	(*src) = (*src)[length:]
	return
}

func DecodeVisibleString(src *[]byte, length uint64) (outByte []byte, outVal string, err error) {
	if uint64(len(*src)) < length {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:length]
	outVal = string(outByte)
	(*src) = (*src)[length:]
	return
}

func DecodeUTF8String(src *[]byte, length uint64) (outByte []byte, outVal string, err error) {
	if uint64(len(*src)) < length {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:length]

	var sb strings.Builder
	for sb.Len() < len(outByte) {
		r, _ := utf8.DecodeRune(outByte[sb.Len():])
		if r == utf8.RuneError {
			err = fmt.Errorf("byte slice contain invalid UTF-8 runes")
			return
		}
		sb.WriteRune(r)
	}

	outVal = sb.String()
	(*src) = (*src)[length:]
	return
}

func DecodeBCD(src *[]byte) (outByte []byte, outVal int8, err error) {
	if len(*src) < 1 {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:1]
	outVal = int8(outByte[0])
	(*src) = (*src)[1:]
	return
}

func DecodeInteger(src *[]byte) (outByte []byte, outVal int8, err error) {
	if len(*src) < 1 {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:1]
	outVal = int8(outByte[0])
	(*src) = (*src)[1:]
	return
}

func DecodeLong(src *[]byte) (outByte []byte, outVal int16, err error) {
	if len(*src) < 2 {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:2]
	outVal |= int16(outByte[0]) << 8
	outVal |= int16(outByte[1])
	(*src) = (*src)[2:]
	return
}

func DecodeUnsigned(src *[]byte) (outByte []byte, outVal uint8, err error) {
	if len(*src) < 1 {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:1]
	outVal = outByte[0]
	(*src) = (*src)[1:]
	return
}

func DecodeLongUnsigned(src *[]byte) (outByte []byte, outVal uint16, err error) {
	if len(*src) < 2 {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:2]
	outVal |= uint16(outByte[0]) << 8
	outVal |= uint16(outByte[1])
	(*src) = (*src)[2:]
	return
}

func DecodeCompactArray(src []byte) (outByte []byte, outVal interface{}, err error) {
	// make carbon copy of src to calc rawValue later
	temp := src

	decoders, errDecoders := getCompactArrayDecoders(&temp)
	if errDecoders != nil {
		err = errDecoders

		return
	}

	// After the element types, the next data is the total length of the parameters in bytes.
	_, lengthInt, err := DecodeLength(&temp)
	if err != nil {
		return
	}

	lengthOfHeaders := len(src) - len(temp)
	fullContainer := make([]*DlmsData, 0)
	parsedBytes := uint64(0)

	for parsedBytes < lengthInt {
		var singleStruct DlmsData
		numOfDecoders := len(decoders)

		if numOfDecoders > 1 {
			singleStructContent := make([]*DlmsData, numOfDecoders)
			for j := 0; j < numOfDecoders; j++ {
				thisDlmsData, thisError := decoders[j].Decode(&temp)
				if thisError != nil {
					err = thisError
					return
				}
				singleStructContent[j] = &thisDlmsData
			}
			singleStruct.Tag = TagStructure
			singleStruct.Value = singleStructContent
		} else {
			singleStruct, err = decoders[0].Decode(&temp)
			if err != nil {
				return
			}
		}

		fullContainer = append(fullContainer, &singleStruct)
		parsedBytes = uint64(len(src) - len(temp) - lengthOfHeaders)
	}

	return src[:len(src)-len(temp)], fullContainer, err
}

func DecodeLong64(src *[]byte) (outByte []byte, outVal int64, err error) {
	if len(*src) < 8 {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:8]
	outVal |= int64(outByte[0]) << 56
	outVal |= int64(outByte[1]) << 48
	outVal |= int64(outByte[2]) << 40
	outVal |= int64(outByte[3]) << 32
	outVal |= int64(outByte[4]) << 24
	outVal |= int64(outByte[5]) << 16
	outVal |= int64(outByte[6]) << 8
	outVal |= int64(outByte[7])
	(*src) = (*src)[8:]
	return
}

func DecodeLong64Unsigned(src *[]byte) (outByte []byte, outVal uint64, err error) {
	if len(*src) < 8 {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:8]
	outVal |= uint64(outByte[0]) << 56
	outVal |= uint64(outByte[1]) << 48
	outVal |= uint64(outByte[2]) << 40
	outVal |= uint64(outByte[3]) << 32
	outVal |= uint64(outByte[4]) << 24
	outVal |= uint64(outByte[5]) << 16
	outVal |= uint64(outByte[6]) << 8
	outVal |= uint64(outByte[7])
	(*src) = (*src)[8:]
	return
}

func DecodeEnum(src *[]byte) (outByte []byte, outVal uint8, err error) {
	if len(*src) < 1 {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:1]
	outVal = outByte[0]
	(*src) = (*src)[1:]
	return
}

func DecodeFloat32(src *[]byte) (outByte []byte, outVal float32, err error) {
	if len(*src) < 4 {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:4]
	outVal = math.Float32frombits(binary.BigEndian.Uint32(outByte))
	(*src) = (*src)[4:]
	return
}

func DecodeFloat64(src *[]byte) (outByte []byte, outVal float64, err error) {
	if len(*src) < 8 {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:8]
	outVal = math.Float64frombits(binary.BigEndian.Uint64(outByte))
	(*src) = (*src)[8:]
	return
}

// Decode 5 bytes data into time.Time object
// year highbyte,
// year lowbyte,
// month,
// day of month,
// day of week
func DecodeDate(src *[]byte) (outByte []byte, outVal time.Time, err error) {
	if len(*src) < 5 {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:5]

	year := int(binary.BigEndian.Uint16(outByte[0:2]))
	month := int(outByte[2])
	day := int(outByte[3])
	// weekday := int(outByte[4])

	outVal = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

	(*src) = (*src)[5:]
	return
}

// Decode 4 bytes data into time.Time object
// hour,
// minute,
// second,
// hundredths
func DecodeTime(src *[]byte) (outByte []byte, outVal time.Time, err error) {
	if len(*src) < 4 {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:4]

	hour := int(outByte[0])
	minute := int(outByte[1])
	second := int(outByte[2])
	hundredths := int(outByte[3])

	outVal = time.Date(0, time.Month(1), 1, hour, minute, second, hundredths, time.UTC)

	(*src) = (*src)[4:]
	return
}

// Decode 12 bytes data into time.Time object
// year highbyte,
// year lowbyte,
// month,
// day of month,
// day of week,
// hour,
// minute,
// second,
// hundredths of second,
// deviation highbyte, -- interpreted as long in minutes of local time of UTC
// deviation lowbyte,
// clock status
func DecodeDateTime(src *[]byte) (outByte []byte, outVal time.Time, err error) {
	if len(*src) < 12 {
		err = ErrLengthLess
		return
	}
	outByte = (*src)[:12]

	// if (outByte[11] & 0x01) != 0 {
	// 	err = fmt.Errorf("invalid clock value (%02X)", outByte[11])
	// 	return
	// }

	year := int(binary.BigEndian.Uint16(outByte[0:2]))
	month := int(outByte[2])
	day := int(outByte[3])
	// weekday := int(outByte[4])
	hour := int(outByte[5])
	minute := int(outByte[6])
	second := int(outByte[7])
	hundredths := int(outByte[8])
	if hundredths == 0xFF {
		hundredths = 0
	}

	deviation := binary.BigEndian.Uint16(outByte[9:11])
	location := time.UTC
	if deviation == 0x8000 || TimeZoneDeviation == TimeZoneIgnored {
		location = time.Local
	} else if deviation != 0 {
		d := int(int16(deviation))

		if TimeZoneDeviation == TimeZoneStandard {
			d = -d
		}

		utc := "UTC"
		if d > 0 {
			utc += "+" + strconv.Itoa(d/60)
		} else if d < 0 {
			utc += "-" + strconv.Itoa(-d/60)
		}

		location = time.FixedZone(utc, d*60)
	}

	str := fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d.%02dZ", year, month, day, hour, minute, second, hundredths)
	if _, err := time.Parse(time.RFC3339Nano, str); err != nil {
		outVal = time.Time{}
	} else {
		outVal = time.Date(year, time.Month(month), day, hour, minute, second, hundredths*10000000, location)
	}

	(*src) = (*src)[12:]
	return
}
