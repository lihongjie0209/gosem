package tcp

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.com/circutor-library/gosem/pkg/dlms"
)

const (
	maxLength = 2048
)

type tcp struct {
	port        int
	host        string
	timeout     time.Duration
	dc          dlms.DataChannel
	conn        net.Conn
	isConnected bool
	logger      *log.Logger
	managerDone chan struct{}
	mutex       sync.Mutex
}

func New(port int, host string, timeout time.Duration) dlms.Transport {
	t := &tcp{
		port:        port,
		host:        host,
		timeout:     timeout,
		dc:          nil,
		conn:        nil,
		isConnected: false,
		logger:      nil,
		managerDone: nil,
		mutex:       sync.Mutex{},
	}

	return t
}

func (t *tcp) Close() {
	t.mutex.Lock()
	t.disconnect()
	done := t.managerDone
	t.mutex.Unlock()

	if done != nil {
		<-done
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.dc != nil {
		close(t.dc)
		t.dc = nil
	}
}

func (t *tcp) Connect() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.isConnected {
		address := net.JoinHostPort(t.host, strconv.Itoa(t.port))

		conn, err := net.DialTimeout("tcp", address, t.timeout)
		if err != nil {
			if t.logger != nil {
				t.logger.Printf("Connect to %s failed: %v", address, err)
			}

			return fmt.Errorf("connect failed: %w", err)
		}

		if t.logger != nil {
			t.logger.Printf("Connected to %s", address)
		}

		t.conn = conn
		t.managerDone = make(chan struct{})
		t.isConnected = true

		go t.manager(t.conn, t.managerDone)
	}

	return nil
}

func (t *tcp) Disconnect() error {
	t.mutex.Lock()
	t.disconnect()
	done := t.managerDone
	t.mutex.Unlock()

	if done != nil {
		<-done
	}

	return nil
}

func (t *tcp) IsConnected() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.isConnected
}

func (t *tcp) SetAddress(_ int, _ int) {
}

func (t *tcp) SetReception(dc dlms.DataChannel) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.dc != nil {
		close(t.dc)
	}

	t.dc = dc
}

func (t *tcp) Send(src []byte) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.isConnected {
		return fmt.Errorf("not connected")
	}

	t.conn.SetWriteDeadline(time.Now().Add(t.timeout))

	_, err := t.conn.Write(src)
	if err != nil {
		t.disconnect()
		return fmt.Errorf("write failed: %w", err)
	}

	if t.logger != nil {
		t.logger.Printf("TX (%s): %s", t.host, encodeHexString(src))
	}

	return nil
}

func (t *tcp) SetLogger(logger *log.Logger) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.logger = logger
}

func (t *tcp) manager(conn net.Conn, done chan struct{}) {
	defer close(done)

	for {
		t.mutex.Lock()
		if !t.isConnected || t.conn != conn {
			t.mutex.Unlock()

			return
		}
		t.mutex.Unlock()

		data, err := t.read(conn)
		if err != nil {
			t.mutex.Lock()
			if t.conn == conn {
				t.disconnect()
			}
			t.mutex.Unlock()

			return
		}

		if len(data) > 0 && t.dc != nil {
			t.dc <- data
		}
	}
}

func (t *tcp) disconnect() {
	if t.isConnected {
		t.isConnected = false

		if t.conn != nil {
			t.conn.Close()
			t.conn = nil
		}

		if t.logger != nil {
			t.logger.Printf("Disconnected from %s", t.host)
		}
	}
}

func (t *tcp) read(conn net.Conn) ([]byte, error) {
	rxBuffer := make([]byte, maxLength)

	if conn == nil {
		return nil, fmt.Errorf("connection is nil")
	}

	rxLen, err := conn.Read(rxBuffer)
	if err != nil {
		var netErr net.Error
		if !errors.As(err, &netErr) || !netErr.Timeout() {
			return nil, fmt.Errorf("read error: %w", err)
		}
	}

	if t.logger != nil {
		t.logger.Printf("RX (%s): %s", t.host, encodeHexString(rxBuffer[:rxLen]))
	}

	return rxBuffer[:rxLen], nil
}

func encodeHexString(b []byte) string {
	return strings.ToUpper(hex.EncodeToString(b))
}
