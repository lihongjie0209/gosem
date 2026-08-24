package hdlc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/circutor-library/gosem/pkg/hdlc"
)

func TestFCS(t *testing.T) {
	// The HCS of the SNRM frame 7EA00802219393DBD87E
	assert.Equal(t, uint16(0xD8DB), hdlc.FCS(decodeHexString("A00802219393")))
}

func TestBuildFrame(t *testing.T) {
	// SNRM, two-byte addressing: client 73, server 1:16
	assert.Equal(t, decodeHexString("7EA00802219393DBD87E"),
		hdlc.BuildFrame(hdlc.AddressingTwoBytes, 73, 1, 16, 0x93, nil))

	// SNRM, one-byte addressing: client 73, server 1
	assert.Equal(t, decodeHexString("7EA007039393D1087E"),
		hdlc.BuildFrame(hdlc.AddressingOneByte, 73, 1, 0, 0x93, nil))

	// SNRM, four-byte addressing: client 2, server 1:16
	assert.Equal(t, decodeHexString("7EA00A000200210593F3807E"),
		hdlc.BuildFrame(hdlc.AddressingFourBytes, 2, 1, 16, 0x93, nil))
}

func TestBuildFrameWithData(t *testing.T) {
	// UI frame with a one-byte information field: client 1, server 1:16
	assert.Equal(t, decodeHexString("7EA00B02210313425880C4427E"),
		hdlc.BuildFrame(hdlc.AddressingTwoBytes, 1, 1, 0x10, hdlc.ControlUI, []byte{0x80}))
}

func TestParseFrame(t *testing.T) {
	// UA frame without negotiated parameters
	rf, err := hdlc.ParseFrame(decodeHexString("7EA0089302217320287E"))
	assert.NoError(t, err)
	assert.Equal(t, 73, rf.ClientAddress)
	assert.Equal(t, 1, rf.UpperAddress)
	assert.Equal(t, 16, rf.LowerAddress)
	assert.Equal(t, uint8(0x73), rf.Control)
	assert.False(t, rf.IsSegmented)
	assert.Nil(t, rf.Data)

	// I frame carrying a GET response
	rf, err = hdlc.ParseFrame(decodeHexString("7EA01805022152BE09E6E700C401C10009055630343131F6B67E"))
	assert.NoError(t, err)
	assert.Equal(t, 2, rf.ClientAddress)
	assert.Equal(t, 1, rf.UpperAddress)
	assert.Equal(t, 16, rf.LowerAddress)
	assert.Equal(t, uint8(0x52), rf.Control)
	assert.Equal(t, decodeHexString("E6E700C401C10009055630343131"), rf.Data)
}

func TestParseFrameFails(t *testing.T) {
	tests := map[string]string{
		"too short":            "7EA00893027E",
		"invalid frame format": "7EB0089302217320287E",
		"invalid client byte":  "7EA0089202217320287E",
		"HCS error":            "7EA0089302217320297E",
		"FCS error":            "7EA01805022152BE09E6E700C401C10009055630343131F6B77E",
	}

	for name, frame := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := hdlc.ParseFrame(decodeHexString(frame))
			assert.Error(t, err)
		})
	}
}

func TestNextFrame(t *testing.T) {
	first := "7EA0089302217320287E"
	second := "7EA00B02210313425880C4427E"

	// Two frames received in a single read, preceded by noise
	frame, rest := hdlc.NextFrame(decodeHexString("00FF" + first + second))
	assert.Equal(t, decodeHexString(first), frame)
	assert.Equal(t, decodeHexString(second), rest)

	frame, rest = hdlc.NextFrame(rest)
	assert.Equal(t, decodeHexString(second), frame)
	assert.Empty(t, rest)

	// Nothing to parse yet: the frame is still incomplete
	frame, rest = hdlc.NextFrame(decodeHexString(first[:10]))
	assert.Nil(t, frame)
	assert.Equal(t, decodeHexString(first[:10]), rest)

	// A byte that only looked like a start flag is skipped as soon as the
	// frame it announces turns out not to end with a flag
	frame, _ = hdlc.NextFrame(decodeHexString("7EA004000011" + first))
	assert.Equal(t, decodeHexString(first), frame)

	// Noise without any start flag is dropped
	frame, rest = hdlc.NextFrame(decodeHexString("00FF00FF"))
	assert.Nil(t, frame)
	assert.Empty(t, rest)
}
