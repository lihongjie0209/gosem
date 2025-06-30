package wrapper

import (
	"encoding/binary"
	"fmt"
	"log"

	"gitlab.com/circutor-library/gosem/pkg/dlms"
)

const (
	version      = 1
	headerLength = 8
	maxLength    = 2048
)

type wrapper struct {
	transport   dlms.Transport
	source      uint16
	destination uint16
	dc          dlms.DataChannel
	tc          dlms.DataChannel
	logger      *log.Logger
}

func New(transport dlms.Transport, client int, server int) dlms.Transport {
	w := &wrapper{
		transport:   transport,
		source:      uint16(client),
		destination: uint16(server),
		dc:          nil,
		tc:          make(dlms.DataChannel, 10),
		logger:      nil,
	}

	transport.SetReception(w.tc)

	go w.manager()

	return w
}

func (w *wrapper) Close() {
	w.transport.Close()
	if w.dc != nil {
		close(w.dc)
		w.dc = nil
	}
}

func (w *wrapper) Connect() error {
	if err := w.transport.Connect(); err != nil {
		return err
	}

	return nil
}

func (w *wrapper) manager() {
	for {
		data, ok := <-w.tc
		if !ok {
			return
		}

		for {
			if len(data) == 0 {
				break
			}

			src, err := w.parseHeader(&data)
			if err != nil {
				if w.logger != nil {
					w.logger.Printf("Invalid received data: %e", err)
				}

				break
			}

			if w.dc != nil {
				w.dc <- src
			}
		}
	}
}

func (w *wrapper) Disconnect() error {
	return w.transport.Disconnect()
}

func (w *wrapper) IsConnected() bool {
	return w.transport.IsConnected()
}

func (w *wrapper) SetAddress(client int, server int) {
	w.source = uint16(client)
	w.destination = uint16(server)
}

func (w *wrapper) SetReception(dc dlms.DataChannel) {
	if w.dc != nil {
		close(w.dc)
	}

	w.dc = dc
}

func (w *wrapper) Send(src []byte) error {
	if !w.transport.IsConnected() {
		return fmt.Errorf("not connected")
	}

	if len(src) > (maxLength - headerLength) {
		return fmt.Errorf("message too long")
	}

	uri := make([]byte, headerLength+len(src))

	binary.BigEndian.PutUint16(uri[0:2], uint16(version))
	binary.BigEndian.PutUint16(uri[2:4], w.source)
	binary.BigEndian.PutUint16(uri[4:6], w.destination)
	binary.BigEndian.PutUint16(uri[6:8], uint16(len(src)))

	copy(uri[headerLength:], src)

	return w.transport.Send(uri)
}

func (w *wrapper) SetLogger(logger *log.Logger) {
	w.logger = logger
	w.transport.SetLogger(logger)
}

func (w *wrapper) parseHeader(ori *[]byte) ([]byte, error) {
	src := *ori

	if len(src) < headerLength {
		return nil, fmt.Errorf("message too short, received only %d bytes", len(src))
	}

	receivedVersion := int(binary.BigEndian.Uint16(src[0:2]))
	if receivedVersion != version {
		return nil, fmt.Errorf("invalid version, expected %d, received %d", version, receivedVersion)
	}

	receivedDestination := binary.BigEndian.Uint16(src[2:4])
	if receivedDestination != w.destination {
		return nil, fmt.Errorf("invalid destination, expected %d, received %d", w.destination, receivedDestination)
	}

	receivedSource := binary.BigEndian.Uint16(src[4:6])
	if receivedSource != w.source {
		return nil, fmt.Errorf("invalid source, expected %d, received %d", w.source, receivedSource)
	}

	length := int(binary.BigEndian.Uint16(src[6:8])) + headerLength
	if length > maxLength {
		return nil, fmt.Errorf("expected message too long (%d)", length)
	}

	if len(src) < length {
		return nil, fmt.Errorf("message length too much short, expected %d, received %d", length, len(src))
	}

	(*ori) = (*ori)[length:]

	return src[headerLength:length], nil
}
