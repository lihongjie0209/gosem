package hdlc_test

import (
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

	w := hdlc.New(transportMock, replyTimeout, 3, interOctetTimeout, 16, 73, 1)

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

func TestHDLC_ConnectFail(t *testing.T) {
	transportMock := mocks.NewTransportMock(t)

	transportMock.On("SetReception", mock.Anything).Once()
	w := hdlc.New(transportMock, replyTimeout, 3, interOctetTimeout, 16, 73, 1)

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

	w := hdlc.New(transportMock, replyTimeout, 3, interOctetTimeout, 16, 2, 1)
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

func sendReceive(tm *mocks.TransportMock, rdc dlms.DataChannel, in string, out string) {
	tm.On("Send", decodeHexString(in)).Run(func(_ mock.Arguments) {
		if rdc != nil {
			rdc <- decodeHexString(out)
		}
	}).Return(nil).Once()
}

func decodeHexString(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
