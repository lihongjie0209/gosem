package dlmsclient_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/circutor-library/gosem/pkg/axdr"
	"gitlab.com/circutor-library/gosem/pkg/dlms"
	"gitlab.com/circutor-library/gosem/pkg/dlms/mocks"
	"gitlab.com/circutor-library/gosem/pkg/dlmsclient"
)

func TestClient_GetRequest(t *testing.T) {
	c, tm, rdc := associate(t)

	var data int16

	sendReceive(tm, rdc, "C001C100080000010000FF0300", "C401C10010003C")
	err := c.GetRequest(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3), &data)
	assert.NoError(t, err)
	assert.Equal(t, int16(0x003C), data)

	tm.AssertExpectations(t)
}

func TestClient_GetRequestFail(t *testing.T) {
	c, tm, rdc := associate(t)

	var data int32
	clockAttributeDescriptor := dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3)

	// Get failed
	sendReceive(tm, rdc, "C001C100080000010000FF0300", "C401C10102")

	err := c.GetRequest(clockAttributeDescriptor, &data)
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorGetRejected, clientError.Code())

	// Unexpected response
	sendReceive(tm, rdc, "C001C100080000010000FF0300", "0E010203")

	err = c.GetRequest(clockAttributeDescriptor, &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	// Invalid response
	sendReceive(tm, rdc, "C001C100080000010000FF0300", "AE12")

	err = c.GetRequest(clockAttributeDescriptor, &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	// Response type doesn't match
	sendReceive(tm, rdc, "C001C100080000010000FF0300", "C401C10010003C")

	err = c.GetRequest(clockAttributeDescriptor, &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	// Send failed
	tm.On("Send", decodeHexString("C001C100080000010000FF0300")).Return(fmt.Errorf("error")).Once()
	tm.On("IsConnected").Return(false).Once()

	err = c.GetRequest(clockAttributeDescriptor, &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorCommunicationFailed, clientError.Code())

	// Not associated
	tm.On("Disconnect").Return(nil).Once()
	c.Disconnect()

	err = c.GetRequest(clockAttributeDescriptor, &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidState, clientError.Code())

	// nil attribute descriptor
	err = c.GetRequest(nil, &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidParameter, clientError.Code())

	tm.AssertExpectations(t)
}

func TestClient_GetRequestWithDataBlock(t *testing.T) {
	c, tm, rdc := associate(t)

	var data []uint32

	sendReceive(tm, rdc, "C001C100070100630100FF0200", "C402C10000000001000C010506000000010600000002")
	sendReceive(tm, rdc, "C002C100000001", "C402C10000000002000A06000000030600000004")
	sendReceive(tm, rdc, "C002C100000002", "C402C1010000000300050600000005")
	err := c.GetRequest(dlms.CreateAttributeDescriptor(7, "1-0:99.1.0.255", 2), &data)
	assert.NoError(t, err)
	assert.Len(t, data, 5)

	tm.AssertExpectations(t)
}

func TestClient_GetRequestWithDataBlockFail(t *testing.T) {
	c, tm, rdc := associate(t)

	var data []uint32

	// Get failed
	sendReceive(tm, rdc, "C001C100070100630100FF0200", "C402C100000000010102")
	err := c.GetRequest(dlms.CreateAttributeDescriptor(7, "1-0:99.1.0.255", 2), &data)
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorGetRejected, clientError.Code())

	// Invalid block number
	sendReceive(tm, rdc, "C001C100070100630100FF0200", "C402C10000000002000C010506000000010600000002")
	err = c.GetRequest(dlms.CreateAttributeDescriptor(7, "1-0:99.1.0.255", 2), &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	// Invalid response
	sendReceive(tm, rdc, "C001C100070100630100FF0200", "C402C10000000001000C010506000000010600000002")
	sendReceive(tm, rdc, "C002C100000001", "AE12")
	err = c.GetRequest(dlms.CreateAttributeDescriptor(7, "1-0:99.1.0.255", 2), &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	// Unexpected response
	sendReceive(tm, rdc, "C001C100070100630100FF0200", "C402C10000000001000C010506000000010600000002")
	sendReceive(tm, rdc, "C002C100000001", "0E010203")
	err = c.GetRequest(dlms.CreateAttributeDescriptor(7, "1-0:99.1.0.255", 2), &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	// Invalid data
	sendReceive(tm, rdc, "C001C100070100630100FF0200", "C402C10100000001000C010506000000010600000002")
	err = c.GetRequest(dlms.CreateAttributeDescriptor(7, "1-0:99.1.0.255", 2), &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	tm.AssertExpectations(t)
}

func TestClient_GetRequestRequestWithSelectiveAccess(t *testing.T) {
	c, tm, rdc := associate(t)

	var data []uint32

	sendReceive(tm, rdc, "C001C100070100630100FF0201010204020412000809060000010000FF0F02120000090C07E40101030A000000000000090C07E40101030B0000000000000100", "C401C100010206000000010600000002")
	timeStart := time.Date(2020, time.January, 1, 10, 0, 0, 0, time.UTC)
	timeEnd := time.Date(2020, time.January, 1, 11, 0, 0, 0, time.UTC)
	atrDescriptor := dlms.CreateAttributeDescriptor(7, "1-0:99.1.0.255", 2)
	firstStruct := *axdr.CreateAxdrStructure([]*axdr.DlmsData{axdr.CreateAxdrLongUnsigned(8), axdr.CreateAxdrOctetString("0000010000ff"), axdr.CreateAxdrInteger(2), axdr.CreateAxdrLongUnsigned(0)})

	selectiveAccess := *axdr.CreateAxdrStructure([]*axdr.DlmsData{&firstStruct, axdr.CreateAxdrOctetString(timeStart), axdr.CreateAxdrOctetString(timeEnd), axdr.CreateAxdrArray([]*axdr.DlmsData{})})

	err := c.GetRequestWithSelectiveAccess(atrDescriptor, selectiveAccess, &data)
	assert.NoError(t, err)
	assert.Len(t, data, 2)

	tm.AssertExpectations(t)
}

func TestClient_GetRequestRequestWithSelectiveAccessByDate(t *testing.T) {
	c, tm, rdc := associate(t)

	var data []uint32

	sendReceive(tm, rdc, "C001C100070100630100FF0201010204020412000809060000010000FF0F02120000090C07E40101030A000000000000090C07E40101030B0000000000000100", "C401C100010206000000010600000002")
	timeStart := time.Date(2020, time.January, 1, 10, 0, 0, 0, time.UTC)
	timeEnd := time.Date(2020, time.January, 1, 11, 0, 0, 0, time.UTC)
	err := c.GetRequestWithSelectiveAccessByDate(dlms.CreateAttributeDescriptor(7, "1-0:99.1.0.255", 2), timeStart, timeEnd, &data)
	assert.NoError(t, err)
	assert.Len(t, data, 2)

	tm.AssertExpectations(t)
}

func TestClient_GetRequestWithSelectiveAccessByDateAndValues(t *testing.T) {
	c, tm, rdc := associate(t)

	var data []uint32

	sendReceive(tm, rdc, "C001C100070100630100FF0201010204020412000809060000010000FF0F02120000090C07E40101030A000000000000090C07E40101030B0000000000000102020412000809060000010000FF0F02120000020412000109060000600A07FF0F02120000", "C401C100010206000000010600000002")
	timeStart := time.Date(2020, time.January, 1, 10, 0, 0, 0, time.UTC)
	timeEnd := time.Date(2020, time.January, 1, 11, 0, 0, 0, time.UTC)

	vad := make([]dlms.AttributeDescriptor, 2)
	vad[0] = *dlms.CreateAttributeDescriptor(8, "0.0.1.0.0.255", 2)
	vad[1] = *dlms.CreateAttributeDescriptor(1, "0.0.96.10.7.255", 2)

	err := c.GetRequestWithSelectiveAccessByDateAndValues(dlms.CreateAttributeDescriptor(7, "1-0:99.1.0.255", 2), timeStart, timeEnd, vad, &data)
	assert.NoError(t, err)
	assert.Len(t, data, 2)

	tm.AssertExpectations(t)
}

func TestClient_GetRequestWithStructOfElements(t *testing.T) {
	var data struct {
		Value1 uint  `obis:"1,1-1:94.34.100.255,2"`
		Value2 *uint `obis:"1,1-1:94.34.104.255,2"`
		Value3 *uint `obis:"70,0-0:96.3.10.255,3"`
		Value4 *uint `obis:"3,0.0.96.10.7.255,2"`
	}

	c, tm, rdc := associate(t)

	sendReceive(tm, rdc, "C001C1000101015E2264FF0200", "C401C1001104")
	sendReceive(tm, rdc, "C001C1000101015E2268FF0200", "C401C1001101")
	sendReceive(tm, rdc, "C001C10046000060030AFF0300", "C401C10109")
	sendReceive(tm, rdc, "C001C100030000600A07FF0200", "C401C10009062043594B3132")
	err := c.GetRequestWithStructOfElements(&data)
	assert.NoError(t, err)
	assert.Equal(t, uint(4), data.Value1)
	assert.Equal(t, uint(1), *data.Value2)
	assert.Nil(t, data.Value3)
	assert.Nil(t, data.Value4)

	tm.AssertExpectations(t)
}

func TestClient_GetRequestWithNestedStructOfElements(t *testing.T) {
	type data2 struct {
		Value uint `obis:"1,1-1:94.34.104.255,2"`
	}

	type data3 struct {
		Value *uint `obis:"70,0-0:96.3.10.255,3"`
	}

	var data struct {
		Value1 uint `obis:"1,1-1:94.34.100.255,2"`
		Data2  data2
		Data3  data3
	}

	c, tm, rdc := associate(t)

	sendReceive(tm, rdc, "C001C1000101015E2264FF0200", "C401C1001104")
	sendReceive(tm, rdc, "C001C1000101015E2268FF0200", "C401C1001101")
	sendReceive(tm, rdc, "C001C10046000060030AFF0300", "C401C10109")
	err := c.GetRequestWithStructOfElements(&data)
	assert.NoError(t, err)
	assert.Equal(t, uint(4), data.Value1)
	assert.Equal(t, uint(1), data.Data2.Value)
	assert.Nil(t, data.Data3.Value)

	tm.AssertExpectations(t)
}

func TestClient_CheckRequestWithStructOfElements(t *testing.T) {
	var data struct {
		Value1 *uint8 `obis:"1,1-1:94.34.104.255,2"`
		Value2 string `obis:"1,0-0:96.1.1.255,2"`
		Value3 []int8 `obis:"3,0-0:94.34.4.255,2"`
	}

	value1 := uint8(4)
	value2 := "2043594B3132"

	data.Value1 = &value1
	data.Value2 = value2
	data.Value3 = []int8{3, 4}

	c, tm, rdc := associate(t)

	sendReceive(tm, rdc, "C001C1000101015E2268FF0200", "C401C1001104")
	sendReceive(tm, rdc, "C001C100010000600101FF0200", "C401C10009062043594B3132")
	sendReceive(tm, rdc, "C001C1000300005E2204FF0200", "C401C10001020F030F04")
	err := c.CheckRequestWithStructOfElements(&data)
	assert.NoError(t, err)

	tm.AssertExpectations(t)

	// If the first value is nil, just check the second value
	data.Value1 = nil

	sendReceive(tm, rdc, "C001C100010000600101FF0200", "C401C10009062043594B3132")
	sendReceive(tm, rdc, "C001C1000300005E2204FF0200", "C401C10001020F030F04")
	err = c.CheckRequestWithStructOfElements(&data)
	assert.NoError(t, err)

	tm.AssertExpectations(t)

	// If the first value doesn't match, should fail
	value1 = 8
	data.Value1 = &value1

	sendReceive(tm, rdc, "C001C1000101015E2268FF0200", "C401C1001104")
	err = c.CheckRequestWithStructOfElements(&data)
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorCheckDoesNotMatch, clientError.Code())

	tm.AssertExpectations(t)

	// Values in a slice should also be checked
	data.Value1 = nil

	sendReceive(tm, rdc, "C001C100010000600101FF0200", "C401C10009062043594B3132")
	sendReceive(tm, rdc, "C001C1000300005E2204FF0200", "C401C10001020F030F05")
	err = c.CheckRequestWithStructOfElements(&data)
	assert.Error(t, err)

	tm.AssertExpectations(t)
}

func TestClient_GetRequest_InterruptedBlockTransfer(t *testing.T) {
	c, tm, rdc := associate(t)

	// Block 1 of expected 2: carries a partial array.
	// Full array would be: TagArray(3 elems) + Long(100) + Long(200) + Long(300)
	// Block 1 carries only the first 8 bytes: header + Long(100) + Long(200).
	// Encoded request:  GetRequestNormal, class=8, OBIS=0-0:1.0.0.255, attr=3
	// Encoded response: GetResponseWithDataBlock, block=1, lastBlock=false,
	//                   data = 01 03 10 00 64 10 00 C8
	sendReceive(tm, rdc, "C001C100080000010000FF0300", "C402C10000000001000801031000641000C8")

	// GetRequestNext for block 1 → communication failure
	tm.On("Send", decodeHexString("C002C100000001")).Return(fmt.Errorf("connection reset")).Once()
	tm.On("IsConnected").Return(true).Once()

	var data []int16
	err := c.GetRequest(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3), &data)

	var partialErr *dlms.Error
	assert.ErrorAs(t, err, &partialErr)
	assert.Equal(t, dlms.ErrorPartialTransfer, partialErr.Code())
	// Two complete Long elements were recovered from the partial block.
	assert.Equal(t, []int16{100, 200}, data)

	tm.AssertExpectations(t)
}

func associate(t *testing.T) (dlms.Client, *mocks.TransportMock, dlms.DataChannel) {
	t.Helper()

	tm := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 10)
	tm.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	tm.On("Connect").Return(nil).Once()
	c.Connect()

	tm.On("IsConnected").Return(true).Once()
	sendReceive(tm, rdc, "601DA109060760857405080101BE10040E01000000065F1F040000181F0100", "6129A109060760857405080101A203020100A305A103020100BE10040E0800065F1F040000101D00800007")

	err := c.Associate()
	assert.NoError(t, err)

	return c, tm, rdc
}
