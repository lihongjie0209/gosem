package dlms

type ReleaseResponseReason uint8

const (
	ReleaseResponseReasonNormal      ReleaseResponseReason = 0
	ReleaseResponseReasonNotFinished ReleaseResponseReason = 1
	ReleaseResponseReasonUserDefined ReleaseResponseReason = 30
)

type RLRE struct {
	ReleaseResponseReason *ReleaseResponseReason
}

func DecodeRLRE(ori *[]byte) (out RLRE, err error) {
	src := *ori

	if len(src) < 2 {
		err = ErrWrongLength(len(src), 3)
		return
	}

	if src[0] != TagRLRE.Value() {
		err = ErrWrongTag(0, src[0], byte(TagRLRE))
		return
	}

	length := int(2 + src[1])
	if len(src) < length {
		err = ErrWrongLength(len(src), length)
		return
	}

	src = src[2:]
	length -= 2

	for length != 0 {
		if len(src) < 2 {
			err = ErrWrongLength(len(src), 2)
			return
		}

		tagLength := int(src[1])
		if len(src) < (2 + tagLength) {
			err = ErrWrongLength(len(src), 2+tagLength)
			return
		}

		tag := src[0]
		if tag == BERTypeContext {
			// ReleaseRequestReasonNormal - 0x80
			response := ReleaseResponseReason(src[2])
			out.ReleaseResponseReason = &response
		}

		if err != nil {
			return
		}

		src = src[2+tagLength:]
		length -= 2 + tagLength
	}

	(*ori) = (*ori)[len((*ori))-len(src):]
	return
}
