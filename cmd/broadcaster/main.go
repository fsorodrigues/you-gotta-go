package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	// "you-gotta-go/cmd/scraper"

	"github.com/joho/godotenv"
)

type Broadcaster interface {
	Listen()
	acceptConnection()
	handleConnection()
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	API_KEY := os.Getenv("API_KEY")
	BASE_URL := os.Getenv("BASE_URL")
	TCP_PORT := os.Getenv("TCP_PORT")
	ENCODING_VERSION, err := strconv.Atoi(os.Getenv("ENCODING_VERSION"))
	if err != nil {
		log.Fatalln("Error converting ENCODING_VERSION. Is it a valid int?")
	}
	USB_DEVICES := strings.Split(os.Getenv("USB_DEVICES"), "|")
	BAUD_RATE, err := strconv.Atoi(os.Getenv("BAUD_RATE"))
	if err != nil {
		log.Fatalln("Error converting BAUD_RATE. Is it a valid int?")
	}

	app := Comms{
		API_KEY:          API_KEY,
		BASE_URL:         BASE_URL,
		TCP_PORT:         TCP_PORT,
		ENCODING_VERSION: uint8(ENCODING_VERSION),
		USB_DEVICES:      USB_DEVICES,
		BAUD_RATE:        BAUD_RATE,
		ConnectedDevices: make(map[string]ConnectedDevice),
		ErrChan:          make(chan DeviceError, 100), // Buffered channel to avoid blocking
	}

	app.Listen()
}
