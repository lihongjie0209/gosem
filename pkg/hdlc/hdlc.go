package hdlc

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"gitlab.com/circutor-library/gosem/pkg/dlms"
)

type AddressingType int

const (
	AddressingOneByte   AddressingType = 1
	AddressingTwoBytes  AddressingType = 2
	AddressingFourBytes AddressingType = 4
)

const (
	maxInfoFieldLength        = 512
	defaultMaxInfoFieldLength = 128
	minInfoFieldLength        = 32

	maxDataLength  = 2048 - 3
	maxFrameLength = 10 + maxDataLength

	sequenceNumberLimit = 7

	startAndEndFlag  = 0x7E
	frameFormatField = 0xA000

	ControlI    = 0x00 // Information
	ControlRR   = 0x01 // Receive Ready
	ControlUI   = 0x13 // Unnumbered Information
	ControlSNRM = 0x93 // Set Normal Response Mode
	ControlDISC = 0x53 // Disconnect
	ControlUA   = 0x73 // Unnumbered Acknowledge
	ControlDM   = 0x1F // Disconnect Mode

	controlMaskI  = 0x01
	controlMaskRR = 0x0F

	finalWindowBit = 0x10
)

type ReceivedFrame struct {
	UpperAddress  int
	LowerAddress  int
	ClientAddress int
	Control       uint8
	IsSegmented   bool
	Data          []byte
	HCS           uint16
	FCS           uint16
}

type hdlc struct {
	maxInfoFieldLengthSend int
	addressing             AddressingType
	upperAddress           int
	lowerAddress           int
	clientAddress          int
	replyTimeout           time.Duration
	retries                int
	interOctetTimeout      time.Duration
	rrr                    int
	sss                    int
	transport              dlms.Transport
	dc                     dlms.DataChannel
	tc                     dlms.DataChannel
	fc                     chan *ReceivedFrame
	logger                 *log.Logger
	mutex                  sync.Mutex
}

func New(transport dlms.Transport, replyTimeout time.Duration, retries int, interOctetTimeout time.Duration, address int, client int, server int, addressing AddressingType) dlms.Transport {
	h := &hdlc{
		maxInfoFieldLengthSend: maxInfoFieldLength,
		addressing:             addressing,
		upperAddress:           server,
		lowerAddress:           address,
		clientAddress:          client,
		replyTimeout:           replyTimeout,
		retries:                retries,
		interOctetTimeout:      interOctetTimeout,
		rrr:                    0,
		sss:                    0,
		transport:              transport,
		dc:                     nil,
		tc:                     make(dlms.DataChannel, 10),
		fc:                     make(chan *ReceivedFrame, 1),
		logger:                 nil,
		mutex:                  sync.Mutex{},
	}

	transport.SetReception(h.tc)

	go h.manager()

	return h
}

func (h *hdlc) Close() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.transport.Close()
	if h.dc != nil {
		close(h.dc)
		h.dc = nil
	}

	close(h.fc)
}

func (h *hdlc) Connect() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if err := h.transport.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	h.maxInfoFieldLengthSend = maxInfoFieldLength
	h.rrr = 0
	h.sss = 0

	frameToSend := h.createFrame(ControlSNRM, nil)
	retries := 0
	var lastErr error

	for {
		rf, err := h.sendReceive(frameToSend)
		if err == nil {
			err = h.handleConnectReply(rf)
		}

		if err == nil {
			return nil
		}

		if !h.transport.IsConnected() {
			return fmt.Errorf("disconnected while waiting for response: %w", err)
		}

		lastErr = err
		retries++
		if retries > h.retries {
			h.transport.Disconnect()
			return fmt.Errorf("maximum retries reached: %w", lastErr)
		}
	}
}

func (h *hdlc) Disconnect() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if !h.transport.IsConnected() {
		return errors.New("not connected")
	}

	defer h.transport.Disconnect()

	frameToSend := h.createFrame(ControlDISC, nil)

	rf, err := h.sendReceive(frameToSend)
	if err != nil {
		return fmt.Errorf("send error: %w", err)
	}

	if rf.Control != ControlUA && rf.Control != ControlDM {
		return fmt.Errorf("invalid control byte, have %02X", rf.Control)
	}

	return nil
}

func (h *hdlc) IsConnected() bool {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.transport.IsConnected()
}

func (h *hdlc) SetAddress(client int, server int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.clientAddress = client
	h.upperAddress = server
}

func (h *hdlc) SetReception(dc dlms.DataChannel) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.dc != nil {
		close(h.dc)
	}

	h.dc = dc
}

func (h *hdlc) Send(src []byte) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if len(src) == 0 {
		return fmt.Errorf("empty data")
	}

	if len(src) > maxDataLength {
		return fmt.Errorf("data too long, have %d, max %d", len(src), maxDataLength)
	}

	if !h.transport.IsConnected() {
		return fmt.Errorf("not connected")
	}

	src = append([]byte{0xE6, 0xE6, 0x00}, src...)

	retries := 0
	remoteReady := true
	firstResponseFrame := true
	response := make([]byte, 0, maxDataLength)
	sendOffset := 0

	for {
		var frameToSend []byte

		// Send I frame if remote is ready, otherwise send RR frame
		if remoteReady {
			if sendOffset >= len(src) {
				return errors.New("remote requested data after complete request")
			}
			end := sendOffset + h.maxInfoFieldLengthSend
			if end > len(src) {
				end = len(src)
			}
			segmented := end < len(src)
			// Create control byte for I Frame
			control := uint8((h.rrr << 5) | (h.sss << 1) | finalWindowBit | ControlI)
			h.sss = h.increaseSequenceNumber(h.sss)

			if segmented {
				frameToSend = BuildSegmentedFrame(h.addressing, h.clientAddress, h.upperAddress, h.lowerAddress, control, src[sendOffset:end])
			} else {
				frameToSend = h.createFrame(control, src[sendOffset:end])
			}
			sendOffset = end
		} else {
			// Create control byte for RR Frame
			control := uint8((h.rrr << 5) | finalWindowBit | ControlRR)

			frameToSend = h.createFrame(control, nil)
		}

		rf, err := h.sendReceive(frameToSend)

		switch {
		case err != nil || rf == nil:
			// Nothing valid received
			remoteReady = false
		case (rf.Control & controlMaskI) == ControlI:
			// I frame
			if sendOffset != len(src) {
				return errors.New("received response before complete request was acknowledged")
			}
			var data []byte
			var complete bool
			data, complete, err = h.handleDataReply(rf, firstResponseFrame)
			if err == nil {
				if len(response)+len(data) > maxDataLength {
					return fmt.Errorf("segmented response is too long")
				}
				response = append(response, data...)
				firstResponseFrame = false
				if complete {
					if h.dc != nil {
						h.dc <- response
					}
					return nil
				}

				// A valid segmented response is normal flow, not a retry. Ask for
				// the next I frame and reset the transient failure budget.
				remoteReady = false
				retries = 0
				continue
			}

			remoteReady = false
		case (rf.Control & controlMaskRR) == ControlRR:
			// RR frame
			if sss := int(rf.Control>>5) & 0x07; sss != h.sss {
				h.sss = h.decreaseSequenceNumber(h.sss)
				remoteReady = false
				break
			}

			remoteReady = true
			retries = 0
			continue
		default:
			return fmt.Errorf("unexpected frame with control %02X", rf.Control)
		}

		if !h.transport.IsConnected() {
			return fmt.Errorf("disconnected while waiting for response: %w", err)
		}

		retries++
		if retries > h.retries {
			return fmt.Errorf("maximum retries reached")
		}
	}
}

func (h *hdlc) increaseSequenceNumber(seq int) int {
	seq++
	if seq > sequenceNumberLimit {
		seq = 0
	}

	return seq
}

func (h *hdlc) decreaseSequenceNumber(seq int) int {
	seq--
	if seq < 0 {
		seq = sequenceNumberLimit
	}

	return seq
}

func (h *hdlc) SetLogger(logger *log.Logger) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.logger = logger
}

func (h *hdlc) manager() {
	frame := make([]byte, 0, maxFrameLength)

	timer := time.NewTimer(h.interOctetTimeout)
	defer timer.Stop()

	for {
		select {
		case data, ok := <-h.tc:
			if !ok {
				return
			}

			frame = append(frame, data...)
			timer.Reset(h.interOctetTimeout)

			for len(frame) > 0 {
				rf := h.searchFrame(&frame)
				if rf == nil {
					break
				}

				h.fc <- rf
			}

		case <-timer.C:
			if len(frame) > 0 {
				frame = make([]byte, 0, maxFrameLength)
			}
			timer.Reset(h.interOctetTimeout)
		}
	}
}

func (h *hdlc) searchFrame(frame *[]byte) *ReceivedFrame {
	src, rest := NextFrame(*frame)
	*frame = rest

	if src == nil {
		return nil
	}

	if h.logger != nil {
		h.logger.Printf("RX: %s", encodeHexString(src))
	}

	// Parse received frame
	receivedFrame, err := h.parseFrame(src)
	if err != nil {
		if h.logger != nil {
			h.logger.Printf("Invalid received data: %v", err)
		}

		return nil
	}

	return receivedFrame
}

func (h *hdlc) parseFrame(src []byte) (*ReceivedFrame, error) {
	rf, err := ParseFrame(src)
	if err != nil {
		return nil, err
	}

	if rf.ClientAddress != h.clientAddress {
		return nil, fmt.Errorf("invalid client address, have %d, expected %d", rf.ClientAddress, h.clientAddress)
	}

	if rf.UpperAddress != h.upperAddress || rf.LowerAddress != h.lowerAddress {
		return nil, fmt.Errorf("invalid source address, have %d:%d", rf.UpperAddress, rf.LowerAddress)
	}

	return rf, nil
}

func (h *hdlc) createFrame(control uint8, data []byte) []byte {
	return BuildFrame(h.addressing, h.clientAddress, h.upperAddress, h.lowerAddress, control, data)
}

func (h *hdlc) sendReceive(src []byte) (*ReceivedFrame, error) {
	if h.logger != nil {
		h.logger.Printf("TX: %s", encodeHexString(src))
	}

	err := h.transport.Send(src)
	if err != nil {
		return nil, fmt.Errorf("send error: %w", err)
	}

	// Wait for the device response
	timeout := time.NewTimer(h.replyTimeout)
	defer timeout.Stop()

	select {
	case data := <-h.fc:
		return data, nil
	case <-timeout.C:
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

func (h *hdlc) handleConnectReply(rf *ReceivedFrame) error {
	if rf.Control != ControlUA {
		return fmt.Errorf("invalid control byte, have %02X, expected %02X", rf.Control, ControlUA)
	}

	data := rf.Data

	if len(data) == 0 {
		h.maxInfoFieldLengthSend = defaultMaxInfoFieldLength

		return nil
	}

	if len(data) < 3 {
		return fmt.Errorf("invalid UA data, have %d", len(data))
	}

	if data[0] != 0x81 || data[1] != 0x80 || data[2] != byte(len(data)-3) {
		return fmt.Errorf("invalid UA data, have %02X:%02X:%02X", data[0], data[1], data[2])
	}

	data = data[3:]

	for len(data) >= 2 {
		code := data[0]
		length := int(data[1])

		if len(data) < length+2 {
			return fmt.Errorf("invalid UA data, have %d", len(data))
		}

		if code == 0x06 {
			var maxInfoFieldLengthSend int

			switch length {
			case 0x01:
				maxInfoFieldLengthSend = int(data[2])
			case 0x02:
				maxInfoFieldLengthSend = int(binary.BigEndian.Uint16(data[2:]))
			default:
				return fmt.Errorf("unexpected length in UA data: %d, expected 1 or 2", length)
			}

			if maxInfoFieldLengthSend < minInfoFieldLength {
				maxInfoFieldLengthSend = minInfoFieldLength
			} else if maxInfoFieldLengthSend > maxInfoFieldLength {
				maxInfoFieldLengthSend = maxInfoFieldLength
			}

			h.maxInfoFieldLengthSend = maxInfoFieldLengthSend
		}

		data = data[length+2:]
	}

	return nil
}

func (h *hdlc) handleDataReply(rf *ReceivedFrame, first bool) ([]byte, bool, error) {
	rrr := int(rf.Control>>1) & 0x07
	sss := int(rf.Control>>5) & 0x07

	if rrr != h.rrr || sss != h.sss {
		return nil, false, fmt.Errorf("invalid control byte, have %02X, expected %02X", rf.Control, (h.rrr<<5)|(h.sss<<1))
	}

	h.rrr = h.increaseSequenceNumber(h.rrr)

	data := rf.Data
	if first {
		if len(data) < 3 {
			return nil, false, fmt.Errorf("invalid I frame data, have %d", len(data))
		}
		if data[0] != 0xE6 || data[1] != 0xE7 || data[2] != 0x00 {
			return nil, false, fmt.Errorf("invalid I frame data, have %02X:%02X:%02X", data[0], data[1], data[2])
		}
		data = data[3:]
	} else if len(data) == 0 {
		return nil, false, errors.New("segmented I frame has no data")
	}

	return data, !rf.IsSegmented, nil
}

func encodeHexString(b []byte) string {
	return strings.ToUpper(hex.EncodeToString(b))
}
