package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
	"you-gotta-go/cmd/broadcaster/messages"

	"go.bug.st/serial"
)

type ConnectedDevice struct {
	Id               string
	Type             string
	TargetStop       string
	TargetService    string
	Connection       CommsConnection
	StatusAlive      bool
	HoldingBuffer    []byte
	IncomingMsg      messages.MsgBuf
	OutgoingMsg      messages.MsgBuf
	BytesRead        uint8
	BytesAvailable   uint8
	Reading          bool
	ENCODING_VERSION uint8
}

type DeviceError struct {
	DeviceID string
	Err      error
}

func (d DeviceError) Error() string {
	return fmt.Sprintf("Error from Device '%s': %v", d.DeviceID, d.Err.Error())
}

func (dev *ConnectedDevice) HoldingBufferReset() {
	dev.BytesAvailable = 0
	dev.HoldingBuffer = make([]byte, 25)
}

func (dev *ConnectedDevice) walkBuffer(conn CommsConnection) error {
	b := make([]byte, dev.BytesAvailable)
	n := 0
	startByte := 0
	stopByte := 0

	for i := dev.BytesRead; i < dev.BytesRead+dev.BytesAvailable; i++ {
		stopByte++
		dev.BytesRead++
		dev.BytesAvailable--

		if dev.HoldingBuffer[i] == 60 {
			dev.IncomingMsg.Reset()
			b[n] = dev.HoldingBuffer[i]
			startByte = n
			stopByte = n + 1
			n++
			continue
		}

		if dev.HoldingBuffer[i] == 62 {
			b[n] = dev.HoldingBuffer[i]
			dev.Reading = false
			dev.IncomingMsg.MsgComplete = true
			dev.HoldingBufferReset()

			break
		}

		if !dev.IncomingMsg.MsgComplete {
			b[n] = dev.HoldingBuffer[i]
			n++
		}
	}

	dev.IncomingMsg.Msg.Write(b[startByte:stopByte])
	if !dev.Reading {
		return nil
	}

	errRead := dev.readToBuffer(conn)
	if errRead != nil {
		return errRead
	}

	return nil
}

func (d *ConnectedDevice) readToBuffer(conn CommsConnection) error {
	errSetTimeout := conn.SetReadTimeout()
	if errSetTimeout != nil {
		return errSetTimeout
	}

	n, errRead := conn.Read(d.HoldingBuffer)
	if errRead != nil {
		if errors.Is(errRead, os.ErrDeadlineExceeded) {
			return errors.New("Nothing to read from device. DeadlineExceeded")
		}
		return errRead
	}

	d.BytesRead = 0
	d.BytesAvailable = uint8(n)

	return nil
}

func (dev *ConnectedDevice) ReadFromConnection(conn CommsConnection) error {
	dev.Reading = true

	for dev.Reading {
		log.Printf("Reading status: %v Bytes available: %d\n", dev.Reading, dev.BytesAvailable)

		errBuffer := dev.walkBuffer(conn)
		if errBuffer != nil {
			return errBuffer
		}

		if dev.BytesAvailable == 0 {
			dev.Reading = false
			break
		}
	}

	expectedBytes := dev.BytesRead + dev.BytesAvailable
	if !dev.IncomingMsg.MsgComplete && expectedBytes != 0 {
		errMsg := fmt.Sprintf(
			"Error reading message. Message incomplete. Expected %d, got %d bytes",
			expectedBytes,
			dev.IncomingMsg.Msg.Len(),
		)

		return errors.New(errMsg)
	}

	return nil
}

func (dev *ConnectedDevice) readForSignal(signal string) (string, bool, error) {
	log.Printf("Reading. Waiting for signal: %s\n", signal)
	// try reading from connection
	errReadFromConnection := dev.ReadFromConnection(dev.Connection)
	if errReadFromConnection != nil {
		return "", false, errReadFromConnection
	}

	if dev.BytesRead > 0 {
		// decode message, parse it
		msg, errDecodeMsg := dev.IncomingMsg.DecodeMsg(dev.ENCODING_VERSION)
		if errDecodeMsg != nil {
			return "", false, errDecodeMsg
		}
		log.Printf("Received: %s\n", msg)

		cutMsg, found := strings.CutPrefix(msg, signal)
		return cutMsg, found, nil
	}

	return "", false, nil
}

func (dev *ConnectedDevice) KillDevice() {
	dev.Connection.Close()
}

type TCPConnection struct {
	Conn        net.Conn
	TimeoutTime time.Duration
}

func (c TCPConnection) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (c TCPConnection) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (c TCPConnection) Close() error {
	err := c.Conn.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c TCPConnection) SetReadTimeout() error {
	err := c.Conn.SetReadDeadline(time.Now().Add(c.TimeoutTime))
	if err != nil {
		return errors.New("Error setting SerialConnection read timeout")
	}
	return nil
}

type SerialConnection struct {
	Conn        serial.Port
	TimeoutTime time.Duration
}

func (c SerialConnection) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (c SerialConnection) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (c SerialConnection) Close() error {
	err := c.Conn.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c SerialConnection) SetReadTimeout() error {
	err := c.Conn.SetReadTimeout(c.TimeoutTime)
	if err != nil {
		return errors.New("Error setting SerialConnection read timeout")
	}
	return nil
}

type CommsConnection interface {
	io.Reader
	io.Writer
	io.Closer
	SetReadTimeout() error
}
