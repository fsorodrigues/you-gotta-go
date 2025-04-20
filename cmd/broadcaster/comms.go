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
	ErrChan          chan error
}

type CommsError struct {
	Comms Comms
	Err   error
}

func (d CommsError) Error() string {
	return fmt.Sprintf("Error at the Comms level. %v", d.Err.Error())
}

func (c *Comms) handleConnection(dev ConnectedDevice) {
	defer func() {
		log.Printf("Killing device: %s\n", dev.Id)
		delete(c.ConnectedDevices, dev.Id)
		dev.KillDevice()
	}()

	initMsg, startSignal, errRead := dev.readForSignal("start")
	if errRead != nil {
		c.ErrChan <- errRead
		return
	}

	if !startSignal {
		return
	}

	log.Println("Configuring device")
	// assign stop/service values to device
	split := strings.Split(initMsg, "|")
	dev.TargetStop = split[0]
	dev.TargetService = split[1]

	for dev.StatusAlive {
		// after reading kick off routine specified in incoming message
		log.Println("Scraping...")
		raw_data, scrapingErr := scraper.Scrape(c.BASE_URL, dev.TargetStop, c.API_KEY)
		if scrapingErr != nil {
			c.ErrChan <- scrapingErr
		}

		data, unmarshallErr := parser.Unmarshal([]byte(raw_data))
		if unmarshallErr != nil {
			c.ErrChan <- scrapingErr
		}

		payload, parserErr := parser.Parse(data, dev.TargetService)
		if parserErr != nil {
			c.ErrChan <- parserErr
		}

		log.Printf("Writing to device: %s\n", dev.Id)
		encodingErr := dev.OutgoingMsg.EncodeMsg(*payload, dev.ENCODING_VERSION)
		if encodingErr != nil {
			c.ErrChan <- encodingErr
			dev.StatusAlive = false
			break
		}

		_, writeErr := dev.Connection.Write(dev.OutgoingMsg.Msg.Bytes())
		if writeErr != nil {
			c.ErrChan <- DeviceError{
				DeviceID: dev.Id,
				Err:      fmt.Errorf("Error writing to device: %w", writeErr),
			}
			dev.StatusAlive = false
			break
		}
		dev.OutgoingMsg.Reset()

		// listen for kill signal
		_, killSignal, killErr := dev.readForSignal("kill")
		if killErr != nil {
			log.Println("Didn't receive kill signal. Let's keep riding the bus.")
		}
		if killSignal {
			dev.StatusAlive = false
			break
		}
	}
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

func (c *Comms) FindUSBDevices() ([]serial.Port, []string, error) {
	items := make([]serial.Port, len(c.USB_DEVICES))
	names := make([]string, len(c.USB_DEVICES))
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, nil, err
	}

	n_found := 0
	for _, port := range ports {
		isAcceptedDev := slices.Contains(c.USB_DEVICES, port)

		_, found := c.ConnectedDevices[port]
		if isAcceptedDev && n_found < len(c.USB_DEVICES) && !found {
			log.Printf("Found usb device: %s\n", port)
			conn, err := openUSBConnection(port, c.BAUD_RATE)
			if err != nil {
				return nil, nil, err
			}

			items[n_found] = conn
			names[n_found] = port
			n_found++
		}
	}

	return items, names, nil
}

func (c *Comms) ListenForUSB() ([]ConnectedDevice, error) {
	usbDevices, usbNames, usbError := c.FindUSBDevices()
	if usbError != nil {
		return nil, usbError
	}
	items := make([]ConnectedDevice, len(c.USB_DEVICES))

	for i, usb := range usbDevices {
		if usb == nil {
			break
		}

		item := ConnectedDevice{
			Id: usbNames[i],
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

		_, isReady, errRead := item.readForSignal("ready")
		if errRead != nil {
			return items, errRead
		}

		if isReady {
			items[i] = item
		}
	}
	return items, nil
}

func (c *Comms) Listen() {
	// Create error handling goroutine
	go func() {
		for err := range c.ErrChan {
			log.Printf("%v\n", err.Error())
		}
	}()

	go func() {
		for {
			usbDevices, errListenUSB := c.ListenForUSB()
			if errListenUSB != nil {
				c.ErrChan <- CommsError{
					Comms: *c,
					Err:   fmt.Errorf("Error attempting to listen for USB devices: %w", errListenUSB),
				}
			}

			for _, usbDev := range usbDevices {
				if usbDev.Id == "" {
					continue
				}

				c.ConnectedDevices[usbDev.Id] = usbDev
				defer usbDev.Connection.Close()
				go c.handleConnection(usbDev)
			}

			// Wait before scanning again
			time.Sleep(5 * time.Second)
		}
	}()

	ln, errListenTCP := c.ListenForTCP()
	if errListenTCP != nil {
		c.ErrChan <- CommsError{
			Comms: *c,
			Err:   fmt.Errorf("Error attempting to open port for TCP connections: %w", errListenTCP),
		}
	}
	defer ln.Close()

	for {
		tcpConn, errAcceptConnection := ln.Accept()
		if errAcceptConnection != nil {
			c.ErrChan <- CommsError{
				Comms: *c,
				Err:   fmt.Errorf("Error attempting to accept TCP connection: %w", errListenTCP),
			}
		}

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
