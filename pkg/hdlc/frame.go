package hdlc

import (
	"encoding/binary"
	"fmt"
)

// minFrameLength is the shortest possible frame.
const minFrameLength = 9

// fcsTable is the ISO/IEC 13239 FCS-16 lookup table (polynomial 0x8408).
//
//nolint:gochecknoglobals
var fcsTable = generateFCSTable()

func generateFCSTable() [256]uint16 {
	var table [256]uint16
	for i := 0; i < 256; i++ {
		crc := uint16(i)
		for j := 0; j < 8; j++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0x8408
			} else {
				crc >>= 1
			}
		}
		table[i] = crc
	}

	return table
}

// FCS returns the ISO/IEC 13239 frame checksum of data.
func FCS(data []byte) uint16 {
	fcs := uint16(0xFFFF)

	for _, b := range data {
		fcs = (fcs >> 8) ^ fcsTable[(fcs^uint16(b))&0xFF]
	}

	return fcs ^ 0xFFFF
}

// BuildFrame builds the frame a client sends to a server.
func BuildFrame(addressing AddressingType, client int, upper int, lower int, control uint8, data []byte) []byte {
	return buildFrame(addressing, client, upper, lower, control, data, false)
}

// BuildSegmentedFrame builds a frame whose format field announces that more
// information frames belong to the same upper-layer PDU.
func BuildSegmentedFrame(addressing AddressingType, client int, upper int, lower int, control uint8, data []byte) []byte {
	return buildFrame(addressing, client, upper, lower, control, data, true)
}

func buildFrame(addressing AddressingType, client int, upper int, lower int, control uint8, data []byte, segmented bool) []byte {
	destLen := int(addressing)
	frame := make([]byte, 0, 4+destLen+1+1+2+len(data)+2+1)

	// Starting flag
	frame = append(frame, startAndEndFlag)

	// Frame format, segmentation and length
	lenAndSeg := frameFormatField
	if segmented {
		lenAndSeg |= 0x0800
	}

	if data != nil {
		lenAndSeg |= destLen + 8 + len(data)
	} else {
		lenAndSeg |= destLen + 6
	}

	frame = append(frame, byte(lenAndSeg>>8))
	frame = append(frame, byte(lenAndSeg))

	// Destination address (server)
	switch addressing {
	case AddressingOneByte:
		frame = append(frame, byte((upper<<1)|0x01))
	case AddressingTwoBytes:
		frame = append(frame, byte(upper<<1))
		frame = append(frame, byte((lower<<1)|0x01))
	case AddressingFourBytes:
		frame = append(frame, byte((upper>>7)<<1))
		frame = append(frame, byte(upper<<1))
		frame = append(frame, byte((lower>>7)<<1))
		frame = append(frame, byte((lower<<1)|0x01))
	}

	// Source address
	frame = append(frame, byte(client<<1)|0x01)

	// Control byte
	frame = append(frame, control)

	// HCS
	checksum := FCS(frame[1:])
	frame = append(frame, byte(checksum))
	frame = append(frame, byte(checksum>>8))

	// Data
	if data != nil {
		frame = append(frame, data...)

		// FCS
		checksum = FCS(frame[1:])
		frame = append(frame, byte(checksum))
		frame = append(frame, byte(checksum>>8))
	}

	// Closing flag
	frame = append(frame, startAndEndFlag)

	return frame
}

// NextFrame returns the first complete frame in buf, together with the bytes that follow it.
func NextFrame(buf []byte) (frame []byte, rest []byte) {
	for {
		// Search for the start flag
		for len(buf) > 0 && buf[0] != startAndEndFlag {
			buf = buf[1:]
		}

		// Check minimum frame length
		if len(buf) < minFrameLength {
			return nil, buf
		}

		// Check if the frame is long enough
		length := int(binary.BigEndian.Uint16(buf[1:]) & 0x07FF)
		if len(buf) < length+2 {
			return nil, buf
		}

		// Check end flag, if it is not found, skip the false start flag
		if buf[length+1] != startAndEndFlag {
			buf = buf[1:]

			continue
		}

		return buf[:length+2], buf[length+2:]
	}
}

// ParseFrame parses a frame received from a server.
func ParseFrame(src []byte) (*ReceivedFrame, error) {
	if len(src) < minFrameLength {
		return nil, fmt.Errorf("frame too short, have %d", len(src))
	}

	if (src[1] & 0xF0) != 0xA0 {
		return nil, fmt.Errorf("invalid frame format, have %02X", src[1])
	}

	isSegmented := (src[1] & 0x08) == 0x08

	// Client address is always 1 byte (destination when server responds)
	clientAddress := int(src[3]) >> 1
	if src[3]&0x01 != 0x01 {
		return nil, fmt.Errorf("invalid client address byte, expected LSB=1, have %02X", src[3])
	}

	// Server address: read bytes until one has LSB=1 (extension bit)
	serverStart := 4
	serverEnd := serverStart
	for serverEnd < len(src) {
		if src[serverEnd]&0x01 == 0x01 {
			serverEnd++
			break
		}
		serverEnd++
	}

	serverLen := serverEnd - serverStart
	if serverLen == 0 || serverLen > 4 {
		return nil, fmt.Errorf("invalid server address length: %d", serverLen)
	}

	// Decode server address fields
	var upperAddress, lowerAddress int

	switch serverLen {
	case 1:
		upperAddress = int(src[serverStart]) >> 1
		lowerAddress = 0
	case 2:
		upperAddress = int(src[serverStart]) >> 1
		lowerAddress = int(src[serverStart+1]) >> 1
	case 4:
		upperAddress = (int(src[serverStart])>>1)<<7 | int(src[serverStart+1])>>1
		lowerAddress = (int(src[serverStart+2])>>1)<<7 | int(src[serverStart+3])>>1
	default:
		return nil, fmt.Errorf("unsupported server address length: %d", serverLen)
	}

	controlIdx := serverEnd
	if len(src) < controlIdx+3 {
		return nil, fmt.Errorf("frame too short for control+HCS, have %d", len(src))
	}

	control := src[controlIdx]
	hcs := binary.LittleEndian.Uint16(src[controlIdx+1:])
	calculatedHCS := FCS(src[1 : controlIdx+1])
	if hcs != calculatedHCS {
		return nil, fmt.Errorf("HCS error, have %04X, expected %04X", hcs, calculatedHCS)
	}

	dataStart := controlIdx + 3
	var data []byte
	var fcs uint16

	if len(src) > dataStart+3 {
		data = src[dataStart : len(src)-3]
		fcs = binary.LittleEndian.Uint16(src[len(src)-3:])
		calculatedFCS := FCS(src[1 : len(src)-3])
		if fcs != calculatedFCS {
			return nil, fmt.Errorf("FCS error, have %04X, expected %04X", fcs, calculatedFCS)
		}
	} else {
		data = nil
		fcs = hcs
	}

	receivedFrame := &ReceivedFrame{
		UpperAddress:  upperAddress,
		LowerAddress:  lowerAddress,
		ClientAddress: clientAddress,
		Control:       control,
		IsSegmented:   isSegmented,
		Data:          data,
		HCS:           hcs,
		FCS:           fcs,
	}

	return receivedFrame, nil
}
