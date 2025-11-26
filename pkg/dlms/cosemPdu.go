package dlms

import "fmt"

type CosemTag uint8

const (
	// ---- standardized DLMS APDUs
	TagInitiateRequest          CosemTag = 1
	TagReadRequest              CosemTag = 5
	TagWriteRequest             CosemTag = 6
	TagInitiateResponse         CosemTag = 8
	TagReadResponse             CosemTag = 12
	TagWriteResponse            CosemTag = 13
	TagConfirmedServiceError    CosemTag = 14
	TagDataNotification         CosemTag = 15
	TagUnconfirmedWriteRequest  CosemTag = 22
	TagInformationReportRequest CosemTag = 24
	TagGloInitiateRequest       CosemTag = 33
	TagGloInitiateResponse      CosemTag = 40
	TagGloConfirmedServiceError CosemTag = 46
	TagAARQ                     CosemTag = 96
	TagAARE                     CosemTag = 97
	TagRLRQ                     CosemTag = 98
	TagRLRE                     CosemTag = 99
	// --- APDUs used for data communication services
	TagGetRequest               CosemTag = 192
	TagSetRequest               CosemTag = 193
	TagEventNotificationRequest CosemTag = 194
	TagActionRequest            CosemTag = 195
	TagGetResponse              CosemTag = 196
	TagSetResponse              CosemTag = 197
	TagActionResponse           CosemTag = 199
	// --- global ciphered pdus
	TagGloGetRequest               CosemTag = 200
	TagGloSetRequest               CosemTag = 201
	TagGloEventNotificationRequest CosemTag = 202
	TagGloActionRequest            CosemTag = 203
	TagGloGetResponse              CosemTag = 204
	TagGloSetResponse              CosemTag = 205
	TagGloActionResponse           CosemTag = 207
	// --- dedicated ciphered pdus
	TagDedGetRequest               CosemTag = 208
	TagDedSetRequest               CosemTag = 209
	TagDedEventNotificationRequest CosemTag = 210
	TagDedActionRequest            CosemTag = 211
	TagDedGetResponse              CosemTag = 212
	TagDedSetResponse              CosemTag = 213
	TagDedActionResponse           CosemTag = 215
	TagExceptionResponse           CosemTag = 216
)

func ErrWrongTag(idx int, get byte, correct byte) error {
	return fmt.Errorf("wrong data tag on index %v, expecting %v instead of %v", idx, correct, get)
}

func ErrWrongLength(current int, correct int) error {
	return fmt.Errorf("wrong data length, received %d, expecting %d", current, correct)
}

func ErrWrongSlice(current []byte, correct []byte) error {
	return fmt.Errorf("wrong data, received %v, expecting %v", current, correct)
}

func IsSecuredTag(tag CosemTag) bool {
	switch tag {
	case TagGloGetRequest,
		TagGloSetRequest,
		TagGloEventNotificationRequest,
		TagGloActionRequest,
		TagGloGetResponse,
		TagGloSetResponse,
		TagGloActionResponse,
		TagDedGetRequest,
		TagDedSetRequest,
		TagDedEventNotificationRequest,
		TagDedActionRequest,
		TagDedGetResponse,
		TagDedSetResponse,
		TagDedActionResponse:
		return true
	default:
		return false
	}
}

// Value will return primitive value of the target.
// This is used for comparing with non custom typed object
func (s CosemTag) Value() uint8 {
	return uint8(s)
}

type CosemI interface {
	New() (out CosemPDU, err error)
	Decode() (out CosemPDU, err error)
}

type CosemPDU interface {
	Encode() ([]byte, error)
}

// DecodeCosem is a global function to decode payload based on implemented DLMS/COSEM APDU en/decoder
func DecodeCosem(src *[]byte) (out CosemPDU, err error) {
	if len(*src) == 0 {
		return nil, fmt.Errorf("couldn't decode an empty frame")
	}

	switch (*src)[0] {
	case TagConfirmedServiceError.Value():
		out, err = DecodeConfirmedServiceError(src)
	case TagDataNotification.Value():
		out, err = DecodeDataNotification(src)
	case TagGetRequest.Value():
		var decoder GetRequest
		out, err = decoder.Decode(src)
	case TagSetRequest.Value():
		var decoder SetRequest
		out, err = decoder.Decode(src)
	case TagActionRequest.Value():
		var decoder ActionRequest
		out, err = decoder.Decode(src)
	case TagGetResponse.Value():
		var decoder GetResponse
		out, err = decoder.Decode(src)
	case TagSetResponse.Value():
		var decoder SetResponse
		out, err = decoder.Decode(src)
	case TagActionResponse.Value():
		var decoder ActionResponse
		out, err = decoder.Decode(src)
	case TagEventNotificationRequest.Value():
		out, err = DecodeEventNotificationRequest(src)
	case TagExceptionResponse.Value():
		out, err = DecodeExceptionResponse(src)
	default:
		err = fmt.Errorf("byte idx 0 (%v) is not recognized, or relevant DLMS/COSEM is not yet implemented", (*src)[0])
	}

	return
}
