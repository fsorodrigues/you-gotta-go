package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"slices"
	"strings"
	"time"
	"you-gotta-go/cmd/broadcaster/messages"
	"you-gotta-go/cmd/parser"
	"you-gotta-go/cmd/scraper"

	"github.com/google/uuid"
	"go.bug.st/serial"
)

type Comms struct {
	API_KEY          string
	BASE_URL         string
	TCP_PORT         string
	ENCODING_VERSION uint8
	USB_DEVICES      []string
	BAUD_RATE        int
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

			log.Printf("Writing to device: %s\n", dev.Id)
			encodingErr := dev.OutgoingMsg.EncodeMsg(*payload, dev.ENCODING_VERSION)
			if encodingErr != nil {
				log.Fatalln("Error encoding message", encodingErr)
			}

			_, writeErr := dev.Connection.Write(dev.OutgoingMsg.Msg.Bytes())
			if writeErr != nil {
				log.Fatalln("Error writing to device", writeErr)
			}
			dev.OutgoingMsg.Reset()

			// listen for kill signal
			_, killSignal := dev.readForSignal("kill")
			if killSignal {
				dev.StatusAlive = false
				break
			}
		}
	}

	log.Printf("Killing device: %s\n", dev.Id)
	delete(c.ConnectedDevices, dev.Id)
	dev.KillDevice()
}

func (c *Comms) ListenForTCP() (net.Listener, error) {
	ln, errListenTCP := net.Listen("tcp", fmt.Sprintf(":%s", c.TCP_PORT))
	if errListenTCP != nil {
		return nil, errListenTCP
	}

	return ln, nil
}

func openUSBConnection(port string, BAUD_RATE int) (serial.Port, error) {
	usbPort, portErr := serial.Open(
		port,
		&serial.Mode{BaudRate: BAUD_RATE},
	)
	if portErr != nil {
		return nil, portErr
	}

	return usbPort, nil
}

func (c *Comms) FindUSBDevices() ([]serial.Port, error) {
	items := make([]serial.Port, len(c.USB_DEVICES))
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}

	n_found := 0
	for _, port := range ports {
		isAcceptedDev := slices.Contains(c.USB_DEVICES, port)
		if isAcceptedDev && n_found < len(c.USB_DEVICES) {
			log.Printf("Found usb device: %s\n", port)
			conn, err := openUSBConnection(port, c.BAUD_RATE)
			if err != nil {
				return nil, err
			}

			items[n_found] = conn
			n_found++
		}
	}

	return items, nil
}

func (c *Comms) ListenForUSB() ([]ConnectedDevice, error) {
	usbDevices, usbError := c.FindUSBDevices()
	if usbError != nil {
		log.Fatalln(usbError)
	}
	items := make([]ConnectedDevice, len(c.USB_DEVICES))

	for i, usb := range usbDevices {
		if usb == nil {
			break
		}

		item := ConnectedDevice{
			Id: uuid.New().String(),
			Connection: SerialConnection{
				Conn:        usb,
				TimeoutTime: time.Second * 30,
			},
			Type:             "usb",
			StatusAlive:      true,
			HoldingBuffer:    make([]byte, 25),
			IncomingMsg:      messages.MsgBuf{Msg: bytes.Buffer{}, MsgComplete: false},
			OutgoingMsg:      messages.MsgBuf{Msg: bytes.Buffer{}, MsgComplete: false},
			BytesAvailable:   0,
			BytesRead:        0,
			Reading:          false,
			ENCODING_VERSION: c.ENCODING_VERSION,
		}

		_, isReady := item.readForSignal("ready")
		if isReady {
			items[i] = item
		}
	}
	return items, nil
}

func (c *Comms) Listen() {
	usbDevices, errListenUSB := c.ListenForUSB()
	if errListenUSB != nil {
		log.Fatalln("Can't open USB connection", errListenUSB)
	}

	for _, usbDev := range usbDevices {
		if usbDev.Id == "" {
			break
		}
		c.ConnectedDevices[usbDev.Id] = usbDev
		defer usbDev.Connection.Close()
		go c.handleConnection(usbDev)
	}

	ln, errListenTCP := c.ListenForTCP()
	if errListenTCP != nil {
		log.Fatalln("Can't open TCP connection", errListenTCP)
	}
	defer ln.Close()

	for {
		tcpConn, errAcceptConnection := ln.Accept()
		if errAcceptConnection != nil {
			// handle error
			log.Fatalln("Error accepting incoming TCP connection", errAcceptConnection)
		}
		// defer tcpConn.Close()

		// create device, assign connection, and append to list of connected devices
		dev := ConnectedDevice{
			Id: uuid.New().String(),
			Connection: TCPConnection{
				Conn:        tcpConn,
				TimeoutTime: time.Second * 30,
			},
			Type:             "tcp",
			StatusAlive:      true,
			HoldingBuffer:    make([]byte, 25),
			IncomingMsg:      messages.MsgBuf{Msg: bytes.Buffer{}, MsgComplete: false},
			OutgoingMsg:      messages.MsgBuf{Msg: bytes.Buffer{}, MsgComplete: false},
			BytesAvailable:   0,
			BytesRead:        0,
			Reading:          false,
			ENCODING_VERSION: c.ENCODING_VERSION,
		}
		c.ConnectedDevices[dev.Id] = dev

		// start routine
		go c.handleConnection(dev)
	}
}
