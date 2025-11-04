package dlms

import (
	"bytes"
	"errors"
)

type AssociationResult uint8

const (
	AssociationResultAccepted          AssociationResult = 0
	AssociationResultPermanentRejected AssociationResult = 1
	AssociationResultTransientRejected AssociationResult = 2
)

type SourceDiagnostic uint8

const (
	SourceDiagnosticNone                                       SourceDiagnostic = 0
	SourceDiagnosticNoReasonGiven                              SourceDiagnostic = 1
	SourceDiagnosticApplicationContextNameNotSupported         SourceDiagnostic = 2
	SourceDiagnosticCallingAPTitleNotRecognized                SourceDiagnostic = 3
	SourceDiagnosticCallingAPInvocationIdentifierNotRecognized SourceDiagnostic = 4
	SourceDiagnosticCallingAEQualifierNotRecognized            SourceDiagnostic = 5
	SourceDiagnosticCallingAEInvocationIdentifierNotRecognized SourceDiagnostic = 6
	SourceDiagnosticCalledAPTitleNotRecognized                 SourceDiagnostic = 7
	SourceDiagnosticCalledAPInvocationIdentifierNotRecognized  SourceDiagnostic = 8
	SourceDiagnosticCalledAEQualifierNotRecognized             SourceDiagnostic = 9
	SourceDiagnosticCalledAEInvocationIdentifierNotRecognized  SourceDiagnostic = 10
	SourceDiagnosticAuthenticationMechanismNameNotRecognized   SourceDiagnostic = 11
	SourceDiagnosticAuthenticationMechanismNameRequired        SourceDiagnostic = 12
	SourceDiagnosticAuthenticationFailure                      SourceDiagnostic = 13
	SourceDiagnosticAuthenticationRequired                     SourceDiagnostic = 14
)

type AARE struct {
	ApplicationContext    ApplicationContext
	AssociationResult     AssociationResult
	SourceDiagnostic      SourceDiagnostic
	SourceSystemTitle     []byte
	InitiateResponse      *InitiateResponse
	ConfirmedServiceError *ConfirmedServiceError
	ReceivedIC            *uint32
}

func DecodeAARE(settings *Settings, ori *[]byte) (out AARE, err error) {
	src := *ori

	if len(src) < 2 {
		err = ErrWrongLength(len(src), 3)
		return
	}

	if src[0] != TagAARE.Value() {
		err = ErrWrongTag(0, src[0], byte(TagAARE))
		return
	}

	length := int(2 + src[1])
	if len(src) < length {
		err = ErrWrongLength(len(src), length)
		return
	}

	src = src[2:]
	length -= 2

	for {
		if length == 0 {
			break
		}

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
		switch tag {
		case BERTypeContext | BERTypeConstructed | PduTypeApplicationContextName:
			// Application context name - 0xA1
			out.ApplicationContext, err = parseApplicationContextName(tagLength, src)
		case BERTypeContext | BERTypeConstructed | PduTypeCalledAPTitle:
			// Association result - 0xA2
			out.AssociationResult, err = parseAssociationResult(tagLength, src)
		case BERTypeContext | BERTypeConstructed | PduTypeCalledAEQualifier:
			// Associate source diagnostic - 0xA3
			out.SourceDiagnostic, err = parseAssociateSourceDiagnostic(tagLength, src)
		case BERTypeContext | BERTypeConstructed | PduTypeCalledAPInvocationID:
			// AP title - 0xA4
			out.SourceSystemTitle, err = parseAPTitle(tagLength, src)
			if settings != nil {
				settings.Ciphering.SourceSystemTitle = out.SourceSystemTitle
			}
		case BERTypeContext | BERTypeConstructed | PduTypeUserInformation:
			// User information - 0xBE
			out.InitiateResponse, out.ConfirmedServiceError, out.ReceivedIC, err = parseUserInformation(settings, tagLength, src)
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

func parseApplicationContextName(tagLength int, src []byte) (out ApplicationContext, err error) {
	if tagLength != 9 {
		err = ErrWrongLength(tagLength, 9)
		return
	}
	rsp := []byte{0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01}
	if !bytes.Equal(src[2:10], rsp) {
		err = ErrWrongSlice(src[2:10], rsp)
		return
	}
	out = ApplicationContext(src[10])
	return
}

func parseAssociationResult(tagLength int, src []byte) (out AssociationResult, err error) {
	if tagLength != 3 {
		err = ErrWrongLength(tagLength, 3)
		return
	}
	rsp := []byte{0x02, 0x01}
	if !bytes.Equal(src[2:4], rsp) {
		err = ErrWrongSlice(src[2:4], rsp)
		return
	}
	out = AssociationResult(src[4])
	return
}

func parseAssociateSourceDiagnostic(tagLength int, src []byte) (out SourceDiagnostic, err error) {
	if tagLength != 5 {
		err = ErrWrongLength(tagLength, 5)
		return
	}
	rsp := []byte{0x03, 0x02, 0x01}
	if !bytes.Equal(src[3:6], rsp) {
		err = ErrWrongSlice(src[3:6], rsp)
		return
	}
	out = SourceDiagnostic(src[6])
	return
}

func parseAPTitle(tagLength int, src []byte) (out []byte, err error) {
	if tagLength != 10 {
		err = ErrWrongLength(tagLength, 10)
		return
	}
	rsp := []byte{0x04, 0x08}
	if !bytes.Equal(src[2:4], rsp) {
		err = ErrWrongSlice(src[2:4], rsp)
		return
	}
	out = make([]byte, 8)
	copy(out, src[4:12])
	return
}

func parseUserInformation(settings *Settings, tagLength int, src []byte) (ir *InitiateResponse, cse *ConfirmedServiceError, ric *uint32, err error) {
	if tagLength < 6 {
		err = ErrWrongLength(tagLength, 10)
		return
	}
	if src[2] != 0x04 || src[3] != byte(tagLength-2) {
		err = errors.New("user information length error")
		return
	}
	src = src[4:]

	if (src[0] == TagGloInitiateResponse.Value() || src[0] == TagGloConfirmedServiceError.Value()) && settings != nil {
		cfg := Cipher{
			Tag:         CosemTag(src[0]),
			Security:    settings.Ciphering.Security,
			SystemTitle: settings.Ciphering.SourceSystemTitle,
			Key:         settings.Ciphering.UnicastKey,
			AuthKey:     settings.Ciphering.AuthenticationKey,
		}

		src, err = DecipherData(&cfg, src)
		if err != nil {
			// As a tricky result, if decipher fails, we return a ConfirmedServiceError with other (10) and value 99
			cse := ConfirmedServiceError{
				ConfirmedServiceError: TagErrInitiateError,
				ServiceError:          TagErrOtherError,
				Value:                 TagOtherDecipheringError,
			}

			return nil, &cse, nil, nil
		}

		ric = &cfg.FrameCounter
	}

	if src[0] == TagInitiateResponse.Value() {
		ir, err := DecodeInitiateResponse(&src)
		return &ir, nil, ric, err
	}

	if src[0] == TagConfirmedServiceError.Value() {
		cse, err := DecodeConfirmedServiceError(&src)
		return nil, &cse, ric, err
	}

	err = errors.New("unexpected user information tag")

	return
}
