package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"you-gotta-go/cmd/manager/messages"

	"github.com/joho/godotenv"
	"go.bug.st/serial"
)

type Controller struct {
	USB_DEVICE     string
	DeviceReady    bool
	DisplayReady   bool
	SwitchOn       bool
	HoldingBuffer  []byte
	BytesAvailable int
	BytesRead      int
	LastMsg        messages.MsgBuf
	LastMsgBytes   int
	Reading        bool
}

type IncomingMsg struct {
	value    string
	callback func()
}

type SerialReader interface {
	openConn() serial.Port
	read() int
	listenForMsg(conn serial.Port, msg IncomingMsg)
}

func (c *Controller) ToggleDeviceStatus() {
	c.DeviceReady = !c.DeviceReady
}

func (c *Controller) ToggleDisplayStatus() {
	c.DisplayReady = !c.DisplayReady
}

func (c *Controller) openConn() serial.Port {
	usbPort, err := serial.Open(c.USB_DEVICE, &serial.Mode{BaudRate: 9600})
	if err != nil {
		log.Fatal(err)
	}
	return usbPort
}

func (c *Controller) clearLastMsg() {
	c.LastMsg.Msg.Reset()
}

func (c *Controller) readToBuffer(conn serial.Port) {
	n, err := conn.Read(c.HoldingBuffer)
	if err != nil {
		log.Fatal(err)
	}
	c.BytesRead = 0
	c.BytesAvailable = n
}

func (c *Controller) walkBuffer(conn serial.Port) {
	b := make([]byte, c.BytesAvailable)
	n := 0
	startByte := 0
	stopByte := 0

	for i := c.BytesRead; i < c.BytesRead+c.BytesAvailable; i++ {
		stopByte++
		c.BytesRead++
		c.BytesAvailable--

		if c.HoldingBuffer[i] == 60 {
			c.clearLastMsg()
			c.LastMsg.MsgComplete = false
			b[n] = c.HoldingBuffer[i]
			startByte = n
			stopByte = n + 1
			n++
			continue
		}

		if c.HoldingBuffer[i] == 62 {
			b[n] = c.HoldingBuffer[i]
			c.Reading = false
			c.LastMsg.MsgComplete = true
			break
		}

		if !c.LastMsg.MsgComplete {
			b[n] = c.HoldingBuffer[i]
			n++
		}
	}

	c.LastMsg.Msg.Write(b[startByte:stopByte])
	if c.Reading {
		c.readToBuffer(conn)
	}
}

func (c *Controller) read(conn serial.Port) {
	c.Reading = true

	for c.Reading {
		c.walkBuffer(conn)
		if c.BytesAvailable == 0 {
			c.Reading = false
			break
		}
	}
}

func (c *Controller) listenForMsg(conn serial.Port, msg IncomingMsg) {
	c.read(conn)
	m, e := messages.DecodeMsg(c.LastMsg, 1)
	if e != nil {
		log.Fatal(e)
	}

	if m == msg.value {
		msg.callback()
	}
}

func (c *Controller) setup(conn serial.Port, msgs ...IncomingMsg) {
	for _, msg := range msgs {
		c.listenForMsg(conn, msg)
	}
}

		}
	}
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	// load env vars
	USB_DEVICE := os.Getenv("USB_DEVICE")

	controller := Controller{
		USB_DEVICE:     USB_DEVICE,
		DeviceReady:    false,
		DisplayReady:   false,
		SwitchOn:       false,
		HoldingBuffer:  make([]byte, 25),
		BytesAvailable: 0,
		BytesRead:      0,
		LastMsg:        messages.MsgBuf{Msg: bytes.Buffer{}, MsgComplete: false},
		LastMsgBytes:   0,
		Reading:        false,
	}

	port := controller.openConn()
	controller.setup(
		port,
		IncomingMsg{"Arduino is ready", controller.ToggleDeviceStatus},
		IncomingMsg{"LCD is ready", controller.ToggleDisplayStatus},
	)

	if controller.DeviceReady && controller.DisplayReady {
		fmt.Println("Arduino ready to rock'n'roll")
	}
}
