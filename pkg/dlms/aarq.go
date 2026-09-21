package dlms

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// APDU types
const (
	PduTypeProtocolVersion            = 0
	PduTypeApplicationContextName     = 1
	PduTypeCalledAPTitle              = 2
	PduTypeCalledAEQualifier          = 3
	PduTypeCalledAPInvocationID       = 4
	PduTypeCalledAEInvocationID       = 5
	PduTypeCallingAPTitle             = 6
	PduTypeCallingAEQualifier         = 7
	PduTypeCallingAPInvocationID      = 8
	PduTypeCallingAEInvocationID      = 9
	PduTypeSenderAcseRequirements     = 10
	PduTypeMechanismName              = 11
	PduTypeCallingAuthenticationValue = 12
	PduTypeImplementationInformation  = 29
	PduTypeUserInformation            = 30
)

// BER encoding enumeration values
const (
	BERTypeContext     = 0x80
	BERTypeApplication = 0x40
	BERTypeConstructed = 0x20
)

type ApplicationContext uint8

// Application context definitions
const (
	ApplicationContextLNNoCiphering ApplicationContext = 1
	ApplicationContextSNNoCiphering ApplicationContext = 2
	ApplicationContextLNCiphering   ApplicationContext = 3
	ApplicationContextSNCiphering   ApplicationContext = 4
)

// Conformance block
const (
	ConformanceBlockReservedZero                = 0b100000000000000000000000
	ConformanceBlockGeneralProtection           = 0b010000000000000000000000
	ConformanceBlockGeneralBlockTransfer        = 0b001000000000000000000000
	ConformanceBlockRead                        = 0b000100000000000000000000
	ConformanceBlockWrite                       = 0b000010000000000000000000
	ConformanceBlockUnconfirmedWrite            = 0b000001000000000000000000
	ConformanceBlockReservedSix                 = 0b000000100000000000000000
	ConformanceBlockReservedSeven               = 0b000000010000000000000000
	ConformanceBlockAttribute0SupportedWithSet  = 0b000000001000000000000000
	ConformanceBlockPriorityMgmtSupported       = 0b000000000100000000000000
	ConformanceBlockAttribute0SupportedWithGet  = 0b000000000010000000000000
	ConformanceBlockBlockTransferWithGetOrRead  = 0b000000000001000000000000
	ConformanceBlockBlockTransferWithSetOrWrite = 0b000000000000100000000000
	ConformanceBlockBlockTransferWithAction     = 0b000000000000010000000000
	ConformanceBlockMultipleReferences          = 0b000000000000001000000000
	ConformanceBlockInformationReport           = 0b000000000000000100000000
	ConformanceBlockDataNotification            = 0b000000000000000010000000
	ConformanceBlockAccess                      = 0b000000000000000001000000
	ConformanceBlockParametrizedAccess          = 0b000000000000000000100000
	ConformanceBlockGet                         = 0b000000000000000000010000
	ConformanceBlockSet                         = 0b000000000000000000001000
	ConformanceBlockSelectiveAccess             = 0b000000000000000000000100
	ConformanceBlockEventNotification           = 0b000000000000000000000010
	ConformanceBlockAction                      = 0b000000000000000000000001
)

func EncodeAARQ(settings *Settings) (out []byte, err error) {
	var buf bytes.Buffer

	// Application Association Request
	buf.WriteByte(BERTypeApplication | BERTypeConstructed)

	// APDU length (to be filled in later)
	buf.WriteByte(0x00)

	buf.Write(generateApplicationContextName(settings))

	auth, err := generateAuthentication(settings)
	if err != nil {
		return nil, err
	}
	buf.Write(auth)

	userInfo, err := generateUserInformation(settings)
	if err != nil {
		return nil, err
	}
	buf.Write(userInfo)

	out = buf.Bytes()

	// Add length
	out[1] = byte(len(out) - 2)

	return
}

func generateApplicationContextName(settings *Settings) (out []byte) {
	var buf bytes.Buffer

	// Application context name - 0xA1
	buf.WriteByte(BERTypeContext | BERTypeConstructed | PduTypeApplicationContextName)
	buf.Write([]byte{0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01})
	if settings.Ciphering.Security == SecurityNone && len(settings.Ciphering.SystemTitle) == 0 {
		buf.WriteByte(byte(ApplicationContextLNNoCiphering))
	} else {
		buf.WriteByte(byte(ApplicationContextLNCiphering))
	}

	if len(settings.Ciphering.SystemTitle) > 0 {
		// Add calling-AP-title - 0xA6
		buf.WriteByte(BERTypeContext | BERTypeConstructed | PduTypeCallingAPTitle)
		buf.Write([]byte{0x0A, 0x04, 0x08})
		buf.Write(settings.Ciphering.SystemTitle)
	}

	out = buf.Bytes()

	return
}

func generateAuthentication(settings *Settings) (out []byte, err error) {
	var buf bytes.Buffer

	if settings.Authentication != AuthenticationNone {
		// Add sender ACSE-requirements field component - 0x8A
		buf.WriteByte(BERTypeContext | PduTypeSenderAcseRequirements)
		buf.Write([]byte{0x02, 0x07, 0x80})

		// Add mechanism name - 0x8B
		buf.WriteByte(BERTypeContext | PduTypeMechanismName)
		buf.Write([]byte{0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02})
		buf.WriteByte(byte(settings.Authentication))

		if len(settings.Password) == 0 {
			err = errors.New("password is required for authentication")
		}

		// Add Calling authentication information - 0xAC
		buf.WriteByte(BERTypeContext | BERTypeConstructed | PduTypeCallingAuthenticationValue)
		buf.WriteByte(byte(2 + len(settings.Password)))
		buf.WriteByte(0x80)
		buf.WriteByte(byte(len(settings.Password)))
		buf.Write(settings.Password)
	}

	out = buf.Bytes()

	return
}

func generateUserInformation(settings *Settings) (out []byte, err error) {
	var buf bytes.Buffer

	// User information - 0xBE
	buf.WriteByte(BERTypeContext | BERTypeConstructed | PduTypeUserInformation)
	initiateRequest := getInitiateRequest(settings)

	if settings.Ciphering.Security != SecurityNone {
		if settings.Ciphering.BeforeInvocationCounter != nil {
			if err = settings.Ciphering.BeforeInvocationCounter(SecurityLevelGlobalKey, settings.Ciphering.UnicastKeyIC); err != nil {
				return nil, err
			}
		}
		cfg := Cipher{
			Tag:          TagGloInitiateRequest,
			Security:     settings.Ciphering.Security,
			SystemTitle:  settings.Ciphering.SystemTitle,
			Key:          settings.Ciphering.UnicastKey,
			AuthKey:      settings.Ciphering.AuthenticationKey,
			FrameCounter: settings.Ciphering.UnicastKeyIC,
		}
		settings.Ciphering.UnicastKeyIC++

		initiateRequest, err = CipherData(cfg, initiateRequest)
		if err != nil {
			return
		}
	}

	buf.WriteByte(byte(2 + len(initiateRequest)))
	buf.WriteByte(0x04)
	buf.WriteByte(byte(len(initiateRequest)))
	buf.Write(initiateRequest)

	out = buf.Bytes()

	return
}

func getInitiateRequest(settings *Settings) (out []byte) {
	var buf bytes.Buffer

	// Application Association Request
	buf.WriteByte(byte(TagInitiateRequest))

	if settings.Ciphering.Security == SecurityNone || len(settings.Ciphering.DedicatedKey) == 0 {
		buf.WriteByte(0x00)
	} else {
		buf.WriteByte(0x01)
		buf.WriteByte(byte(len(settings.Ciphering.DedicatedKey)))
		buf.Write(settings.Ciphering.DedicatedKey)
	}

	buf.Write([]byte{0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00})

	bytesConformanceBlock := make([]byte, 4)
	binary.BigEndian.PutUint32(bytesConformanceBlock, uint32(settings.ConformanceBlock))
	buf.Write(bytesConformanceBlock[1:])

	maxPduSize := make([]byte, 2)
	binary.BigEndian.PutUint16(maxPduSize, uint16(settings.MaxPduRecvSize))
	buf.Write(maxPduSize)

	out = buf.Bytes()

	return
}
