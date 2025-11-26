package dlmsclient_test

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/circutor-library/gosem/pkg/dlms"
	"gitlab.com/circutor-library/gosem/pkg/dlms/mocks"
	"gitlab.com/circutor-library/gosem/pkg/dlmsclient"
)

func TestClient_Connect(t *testing.T) {
	tm := mocks.NewTransportMock(t)
	tm.On("Connect").Return(nil).Once()

	tm.On("SetReception", mock.Anything).Once()
	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	err := c.Connect()
	assert.NoError(t, err)

	tm.On("IsConnected").Return(true).Once()
	assert.True(t, c.IsConnected())

	tm.AssertExpectations(t)
}

func TestClient_ConnectFail(t *testing.T) {
	tm := mocks.NewTransportMock(t)
	tm.On("Connect").Return(fmt.Errorf("error connecting"))

	tm.On("SetReception", mock.Anything).Once()
	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	err := c.Connect()
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorCommunicationFailed, clientError.Code())
}

func TestClient_Disconnect(t *testing.T) {
	tm := mocks.NewTransportMock(t)
	tm.On("Connect").Return(nil).Once()
	tm.On("Disconnect").Return(fmt.Errorf("error disconnecting")).Once()

	tm.On("SetReception", mock.Anything).Once()
	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	err := c.Disconnect()
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorCommunicationFailed, clientError.Code())

	c.Connect()

	tm.On("Disconnect").Return(nil).Once()
	err = c.Disconnect()
	assert.NoError(t, err)

	tm.AssertExpectations(t)
}

func TestClient_Associate(t *testing.T) {
	c, tm, _ := associate(t)

	tm.On("IsConnected").Return(true).Once()
	assert.True(t, c.IsAssociated())

	tm.On("Disconnect").Return(nil).Once()
	c.Disconnect()

	tm.On("IsConnected").Return(false).Once()
	assert.False(t, c.IsAssociated())

	tm.AssertExpectations(t)
}

func TestClient_InvalidPassword(t *testing.T) {
	tm := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 10)
	tm.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	settings, _ := dlms.NewSettingsWithLowAuthentication([]byte("00000001"))
	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	tm.On("Connect").Return(nil).Once()
	c.Connect()

	tm.On("IsConnected").Return(true).Once()
	sendReceive(tm, rdc, "6036A1090607608574050801018A0207808B0760857405080201AC0A80083030303030303031BE10040E01000000065F1F040000181F0100", "6129A109060760857405080101A203020101A305A10302010DBE10040E0800065F1F040000101400800007")

	err := c.Associate()
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidPassword, clientError.Code())

	tm.AssertExpectations(t)
}

func TestClient_InvalidPasswordLGZ(t *testing.T) {
	tm := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 10)
	tm.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	settings, _ := dlms.NewSettingsWithLowAuthentication([]byte("00000002"))
	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	tm.On("Connect").Return(nil).Once()
	c.Connect()

	tm.On("IsConnected").Return(true).Once()
	sendReceive(tm, rdc, "6036A1090607608574050801018A0207808B0760857405080201AC0A80083030303030303032BE10040E01000000065F1F040000181F0100", "6117A109060760857405080101A203020101A305A10302010D")

	err := c.Associate()
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidPassword, clientError.Code())

	tm.AssertExpectations(t)
}

func TestClient_AssociationRejected(t *testing.T) {
	tm := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 10)
	tm.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	settings, _ := dlms.NewSettingsWithLowAuthentication([]byte("00000002"))
	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	tm.On("Connect").Return(nil).Once()
	c.Connect()

	tm.On("IsConnected").Return(true).Once()
	sendReceive(tm, rdc, "6036A1090607608574050801018A0207808B0760857405080201AC0A80083030303030303032BE10040E01000000065F1F040000181F0100", "611FA109060760857405080101A203020101A305A103020101BE0604040E010604")

	err := c.Associate()
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorAuthenticationFailed, clientError.Code())

	tm.AssertExpectations(t)
}

func TestClient_AssociationWithExceptionResponse(t *testing.T) {
	tm := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 10)
	tm.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	settings, _ := dlms.NewSettingsWithLowAuthentication([]byte("00000002"))
	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	tm.On("Connect").Return(nil).Once()
	c.Connect()

	tm.On("IsConnected").Return(true).Once()
	sendReceive(tm, rdc, "6036A1090607608574050801018A0207808B0760857405080201AC0A80083030303030303032BE10040E01000000065F1F040000181F0100", "D80101")

	err := c.Associate()
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorAuthenticationFailed, clientError.Code())

	tm.AssertExpectations(t)
}

func TestClient_AssociationWithConfirmedServiceError(t *testing.T) {
	tm := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 10)
	tm.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	settings, _ := dlms.NewSettingsWithLowAuthentication([]byte("00000002"))
	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	tm.On("Connect").Return(nil).Once()
	c.Connect()

	tm.On("IsConnected").Return(true).Once()
	sendReceive(tm, rdc, "6036A1090607608574050801018A0207808B0760857405080201AC0A80083030303030303032BE10040E01000000065F1F040000181F0100", "0E010302")

	err := c.Associate()
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorAuthenticationFailed, clientError.Code())

	tm.AssertExpectations(t)
}

func TestClient_AssociationWithWrongKeys(t *testing.T) {
	tm := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 10)
	tm.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	ciphering, _ := dlms.NewCiphering(
		dlms.SecurityLevelDedicatedKey,
		dlms.SecurityEncryption|dlms.SecurityAuthentication,
		decodeHexString("4349520000000001"),
		decodeHexString("00112233445566778899AABBCCDDEEFF"),
		0x00000059,
		decodeHexString("00112233445566778899AABBCCDDEEFF"),
	)
	ciphering.DedicatedKey = decodeHexString("5E168412318BA71848C99B2B2AB33294")

	settings, _ := dlms.NewSettingsWithLowAuthenticationAndCiphering([]byte("JuS66BCZ"), ciphering)
	settings.MaxPduRecvSize = 512

	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	tm.On("Connect").Return(nil).Once()
	c.Connect()

	tm.On("IsConnected").Return(true)
	sendReceive(tm, rdc, "6066A109060760857405080103A60A040843495200000000018A0207808B0760857405080201AC0A80084A7553363642435ABE3404322130300000005992D807DBCF8533E9AD675AE0948241FB8E6CF9AFA7006BAA134A473C9151B3362F56DC12F89E85DA97E176",
		"613EA109060760857405080103A203020101A305A103020101A40A04084B464D3434383831BE1904172E153000000027AB078E5DA8EECF61040812F75CB5B5EA")

	err := c.Associate()
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorWrongKeys, clientError.Code())

	tm.AssertExpectations(t)
}

func TestClient_AssociationWithInvocationCounter(t *testing.T) {
	tm := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 10)
	tm.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	ciphering, _ := dlms.NewCiphering(
		dlms.SecurityLevelDedicatedKey,
		dlms.SecurityEncryption|dlms.SecurityAuthentication,
		decodeHexString("4349520000000001"),
		decodeHexString("00112233445566778899AABBCCDDEEFF"),
		0x00000059,
		decodeHexString("00112233445566778899AABBCCDDEEFF"),
	)
	ciphering.DedicatedKey = decodeHexString("5E168412318BA71848C99B2B2AB33294")

	settings, _ := dlms.NewSettingsWithLowAuthenticationAndCiphering([]byte("JuS66BCZ"), ciphering)
	settings.MaxPduRecvSize = 512

	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	tm.On("Connect").Return(nil).Once()
	c.Connect()

	tm.On("IsConnected").Return(true)
	sendReceive(tm, rdc, "6066A109060760857405080103A60A040843495200000000018A0207808B0760857405080201AC0A80084A7553363642435ABE3404322130300000005992D807DBCF8533E9AD675AE0948241FB8E6CF9AFA7006BAA134A473C9151B3362F56DC12F89E85DA97E176",
		"611FA109060760857405080101A203020101A305A103020101BE0604040E010005")

	err := c.Associate()
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorFailureInvocationCounter, clientError.Code())

	tm.AssertExpectations(t)
}

func TestClient_AssociationWithFailureInIC(t *testing.T) {
	tm := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 10)
	tm.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	ciphering, _ := dlms.NewCiphering(
		dlms.SecurityLevelDedicatedKey,
		dlms.SecurityEncryption|dlms.SecurityAuthentication,
		decodeHexString("4349520000000001"),
		decodeHexString("00112233445566778899AABBCCDDEEFF"),
		0x00000059,
		decodeHexString("00112233445566778899AABBCCDDEEFF"),
	)
	ciphering.DedicatedKey = decodeHexString("5E168412318BA71848C99B2B2AB33294")
	ciphering.UnicastExpectedIC = 0x1000005A

	settings, _ := dlms.NewSettingsWithLowAuthenticationAndCiphering([]byte("JuS66BCZ"), ciphering)
	settings.MaxPduRecvSize = 512

	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	tm.On("Connect").Return(nil).Once()
	c.Connect()

	tm.On("IsConnected").Return(true)
	sendReceive(tm, rdc, "6066A109060760857405080103A60A040843495200000000018A0207808B0760857405080201AC0A80084A7553363642435ABE3404322130300000005992D807DBCF8533E9AD675AE0948241FB8E6CF9AFA7006BAA134A473C9151B3362F56DC12F89E85DA97E176",
		"6148A109060760857405080103A203020100A305A103020100A40A04084C475A2022604828BE230421281F300000005AE916783AF33B5317AD0E453A799A65F26AE97660CF8B14FEB7B0")

	err := c.Associate()
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorFailureInvocationCounter, clientError.Code())

	tm.AssertExpectations(t)
}

func TestClient_CloseAssociation(t *testing.T) {
	c, tm, rdc := associate(t)

	tm.On("IsConnected").Return(true).Once()
	sendReceive(tm, rdc, "6200", "6300")

	err := c.CloseAssociation()
	assert.NoError(t, err)

	tm.AssertExpectations(t)
}

func TestClient_Timeout(t *testing.T) {
	tm := mocks.NewTransportMock(t)

	tm.On("SetReception", mock.Anything).Once()
	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 5*time.Second, 100*time.Millisecond)

	// Check connection is closed after timeout.
	tm.On("Connect").Return(nil).Once()
	err := c.Connect()
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	tm.On("Disconnect").Return(nil).Once()
	time.Sleep(60 * time.Millisecond)

	// Check connection isn't closed by timeout if already is closed.
	tm.On("Connect").Return(nil).Once()
	err = c.Connect()
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	tm.On("Disconnect").Return(nil).Once()
	c.Disconnect()

	time.Sleep(60 * time.Millisecond)

	tm.AssertExpectations(t)
}

func TestClient_TimeoutRefreshWithCommunications(t *testing.T) {
	tm := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 10)
	tm.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 5*time.Second, 100*time.Millisecond)

	tm.On("Connect").Return(nil).Once()
	assert.NoError(t, c.Connect())

	tm.On("IsConnected").Return(true).Once()
	sendReceive(tm, rdc, "601DA109060760857405080101BE10040E01000000065F1F040000181F0100", "6129A109060760857405080101A203020100A305A103020100BE10040E0800065F1F040000101D00800007")
	assert.NoError(t, c.Associate())

	time.Sleep(50 * time.Millisecond)

	sendReceive(tm, rdc, "C001C100080000010000FF0300", "C401C10010003C")
	err := c.GetRequest(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3), nil)
	assert.NoError(t, err)

	time.Sleep(80 * time.Millisecond)
	tm.On("Disconnect").Return(nil).Once()
	time.Sleep(40 * time.Millisecond)

	tm.AssertExpectations(t)
}

func TestClient_GetAndSetSettings(t *testing.T) {
	tm := mocks.NewTransportMock(t)

	tm.On("SetReception", mock.Anything).Once()
	settings, _ := dlms.NewSettingsWithLowAuthentication([]byte("00000002"))
	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	settings = c.GetSettings()
	assert.Equal(t, []byte("00000002"), settings.Password)

	settings.Password = []byte("00000003")
	c.SetSettings(settings)

	settings = c.GetSettings()
	assert.Equal(t, []byte("00000003"), settings.Password)
}

func TestClient_Notification(t *testing.T) {
	c, tm, rdc := associate(t)

	notification := make(chan dlms.Notification, 10)
	c.SetNotificationChannel("My ID", notification)

	rdc <- decodeHexString("0F0063D76A0003FF")

	dc := <-notification
	assert.Equal(t, "My ID", dc.ID)
	assert.Equal(t, uint32(6543210), dc.DataNotification.InvokeIDAndPriority)

	tm.AssertExpectations(t)
}

func TestClient_NotificationWithGet(t *testing.T) {
	c, tm, rdc := associate(t)

	tm.On("Send", decodeHexString("C001C100080000010000FF0300")).Run(func(_ mock.Arguments) {
		if rdc != nil {
			rdc <- decodeHexString("0F00000461000101020509060009190900FF12004612046E120A451204FE")
			rdc <- decodeHexString("C401C10010003C")
		}
	}).Return(nil).Once()

	err := c.GetRequest(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3), nil)
	assert.NoError(t, err)

	tm.AssertExpectations(t)
}

func TestClient_CompleteCommunication(t *testing.T) {
	tm := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 10)
	tm.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	tm.On("Connect").Return(nil).Once()
	assert.NoError(t, c.Connect())

	tm.On("IsConnected").Return(true)
	sendReceive(tm, rdc, "601DA109060760857405080101BE10040E01000000065F1F040000181F0100", "6129A109060760857405080101A203020100A305A103020100BE10040E0800065F1F040000101D00800007")
	assert.NoError(t, c.Associate())
	assert.Equal(t, 128, c.GetSettings().MaxPduSendSize)

	sendReceive(tm, rdc, "C001C100080000010000FF0300", "C401C10010003C")
	err := c.GetRequest(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3), nil)
	assert.NoError(t, err)

	sendReceive(tm, rdc, "6200", "6300")
	err = c.CloseAssociation()
	assert.NoError(t, err)

	tm.AssertExpectations(t)
}

func TestClient_CompleteSecureCommunication(t *testing.T) {
	tm := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 10)
	tm.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	ciphering, _ := dlms.NewCiphering(
		dlms.SecurityLevelDedicatedKey,
		dlms.SecurityEncryption|dlms.SecurityAuthentication,
		decodeHexString("4349520000000001"),
		decodeHexString("00112233445566778899AABBCCDDEEFF"),
		0x00000059,
		decodeHexString("00112233445566778899AABBCCDDEEFF"),
	)
	ciphering.DedicatedKey = decodeHexString("5E168412318BA71848C99B2B2AB33294")

	settings, _ := dlms.NewSettingsWithLowAuthenticationAndCiphering([]byte("JuS66BCZ"), ciphering)
	settings.MaxPduRecvSize = 512

	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	tm.On("Connect").Return(nil).Once()
	assert.NoError(t, c.Connect())

	tm.On("IsConnected").Return(true)
	sendReceive(tm, rdc, "6066A109060760857405080103A60A040843495200000000018A0207808B0760857405080201AC0A80084A7553363642435ABE3404322130300000005992D807DBCF8533E9AD675AE0948241FB8E6CF9AFA7006BAA134A473C9151B3362F56DC12F89E85DA97E176",
		"6148A109060760857405080103A203020100A305A103020100A40A04084C475A2022604828BE230421281F300000005AE916783AF33B5317AD0E453A799A65F26AE97660CF8B14FEB7B0")
	assert.NoError(t, c.Associate())
	assert.Equal(t, 250, c.GetSettings().MaxPduSendSize)

	sendReceive(tm, rdc, "D01E3000000001D3B903996D9508C5B6BCDEB025DD1800A5C92775FB55F317CF", "D4233000000001AA07A549F82E6B8EEA919659D91689BF995BE6F93C95A7208718A3B84EE4")
	err := c.GetRequest(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 2), nil)
	assert.NoError(t, err)

	// Get fails due failure invocation counter (two replies with the same invocation counter)
	sendReceive(tm, rdc, "D01E3000000002CDA00847BC0032D323FB29C26A51683A141E2CE1C14FDEB851", "D4233000000001AA07A549F82E6B8EEA919659D91689BF995BE6F93C95A7208718A3B84EE4")
	err = c.GetRequest(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 2), nil)
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorFailureInvocationCounter, clientError.Code())

	// Should work with non ciphered reply
	sendReceive(tm, rdc, "d01e300000000360dee75200e0f8a10984a2916149e31f2fc5ba7e9b264e11c5", "C401C10010003C")
	err = c.GetRequest(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 2), nil)
	assert.NoError(t, err)

	sendReceive(tm, rdc, "6239800100BE3404322130300000005A8E9B83D641B89FAAF36DA504132C34F87E4BA66175A7DCED015460239699C72C18C06DB29C54673B83BAC0", "6328800100BE230421281F300000005BCD34827974EDCF8B1DAB306F62C58AB42052DB67361377507825")
	assert.NoError(t, c.CloseAssociation())

	assert.Equal(t, uint32(0x0000005B), c.GetSettings().Ciphering.UnicastKeyIC)
	assert.Equal(t, uint32(0x00000004), c.GetSettings().Ciphering.DedicatedKeyIC)

	tm.AssertExpectations(t)
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
