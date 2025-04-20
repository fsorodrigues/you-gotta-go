package main

import (
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Broadcaster interface {
	Listen()
	acceptConnection()
	handleConnection()
}

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found")
	}
}

func main() {
	API_KEY := os.Getenv("API_KEY")
	BASE_URL := os.Getenv("BASE_URL")
	TCP_PORT := os.Getenv("TCP_PORT")
	ENCODING_VERSION, err := strconv.Atoi(os.Getenv("ENCODING_VERSION"))
	if err != nil {
		slog.Error("Error converting ENCODING_VERSION. Is it a valid int?\n")
		log.Fatalln("Crashing program.")
	}
	USB_DEVICES := strings.Split(os.Getenv("USB_DEVICES"), "|")
	BAUD_RATE, err := strconv.Atoi(os.Getenv("BAUD_RATE"))
	if err != nil {
		slog.Error("Error converting BAUD_RATE. Is it a valid int?")
		log.Fatalln("Crashing program.")
	}

	app := Comms{
		API_KEY:          API_KEY,
		BASE_URL:         BASE_URL,
		TCP_PORT:         TCP_PORT,
		ENCODING_VERSION: uint8(ENCODING_VERSION),
		USB_DEVICES:      USB_DEVICES,
		BAUD_RATE:        BAUD_RATE,
		ConnectedDevices: make(map[string]ConnectedDevice),
		ErrChan:          make(chan error, 100), // Buffered channel to avoid blocking
	}

	app.Listen()
}
