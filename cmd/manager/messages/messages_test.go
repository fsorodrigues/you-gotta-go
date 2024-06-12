package messages

import (
	"bytes"
	"testing"
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
	version := 1
	got, _ := DecodeMsg(encodedMsg, version)
	expected := "Heya!"

	if got != expected {
		t.Errorf("got %s, expected %s", got, expected)
	}
}
