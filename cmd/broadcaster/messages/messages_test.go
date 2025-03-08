package messages

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytesToInt(t *testing.T) {
	byteArray := []byte(" 16")
	got := BytesToInt(byteArray)
	expected := 16

	if got != expected {
		t.Errorf("got %d, expected %d", got, expected)
	}
}

func TestDecodeMsg(t *testing.T) {
	// msg := MsgBuf{Msg: bytes.NewBuffer([]byte("<  1  5Heya!>")), MsgComplete: true}
	var msg bytes.Buffer
	msg.WriteString("<  1  5Heya!>")

	encodedMsg := MsgBuf{Msg: msg, MsgComplete: true}
	version := uint8(1)
	got, _ := encodedMsg.DecodeMsg(version)

	expected := "Heya!"

	if got != expected {
		t.Errorf("got %s, expected %s", got, expected)
	}
}

func TestIntToBytes(t *testing.T) {
	got := IntToBytes(1, 3)
	expected := []byte{0, 0, 1}

	assert.ElementsMatch(t, got, expected, fmt.Sprintf("got %s, expected %s", got, expected))
}

func TestEncodeMsg(t *testing.T) {
	outgoingMsg := MsgBuf{MsgComplete: false}
	version := uint8(1)
	outgoingMsg.EncodeMsg("Hello, mom", version)
	got := outgoingMsg.Msg.String()

	var msg bytes.Buffer
	msg.WriteString("<  1 10Hello, mom>")
	expectedMsg := MsgBuf{Msg: msg, MsgComplete: true}

	expected := expectedMsg.Msg.String()

	if got != expected {
		t.Errorf("got %s, expected %s", got, expected)
	}
}
