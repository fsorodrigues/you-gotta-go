package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"you-gotta-go/cmd/broadcaster/messages"
)

type Comms struct {
	API_KEY            string
	BASE_URL           string
	TCP_PORT           string
	HoldingBuffer      []byte
	IncomingMsg        messages.MsgBuf
	BytesRead          uint8
	BytesAvailable     uint8
	Reading            bool
	MsgEncodingVersion uint8
	ConnectedDevices   []Device
}

func (c *Comms) clearIncomingMsg() {
	c.IncomingMsg.Msg.Reset()
	c.IncomingMsg.MsgComplete = false
}

func (c *Comms) walkBuffer(conn io.ReadWriteCloser) {
	b := make([]byte, c.BytesAvailable)
	n := 0
	startByte := 0
	stopByte := 0

	for i := c.BytesRead; i < c.BytesRead+c.BytesAvailable; i++ {
		stopByte++
		c.BytesRead++
		c.BytesAvailable--

		if c.HoldingBuffer[i] == 60 {
			c.clearIncomingMsg()
			b[n] = c.HoldingBuffer[i]
			startByte = n
			stopByte = n + 1
			n++
			continue
		}

		if c.HoldingBuffer[i] == 62 {
			b[n] = c.HoldingBuffer[i]
			c.Reading = false
			c.IncomingMsg.MsgComplete = true
			break
		}

		if !c.IncomingMsg.MsgComplete {
			b[n] = c.HoldingBuffer[i]
			n++
		}
	}

	c.IncomingMsg.Msg.Write(b[startByte:stopByte])
	if c.Reading {
		c.readToBuffer(conn)
	}
}

func (c *Comms) readToBuffer(conn io.ReadWriteCloser) {
	n, err := conn.Read(c.HoldingBuffer)
	if err != nil {
		log.Fatalln("Error reading from io.ReadWriteCloser", err)
	}

	c.BytesRead = 0
	c.BytesAvailable = uint8(n)
}

func (c *Comms) ReadFromConnection(conn io.ReadWriteCloser) error {
	c.Reading = true

	for c.Reading {
		c.walkBuffer(conn)
		if c.BytesAvailable == 0 {
			c.Reading = false
			break
		}
	}

	if !c.IncomingMsg.MsgComplete {
		errMsg := fmt.Sprintf(
			"Error reading message. Message incomplete. Expected %d, got %d bytes",
			c.BytesRead+c.BytesAvailable,
			c.IncomingMsg.Msg.Len(),
		)

		return errors.New(errMsg)
	}

	return nil
}

func (c *Comms) handleConnection(dev Device) {
	// try reading from connection
	errReadFromConnection := c.ReadFromConnection(dev.Connection)
	if errReadFromConnection != nil {
		log.Fatalln(errReadFromConnection)
	}

	// decode message, parse it
	msg, errDecodeMsg := c.IncomingMsg.DecodeMsg(c.MsgEncodingVersion)
	if errDecodeMsg != nil {
		log.Fatalln(errDecodeMsg)
	}

	m, startSignal := strings.CutPrefix(msg, "start")

	if startSignal {
	}

	dev.Connection.Close()
}

func (c *Comms) Listen() {
	ln, errListenTCP := net.Listen("tcp", fmt.Sprintf(":%s", c.TCP_PORT))
	if errListenTCP != nil {
		log.Fatalln("Can't open TCP connection", errListenTCP)
	}
	defer ln.Close()

	for {
		conn, errAcceptConnection := ln.Accept()
		if errAcceptConnection != nil {
			// handle error
			log.Fatalln("Error accepting incoming TCP connection", errAcceptConnection)
		}
		// create device, assign connection, and append to list of connected devices
		dev := Device{
			Connection: conn,
		}
		c.ConnectedDevices = append(c.ConnectedDevices, dev)

		// start routine
		go c.handleConnection(dev)
	}
}
