package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
	"you-gotta-go/cmd/broadcaster/messages"
	"you-gotta-go/cmd/parser"
	"you-gotta-go/cmd/scraper"

	"github.com/google/uuid"
)

type Comms struct {
	API_KEY          string
	BASE_URL         string
	TCP_PORT         string
	ConnectedDevices map[string]ConnectedDevice
}

func (c *Comms) handleConnection(dev ConnectedDevice) {
	initMsg, startSignal := dev.readForSignal("start")

	if startSignal { // start loop
		log.Println("Configuring device")
		// assign stop/service values to device
		split := strings.Split(initMsg, "|")
		dev.TargetStop = split[0]
		dev.TargetService = split[1]

		for dev.StatusAlive {
			log.Println("Loop iter")

			// after reading kick off routine specified in incoming message
			log.Println("Scraping...")
			data := scraper.Scrape(c.BASE_URL, dev.TargetStop, c.API_KEY)
			payload := parser.Parse(
				parser.Unmarshal([]byte(data)),
				dev.TargetService,
			)
			log.Print(*payload)

			log.Println(fmt.Sprintf("Writing to device: %s", dev.Id))
			dev.OutgoingMsg.Msg.WriteString(*payload)
			_, writeErr := dev.Connection.Write(dev.OutgoingMsg.Msg.Bytes())
			if writeErr != nil {
				log.Fatalln("Error writing to device", writeErr)
			}
			dev.OutgoingMsg.Msg.Reset()

			// listen for kill signal
			_, killSignal := dev.readForSignal("kill")
			if killSignal {
				dev.StatusAlive = false
				break
			}
		}
	}

	log.Println(fmt.Sprintf("Killing device: %s", dev.Id))
	delete(c.ConnectedDevices, dev.Id)
	dev.KillDevice()
}

func (c *Comms) Listen(ENCODING_VERSION int) {
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
		defer conn.Close()

		// create device, assign connection, and append to list of connected devices
		dev := ConnectedDevice{
			Id: uuid.New().String(),
			Connection: TCPConnection{
				Conn:        conn,
				TimeoutTime: time.Second * 30,
			},
			Type:               "tcp",
			StatusAlive:        true,
			HoldingBuffer:      make([]byte, 25),
			IncomingMsg:        messages.MsgBuf{Msg: bytes.Buffer{}, MsgComplete: false},
			OutgoingMsg:        messages.MsgBuf{Msg: bytes.Buffer{}, MsgComplete: false},
			BytesAvailable:     0,
			BytesRead:          0,
			Reading:            false,
			MsgEncodingVersion: uint8(ENCODING_VERSION),
		}
		c.ConnectedDevices[dev.Id] = dev

		// start routine
		go c.handleConnection(dev)
	}
}
