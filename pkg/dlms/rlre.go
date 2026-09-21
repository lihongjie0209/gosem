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
		return out, ErrWrongLength(len(src), 2)
	}
	if src[0] != TagRLRE.Value() {
		return out, ErrWrongTag(0, src[0], byte(TagRLRE))
	}

	declaredEnd := 2 + int(src[1])
	if declaredEnd > len(src) {
		return out, ErrWrongLength(len(src), declaredEnd)
	}
	position := 2
	for position < len(src) {
		if len(src)-position < 2 {
			return out, ErrWrongLength(len(src)-position, 2)
		}
		tag := src[position]
		tagLength := int(src[position+1])
		contentStart := position + 2

		switch tag {
		case BERTypeContext:
			if tagLength != 1 || contentStart+tagLength > len(src) {
				return out, ErrWrongLength(len(src)-contentStart, tagLength)
			}
			response := ReleaseResponseReason(src[contentStart])
			out.ReleaseResponseReason = &response
			position = contentStart + tagLength

		case BERTypeContext | BERTypeConstructed | PduTypeUserInformation:
			// User information is the final RLRE field. GuruxDLMS.c emits a
			// secured variant whose outer length is four bytes short and whose
			// BE length is one byte short. Accept only that exact deviation when
			// the nested octet string still proves the real boundary.
			actualLength := len(src) - contentStart
			guruxVariant := declaredEnd+4 == len(src) && tagLength+1 == actualLength
			if !guruxVariant && (declaredEnd != len(src) || tagLength != actualLength) {
				return out, ErrWrongLength(actualLength, tagLength)
			}
			if actualLength < 2 {
				return out, ErrWrongLength(actualLength, 2)
			}
			if src[contentStart] != BERTypeOctetString {
				return out, ErrWrongTag(contentStart, src[contentStart], BERTypeOctetString)
			}
			nestedLength := int(src[contentStart+1])
			if nestedLength+2 != actualLength {
				return out, ErrWrongLength(actualLength-2, nestedLength)
			}
			position = len(src)

		default:
			return out, ErrWrongTag(position, tag, BERTypeContext)
		}
	}

	if declaredEnd != len(src) && declaredEnd+4 != len(src) {
		return out, ErrWrongLength(len(src), declaredEnd)
	}
	*ori = (*ori)[len(src):]
	return out, nil
}
