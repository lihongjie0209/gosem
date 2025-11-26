package dlmsclient

import (
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"gitlab.com/circutor-library/gosem/pkg/dlms"
)

const (
	unicastInvokeID   uint8 = 0xC1
	broadcastInvokeID uint8 = 0x87
)

type client struct {
	settings           dlms.Settings
	transport          dlms.Transport
	replyTimeout       time.Duration
	associationTimeout time.Duration
	isAssociated       bool
	timeoutTimer       *time.Timer
	tc                 dlms.DataChannel
	dc                 dlms.DataChannel
	notificationID     string
	notificationChan   chan dlms.Notification
	mutex              sync.Mutex
	subsMutex          sync.Mutex
	logger             *log.Logger
}

func New(settings dlms.Settings, transport dlms.Transport, replyTimeout time.Duration, associationTimeout time.Duration) dlms.Client {
	c := &client{
		settings:           settings,
		transport:          transport,
		replyTimeout:       replyTimeout,
		associationTimeout: associationTimeout,
		isAssociated:       false,
		timeoutTimer:       nil,
		tc:                 make(dlms.DataChannel, 10),
		dc:                 nil,
		notificationID:     "",
		notificationChan:   nil,
		mutex:              sync.Mutex{},
		subsMutex:          sync.Mutex{},
		logger:             nil,
	}

	transport.SetReception(c.tc)

	go c.manager()

	return c
}

func (c *client) SetAddress(client int, server int) {
	c.transport.SetAddress(client, server)
}

func (c *client) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.transport.Connect()
	if err != nil {
		return dlms.NewError(dlms.ErrorCommunicationFailed, fmt.Sprintf("error connecting: %v", err))
	}

	if c.associationTimeout != 0 {
		c.timeoutTimer = time.AfterFunc(c.associationTimeout, func() {
			c.Disconnect()
		})
	}

	return nil
}

func (c *client) Disconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.closeAssociation()

	err := c.transport.Disconnect()
	if err != nil {
		return dlms.NewError(dlms.ErrorCommunicationFailed, fmt.Sprintf("error disconnecting: %v", err))
	}

	return nil
}

func (c *client) IsConnected() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	isConnected := c.transport.IsConnected()
	if !isConnected {
		c.closeAssociation()
	}

	return isConnected
}

func (c *client) SetNotificationChannel(id string, nc chan dlms.Notification) {
	c.subsMutex.Lock()
	defer c.subsMutex.Unlock()

	c.notificationID = id
	c.notificationChan = nc
}

func (c *client) GetSettings() dlms.Settings {
	return c.settings
}

func (c *client) SetSettings(settings dlms.Settings) {
	c.settings = settings
}

func (c *client) SetLogger(logger *log.Logger) {
	c.logger = logger
	c.transport.SetLogger(logger)
}

func (c *client) Associate() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.transport.IsConnected() {
		return dlms.NewError(dlms.ErrorInvalidState, "not connected")
	}

	src, err := dlms.EncodeAARQ(&c.settings)
	if err != nil {
		return dlms.NewError(dlms.ErrorInvalidParameter, fmt.Sprintf("error encoding AARQ: %v", err))
	}

	out, err := c.sendReceive(src)
	if err != nil {
		return err
	}

	aare, err := dlms.DecodeAARE(&c.settings, &out)
	if err != nil {
		er, eerr := dlms.DecodeExceptionResponse(&out)
		if eerr == nil {
			return dlms.NewError(dlms.ErrorAuthenticationFailed, fmt.Sprintf("association failed (exception): %d - %d", er.StateError, er.ServiceError))
		}

		cse, eerr := dlms.DecodeConfirmedServiceError(&out)
		if eerr == nil {
			return dlms.NewError(dlms.ErrorAuthenticationFailed, fmt.Sprintf("association failed (service error): %d - %d - %d", cse.ConfirmedServiceError, cse.ServiceError, cse.Value))
		}

		return dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("error decoding AARE: %v", err))
	}

	if aare.AssociationResult != dlms.AssociationResultAccepted || aare.SourceDiagnostic != dlms.SourceDiagnosticNone || aare.InitiateResponse == nil {
		if aare.SourceDiagnostic == dlms.SourceDiagnosticAuthenticationFailure {
			return dlms.NewError(dlms.ErrorInvalidPassword, fmt.Sprintf("association failed (invalid password): %d - %d", aare.AssociationResult, aare.SourceDiagnostic))
		}

		if aare.ConfirmedServiceError != nil && aare.ConfirmedServiceError.ServiceError == dlms.TagErrOtherError && aare.ConfirmedServiceError.Value == dlms.TagOtherDecipheringError {
			return dlms.NewError(dlms.ErrorWrongKeys, fmt.Sprintf("association failed (invalid keys): %d - %d (%s)", aare.AssociationResult, aare.SourceDiagnostic, aare.ConfirmedServiceError.String()))
		}

		if aare.ConfirmedServiceError != nil && aare.ConfirmedServiceError.ServiceError == dlms.TagErrApplicationReference && aare.ConfirmedServiceError.Value == dlms.TagApplicationReferenceProviderCommunicationError {
			return dlms.NewError(dlms.ErrorFailureInvocationCounter, fmt.Sprintf("association failed (invalid invocation counter): %d - %d (%s)", aare.AssociationResult, aare.SourceDiagnostic, aare.ConfirmedServiceError.String()))
		}

		if aare.ConfirmedServiceError != nil {
			return dlms.NewError(dlms.ErrorAuthenticationFailed, fmt.Sprintf("association failed: %d - %d (%s)", aare.AssociationResult, aare.SourceDiagnostic, aare.ConfirmedServiceError.String()))
		}

		return dlms.NewError(dlms.ErrorAuthenticationFailed, fmt.Sprintf("association failed: %d - %d", aare.AssociationResult, aare.SourceDiagnostic))
	}

	if aare.ReceivedIC != nil {
		if *aare.ReceivedIC < c.settings.Ciphering.UnicastExpectedIC {
			return dlms.NewError(dlms.ErrorFailureInvocationCounter, fmt.Sprintf("wrong expected invocation counter: %d is lower than %d", *aare.ReceivedIC, c.settings.Ciphering.UnicastExpectedIC))
		}
		c.settings.Ciphering.UnicastExpectedIC = *aare.ReceivedIC + 1
	}

	if aare.InitiateResponse != nil {
		maxPduSendSize := int(aare.InitiateResponse.ServerMaxReceivePduSize)
		if maxPduSendSize < c.settings.MaxPduSendSize {
			c.settings.MaxPduSendSize = maxPduSendSize
		}
	}

	c.isAssociated = true
	return nil
}

func (c *client) CloseAssociation() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.transport.IsConnected() {
		return dlms.NewError(dlms.ErrorInvalidState, "not connected")
	}

	src, err := dlms.EncodeRLRQ(&c.settings)
	if err != nil {
		return dlms.NewError(dlms.ErrorInvalidParameter, fmt.Sprintf("error encoding RLRQ: %v", err))
	}

	out, err := c.sendReceive(src)
	if err != nil {
		return err
	}

	_, err = dlms.DecodeRLRE(&out)
	if err != nil {
		return dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("error decoding RLRE: %v", err))
	}

	c.closeAssociation()

	return nil
}

func (c *client) IsAssociated() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.transport.IsConnected() {
		c.closeAssociation()
	}

	return c.isAssociated
}

func (c *client) manager() {
	for {
		data, ok := <-c.tc
		if !ok {
			c.closeAssociation()

			return
		}

		dn, err := dlms.DecodeDataNotification(&data)
		if err == nil {
			nc := dlms.Notification{
				ID:               c.notificationID,
				DataNotification: dn,
			}

			if c.timeoutTimer != nil {
				c.timeoutTimer.Reset(c.associationTimeout)
			}

			c.subsMutex.Lock()
			if c.notificationChan != nil {
				c.notificationChan <- nc
			}
			c.subsMutex.Unlock()
		} else {
			c.subsMutex.Lock()
			if c.dc != nil {
				c.dc <- data
			}
			c.subsMutex.Unlock()
		}
	}
}

func (c *client) sendReceive(src []byte) ([]byte, error) {
	c.subscribe()
	defer c.unsubscribe()

	if c.settings.UseBroadcast {
		if twb, ok := c.transport.(dlms.TransportWithBroadcast); ok {
			return nil, twb.SendBroadcast(src)
		}
	}

	err := c.transport.Send(src)
	if err != nil {
		return nil, dlms.NewError(dlms.ErrorCommunicationFailed, fmt.Sprintf("error sending data: %v", err))
	}

	if c.settings.UseBroadcast {
		return nil, nil
	}

	// Wait for the device response
	timeout := time.NewTimer(c.replyTimeout)
	defer timeout.Stop()

	select {
	case data := <-c.dc:
		return data, nil
	case <-timeout.C:
		return nil, dlms.NewError(dlms.ErrorCommunicationFailed, "timeout reached")
	}
}

func (c *client) subscribe() {
	c.subsMutex.Lock()
	defer c.subsMutex.Unlock()

	c.dc = make(dlms.DataChannel, 1)
}

func (c *client) unsubscribe() {
	c.subsMutex.Lock()
	defer c.subsMutex.Unlock()

	c.dc = nil
}

func (c *client) encodeSendReceiveAndDecode(req dlms.CosemPDU) (dlms.CosemPDU, error) {
	if !c.isAssociated {
		return nil, dlms.NewError(dlms.ErrorInvalidState, "client is not associated")
	}

	src, err := req.Encode()
	if err != nil {
		return nil, dlms.NewError(dlms.ErrorInvalidParameter, fmt.Sprintf("error encoding PDU: %v", err))
	}

	if c.settings.Ciphering.Level != dlms.SecurityLevelNone {
		if c.logger != nil {
			c.logger.Printf("TC: %s", encodeHexString(src))
		}

		src, err = c.cipherData(src)
		if err != nil {
			return nil, err
		}
	}

	out, err := c.sendReceive(src)
	if err != nil {
		if !c.transport.IsConnected() {
			c.closeAssociation()
		}

		return nil, err
	}

	if c.timeoutTimer != nil {
		c.timeoutTimer.Reset(c.associationTimeout)
	}

	if out == nil && c.settings.UseBroadcast {
		return nil, nil
	}

	if c.settings.Ciphering.Level != dlms.SecurityLevelNone && dlms.IsSecuredTag(dlms.CosemTag(out[0])) {
		out, err = c.decipherData(out)
		if err != nil {
			return nil, err
		}

		if c.logger != nil {
			c.logger.Printf("RC: %s", encodeHexString(out))
		}
	}

	pdu, err := dlms.DecodeCosem(&out)
	if err != nil {
		err = dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("error decoding PDU: %v", err))
		return nil, err
	}

	return pdu, nil
}

func encodeHexString(b []byte) string {
	return strings.ToUpper(hex.EncodeToString(b))
}

func (c *client) cipherData(src []byte) ([]byte, error) {
	tag := dlms.CosemTag(src[0])
	if tag != dlms.TagGetRequest && tag != dlms.TagSetRequest && tag != dlms.TagActionRequest {
		return nil, fmt.Errorf("unexpected tag %d", tag)
	}

	cipher := dlms.Cipher{
		Security:    c.settings.Ciphering.Security,
		SystemTitle: c.settings.Ciphering.SystemTitle,
		AuthKey:     c.settings.Ciphering.AuthenticationKey,
	}

	if c.settings.Ciphering.Level == dlms.SecurityLevelGlobalKey {
		switch dlms.CosemTag(src[0]) {
		case dlms.TagGetRequest:
			cipher.Tag = dlms.TagGloGetRequest
		case dlms.TagSetRequest:
			cipher.Tag = dlms.TagGloSetRequest
		case dlms.TagActionRequest:
			cipher.Tag = dlms.TagGloActionRequest
		}

		if len(c.settings.Ciphering.UnicastKey) != 16 {
			return nil, fmt.Errorf("invalid unicast key")
		}

		cipher.Key = c.settings.Ciphering.UnicastKey
		cipher.FrameCounter = c.settings.Ciphering.UnicastKeyIC
		c.settings.Ciphering.UnicastKeyIC++
	} else {
		switch dlms.CosemTag(src[0]) {
		case dlms.TagGetRequest:
			cipher.Tag = dlms.TagDedGetRequest
		case dlms.TagSetRequest:
			cipher.Tag = dlms.TagDedSetRequest
		case dlms.TagActionRequest:
			cipher.Tag = dlms.TagDedActionRequest
		}

		if len(c.settings.Ciphering.DedicatedKey) != 16 {
			return nil, fmt.Errorf("invalid dedicated key")
		}

		cipher.Key = c.settings.Ciphering.DedicatedKey
		cipher.FrameCounter = c.settings.Ciphering.DedicatedKeyIC
		c.settings.Ciphering.DedicatedKeyIC++
	}

	return dlms.CipherData(cipher, src)
}

func (c *client) decipherData(src []byte) ([]byte, error) {
	cipher := dlms.Cipher{
		Tag:         dlms.CosemTag(src[0]),
		Security:    c.settings.Ciphering.Security,
		SystemTitle: c.settings.Ciphering.SourceSystemTitle,
		AuthKey:     c.settings.Ciphering.AuthenticationKey,
	}

	if c.settings.Ciphering.Level == dlms.SecurityLevelGlobalKey {
		cipher.Key = c.settings.Ciphering.UnicastKey
	} else {
		cipher.Key = c.settings.Ciphering.DedicatedKey
	}

	out, err := dlms.DecipherData(&cipher, src)
	if err != nil {
		return nil, err
	}

	if c.settings.Ciphering.Level == dlms.SecurityLevelGlobalKey {
		if cipher.FrameCounter < c.settings.Ciphering.UnicastExpectedIC {
			return nil, dlms.NewError(dlms.ErrorFailureInvocationCounter, fmt.Sprintf("wrong expected invocation counter: %d is lower than %d", cipher.FrameCounter, c.settings.Ciphering.UnicastExpectedIC))
		}
		c.settings.Ciphering.UnicastExpectedIC = cipher.FrameCounter + 1
	} else {
		if cipher.FrameCounter < c.settings.Ciphering.DedicatedExpectedIC {
			return nil, dlms.NewError(dlms.ErrorFailureInvocationCounter, fmt.Sprintf("wrong expected invocation counter: %d is lower than %d", cipher.FrameCounter, c.settings.Ciphering.DedicatedExpectedIC))
		}
		c.settings.Ciphering.DedicatedExpectedIC = cipher.FrameCounter + 1
	}

	return out, nil
}

func (c *client) closeAssociation() {
	c.isAssociated = false
	if c.timeoutTimer != nil {
		c.timeoutTimer.Stop()
		c.timeoutTimer = nil
	}
}
