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

type Device struct {
	Id                 string
	Type               string
	TargetStop         string
	TargetService      string
	Connection         CommsConnection
	StatusAlive        bool
	HoldingBuffer      []byte
	IncomingMsg        messages.MsgBuf
	OutgoingMsg        messages.MsgBuf
	BytesRead          uint8
	BytesAvailable     uint8
	Reading            bool
	MsgEncodingVersion uint8
}

func (dev *Device) clearIncomingMsg() {
	dev.IncomingMsg.Msg.Reset()
	dev.IncomingMsg.MsgComplete = false
}

func (dev *Device) HoldingBufferReset() {
	dev.BytesAvailable = 0
	dev.HoldingBuffer = make([]byte, 25)
}

func (dev *Device) walkBuffer(conn CommsConnection) {
	b := make([]byte, dev.BytesAvailable)
	n := 0
	startByte := 0
	stopByte := 0

	for i := dev.BytesRead; i < dev.BytesRead+dev.BytesAvailable; i++ {
		stopByte++
		dev.BytesRead++
		dev.BytesAvailable--

		if dev.HoldingBuffer[i] == 60 {
			dev.clearIncomingMsg()
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
	if dev.Reading {
		dev.readToBuffer(conn)
	}
}

func (d *Device) readToBuffer(conn CommsConnection) {
	conn.SetReadTimeout()
	n, err := conn.Read(d.HoldingBuffer)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			log.Println("Nothing to read from device.")
			return
		}
		log.Fatalln("Error reading from io.ReadWriteCloser", err)
	}

	d.BytesRead = 0
	d.BytesAvailable = uint8(n)
}

func (dev *Device) ReadFromConnection(conn CommsConnection) error {
	dev.Reading = true

	for dev.Reading {
		log.Println(fmt.Sprintf("Reading status: %v Bytes available: %d", dev.Reading, dev.BytesAvailable))
		dev.walkBuffer(conn)
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

func (dev *Device) readForSignal(signal string) (string, bool) {
	log.Println(fmt.Sprintf("Reading. Waiting for signal: %s", signal))
	// try reading from connection
	errReadFromConnection := dev.ReadFromConnection(dev.Connection)
	if errReadFromConnection != nil {
		log.Fatalln(errReadFromConnection)
	}

	if dev.BytesRead > 0 {
		// decode message, parse it
		msg, errDecodeMsg := dev.IncomingMsg.DecodeMsg(dev.MsgEncodingVersion)
		if errDecodeMsg != nil {
			log.Fatalln(errDecodeMsg)
		}
		log.Println(fmt.Sprintf("Received: %s", msg))

		return strings.CutPrefix(msg, signal)
	}

	return "", false
}

func (dev *Device) KillDevice() {
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

func (c TCPConnection) SetReadTimeout() {
	c.Conn.SetReadDeadline(time.Now().Add(c.TimeoutTime))
}

type SerialConnection struct {
	Conn        serial.Port
	TimeoutTime time.Duration
}

func (c SerialConnection) Close() error {
	err := c.Conn.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c SerialConnection) SetReadTimeout() {
	c.Conn.SetReadTimeout(c.TimeoutTime)
}

type CommsConnection interface {
	io.Reader
	io.Writer
	io.Closer
	SetReadTimeout()
}
