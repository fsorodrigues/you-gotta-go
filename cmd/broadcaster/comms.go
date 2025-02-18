package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"you-gotta-go/cmd/broadcaster/messages"
)

type Comms struct {
	API_KEY        string
	BASE_URL       string
	TCP_PORT       string
	HoldingBuffer  []byte
	IncomingMsg    messages.MsgBuf
	BytesRead      int
	BytesAvailable int
	Reading        bool
}

func (c *Comms) clearIncomingMsg() {
	c.IncomingMsg.Msg.Reset()
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
			c.IncomingMsg.MsgComplete = false
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
		log.Fatal(err)
	}
	c.BytesRead = 0
	c.BytesAvailable = n
}

func (c *Comms) ReadFromConnection(conn io.ReadWriteCloser) {
	c.Reading = true

	for c.Reading {
		c.walkBuffer(conn)
		if c.BytesAvailable == 0 {
			c.Reading = false
			break
		}
	}
}

func (c *Comms) handleConnection(conn io.ReadWriteCloser) {
	// try reading from connection
	c.ReadFromConnection(conn)

	// after reading kick off routine specified in incoming message
	fmt.Println(string(c.IncomingMsg.Msg.Bytes()[:]))

	conn.Close()
}

func (c *Comms) Listen() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", c.TCP_PORT))
	if err != nil {
		log.Fatalln("Can't open TCP connection", err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			log.Fatalln("Error accepting incoming TCP connection", err)
		}
		go c.handleConnection(conn)
	}
}
