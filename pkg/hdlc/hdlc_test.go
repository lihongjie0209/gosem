package hdlc_test

import (
	"encoding/binary"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/circutor-library/gosem/pkg/dlms"
	"gitlab.com/circutor-library/gosem/pkg/dlms/mocks"
	"gitlab.com/circutor-library/gosem/pkg/hdlc"
)

const (
	replyTimeout      = 100 * time.Millisecond
	interOctetTimeout = 20 * time.Millisecond
)

func TestHDLC_Connect(t *testing.T) {
	transportMock := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 10)
	transportMock.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	w := hdlc.New(transportMock, replyTimeout, 3, interOctetTimeout, 16, 73, 1, hdlc.AddressingTwoBytes)

	transportMock.On("Connect").Return(nil).Once()
	sendReceive(transportMock, rdc, "7EA00802219393DBD87E", "7EA01F93022173BCAC8180120501F80601F00704000000010804000000013D9B7E")
	assert.NoError(t, w.Connect())

	transportMock.On("IsConnected").Return(true).Once()
	assert.True(t, w.IsConnected())

	transportMock.On("IsConnected").Return(true).Once()
	transportMock.On("Disconnect").Return(nil).Once()
	sendReceive(transportMock, rdc, "7EA00802219353D71E7E", "7EA01F93022173BCAC8180120501F80601F00704000000010804000000013D9B7E")
	assert.NoError(t, w.Disconnect())

	transportMock.On("Close").Return(nil).Once()
	w.Close()

	transportMock.AssertExpectations(t)
}

func TestHDLC_ConnectWithoutNegotiation(t *testing.T) {
	transportMock := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 10)
	transportMock.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	w := hdlc.New(transportMock, replyTimeout, 3, interOctetTimeout, 16, 73, 1, hdlc.AddressingTwoBytes)

	transportMock.On("Connect").Return(nil).Once()
	sendReceive(transportMock, rdc, "7EA00802219393DBD87E", "7EA0089302217320287E")
	assert.NoError(t, w.Connect())

	transportMock.On("IsConnected").Return(true).Once()
	assert.True(t, w.IsConnected())

	transportMock.On("Close").Return(nil).Once()
	w.Close()

	transportMock.AssertExpectations(t)
}

func TestHDLC_ConnectWithRetry(t *testing.T) {
	transportMock := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 10)
	transportMock.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	w := hdlc.New(transportMock, replyTimeout, 3, interOctetTimeout, 16, 73, 1, hdlc.AddressingTwoBytes)

	transportMock.On("Connect").Return(nil).Once()
	sendWithoutReceive(transportMock, "7EA00802219393DBD87E")
	transportMock.On("IsConnected").Return(true).Once()
	sendReceive(transportMock, rdc, "7EA00802219393DBD87E", "7EA0089302217320287E")
	assert.NoError(t, w.Connect())

	transportMock.On("Close").Return(nil).Once()
	w.Close()

	transportMock.AssertExpectations(t)
}

func TestHDLC_ConnectFail(t *testing.T) {
	transportMock := mocks.NewTransportMock(t)

	transportMock.On("SetReception", mock.Anything).Once()
	w := hdlc.New(transportMock, replyTimeout, 3, interOctetTimeout, 16, 73, 1, hdlc.AddressingTwoBytes)

	transportMock.On("Connect").Return(assert.AnError).Once()
	assert.Error(t, w.Connect())

	transportMock.On("IsConnected").Return(false).Once()
	assert.False(t, w.IsConnected())

	transportMock.On("Close").Return(nil).Once()
	w.Close()

	transportMock.AssertExpectations(t)
}

func TestHDLC_SendAndReceive(t *testing.T) {
	transportMock := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 10)
	hdc := make(dlms.DataChannel, 10)

	transportMock.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	w := hdlc.New(transportMock, replyTimeout, 3, interOctetTimeout, 16, 2, 1, hdlc.AddressingTwoBytes)
	w.SetReception(hdc)

	transportMock.On("Connect").Return(nil).Once()
	sendReceive(transportMock, rdc, "7EA0080221059356957E", "7EA01F05022173E9098180120501F80601F00704000000010804000000013D9B7E")
	assert.NoError(t, w.Connect())

	transportMock.On("IsConnected").Return(true).Once()
	sendReceive(transportMock, rdc, "7EA04502210510939EE6E6006036A1090607608574050801018A0207808B0760857405080201AC0A80083030303030303031BE10040E01000000065F1F040000181F02003F537E", "7EA038050221303B29E6E7006129A109060760857405080101A203020100A305A103020100BE10040E0800065F1F040000101400F00007FE307E")
	assert.NoError(t, w.Send(decodeHexString("6036A1090607608574050801018A0207808B0760857405080201AC0A80083030303030303031BE10040E01000000065F1F040000181F0200")))
	assert.Equal(t, decodeHexString("6129A109060760857405080101A203020100A305A103020100BE10040E0800065F1F040000101400F00007"), <-hdc)

	transportMock.On("IsConnected").Return(true).Once()
	sendReceive(transportMock, rdc, "7EA01A022105321D83E6E600C001C100010100000200FF02004BBB7E", "7EA01805022152BE09E6E700C401C10009055630343131F6B67E")
	assert.NoError(t, w.Send(decodeHexString("C001C100010100000200FF0200")))
	assert.Equal(t, decodeHexString("C401C10009055630343131"), <-hdc)

	transportMock.On("Close").Return(nil).Once()
	w.Close()

	transportMock.AssertExpectations(t)
}

func TestHDLC_ReassemblesSegmentedResponse(t *testing.T) {
	transportMock := mocks.NewTransportMock(t)
	rdc := make(dlms.DataChannel, 10)
	hdc := make(dlms.DataChannel, 1)
	transportMock.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	w := hdlc.New(transportMock, replyTimeout, 1, interOctetTimeout, 0, 2, 1, hdlc.AddressingOneByte)
	w.SetReception(hdc)
	transportMock.On("Connect").Return(nil).Once()
	sendReceiveBytes(transportMock, rdc,
		hdlc.BuildFrame(hdlc.AddressingOneByte, 2, 1, 0, hdlc.ControlSNRM, nil),
		buildServerFrame(2, 1, hdlc.ControlUA, false, nil))
	assert.NoError(t, w.Connect())

	payload := make([]byte, 180)
	for index := range payload {
		payload[index] = byte(index)
	}
	first := append([]byte{0xE6, 0xE7, 0x00}, payload[:60]...)
	second := payload[60:]
	transportMock.On("IsConnected").Return(true).Once()
	sendReceiveBytes(transportMock, rdc,
		hdlc.BuildFrame(hdlc.AddressingOneByte, 2, 1, 0, 0x10, []byte{0xE6, 0xE6, 0x00, 0xC0}),
		buildServerFrame(2, 1, 0x30, true, first))
	sendReceiveBytes(transportMock, rdc,
		hdlc.BuildFrame(hdlc.AddressingOneByte, 2, 1, 0, 0x31, nil),
		buildServerFrame(2, 1, 0x32, false, second))
	assert.NoError(t, w.Send([]byte{0xC0}))
	assert.Equal(t, payload, <-hdc)

	transportMock.On("Close").Return(nil).Once()
	w.Close()
	transportMock.AssertExpectations(t)
}

func TestHDLC_SegmentsOversizedRequest(t *testing.T) {
	transportMock := mocks.NewTransportMock(t)
	rdc := make(dlms.DataChannel, 10)
	hdc := make(dlms.DataChannel, 1)
	transportMock.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	w := hdlc.New(transportMock, replyTimeout, 1, interOctetTimeout, 0, 2, 1, hdlc.AddressingOneByte)
	w.SetReception(hdc)
	transportMock.On("Connect").Return(nil).Once()
	ua := []byte{0x81, 0x80, 0x03, 0x06, 0x01, 0x20}
	sendReceiveBytes(transportMock, rdc,
		hdlc.BuildFrame(hdlc.AddressingOneByte, 2, 1, 0, hdlc.ControlSNRM, nil),
		buildServerFrame(2, 1, hdlc.ControlUA, false, ua))
	assert.NoError(t, w.Connect())

	payload := make([]byte, 50)
	for index := range payload {
		payload[index] = byte(index)
	}
	withLLC := append([]byte{0xE6, 0xE6, 0x00}, payload...)
	transportMock.On("IsConnected").Return(true).Twice()
	sendReceiveBytes(transportMock, rdc,
		hdlc.BuildSegmentedFrame(hdlc.AddressingOneByte, 2, 1, 0, 0x10, withLLC[:32]),
		buildServerFrame(2, 1, 0x31, false, nil))
	sendReceiveBytes(transportMock, rdc,
		hdlc.BuildFrame(hdlc.AddressingOneByte, 2, 1, 0, 0x12, withLLC[32:]),
		buildServerFrame(2, 1, 0x50, false, []byte{0xE6, 0xE7, 0x00, 0xC5, 0x01, 0xC1, 0x00}))
	assert.NoError(t, w.Send(payload))
	assert.Equal(t, []byte{0xC5, 0x01, 0xC1, 0x00}, <-hdc)

	transportMock.On("Close").Return(nil).Once()
	w.Close()
	transportMock.AssertExpectations(t)
}

func TestHDLC_OneByte_Connect(t *testing.T) {
	transportMock := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 10)
	transportMock.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	// lowerAddress=0 (unused in 1-byte mode), clientAddress=73, upperAddress=1
	w := hdlc.New(transportMock, replyTimeout, 3, interOctetTimeout, 0, 73, 1, hdlc.AddressingOneByte)

	transportMock.On("Connect").Return(nil).Once()
	sendReceive(transportMock, rdc, "7ea007039393d1087e", "7ea007930373fb7f7e")
	assert.NoError(t, w.Connect())

	transportMock.On("IsConnected").Return(true).Once()
	assert.True(t, w.IsConnected())

	transportMock.On("IsConnected").Return(true).Once()
	transportMock.On("Disconnect").Return(nil).Once()
	sendReceive(transportMock, rdc, "7ea007039353ddce7e", "7ea007930373fb7f7e")
	assert.NoError(t, w.Disconnect())

	transportMock.On("Close").Return(nil).Once()
	w.Close()

	transportMock.AssertExpectations(t)
}

func TestHDLC_FourBytes_Connect(t *testing.T) {
	transportMock := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 10)
	transportMock.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	// lowerAddress=16, clientAddress=2, upperAddress=1
	w := hdlc.New(transportMock, replyTimeout, 3, interOctetTimeout, 16, 2, 1, hdlc.AddressingFourBytes)

	transportMock.On("Connect").Return(nil).Once()
	sendReceive(transportMock, rdc, "7ea00a000200210593f3807e", "7ea00a05000200217330417e")
	assert.NoError(t, w.Connect())

	transportMock.On("IsConnected").Return(true).Once()
	assert.True(t, w.IsConnected())

	transportMock.On("IsConnected").Return(true).Once()
	transportMock.On("Disconnect").Return(nil).Once()
	sendReceive(transportMock, rdc, "7ea00a000200210553ff467e", "7ea00a05000200217330417e")
	assert.NoError(t, w.Disconnect())

	transportMock.On("Close").Return(nil).Once()
	w.Close()

	transportMock.AssertExpectations(t)
}

func sendReceive(tm *mocks.TransportMock, rdc dlms.DataChannel, in string, out string) {
	tm.On("Send", decodeHexString(in)).Run(func(_ mock.Arguments) {
		if rdc != nil {
			rdc <- decodeHexString(out)
		}
	}).Return(nil).Once()
}

func sendReceiveBytes(tm *mocks.TransportMock, rdc dlms.DataChannel, in, out []byte) {
	tm.On("Send", in).Run(func(_ mock.Arguments) {
		rdc <- out
	}).Return(nil).Once()
}

func buildServerFrame(client, server int, control uint8, segmented bool, data []byte) []byte {
	length := 7
	if data != nil {
		length += len(data) + 2
	}
	format := 0xA000 | length
	if segmented {
		format |= 0x0800
	}
	frame := []byte{0x7E, byte(format >> 8), byte(format), byte(client<<1) | 1, byte(server<<1) | 1, control}
	checksum := hdlc.FCS(frame[1:])
	frame = binary.LittleEndian.AppendUint16(frame, checksum)
	if data != nil {
		frame = append(frame, data...)
		checksum = hdlc.FCS(frame[1:])
		frame = binary.LittleEndian.AppendUint16(frame, checksum)
	}
	return append(frame, 0x7E)
}

func sendWithoutReceive(tm *mocks.TransportMock, in string) {
	tm.On("Send", decodeHexString(in)).Return(nil).Once()
}

func decodeHexString(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
