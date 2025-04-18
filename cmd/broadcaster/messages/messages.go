package messages

import (
	"bytes"
	"errors"
	"fmt"
	"math"
)

type MsgBuf struct {
	Msg         bytes.Buffer
	MsgComplete bool
}

func BytesToInt(byteArray []byte) int {
	var num float64 = 0
	for i, b := range byteArray {
		if b < 48 || b > 57 {
			continue
		}

		num = num + float64(b-48)*math.Pow10(len(byteArray)-(i+1))
	}

	return int(num)
}

// function to convert an integer to a byte array of a given length
func IntToBytes(num int, length int) []byte {
	byteArray := make([]byte, length)

	for i := length - 1; i >= 0; i-- {
		b := byte((num % 10) + 48)
		if b == 48 && i != length-1 {
			byteArray[i] = 32
		} else {
			byteArray[i] = b
		}
		num = num / 10
	}

	return byteArray
}

func (m *MsgBuf) Reset() {
	m.Msg.Reset()
	m.MsgComplete = false
}

func (m *MsgBuf) DecodeMsg(version uint8) (string, error) {
	if !m.MsgComplete {
		return "", errors.New("Message incomplete")
	}

	msgBytes := m.Msg.Bytes()
	if (msgBytes[0] != 60) || (msgBytes[len(msgBytes)-1] != 62) {
		errMsg := fmt.Sprintf("Invalid message format: '%s'", string(msgBytes))
		return "", errors.New(errMsg)
	}

	version_start := 1
	version_bytes := 3

	if BytesToInt(msgBytes[version_start:version_start+version_bytes]) != int(version) {
		errMsg := fmt.Sprintf(
			"Invalid message version: '%s'. Expected %v, Got %d",
			string(msgBytes),
			msgBytes[version_start:version_start+version_bytes],
			version,
		)
		return "", errors.New(errMsg)
	}

	msg_len_start := version_start + version_bytes
	msg_len_bytes := 3
	msg_start := msg_len_start + msg_len_bytes
	msg_bytes := BytesToInt(msgBytes[msg_len_start : msg_len_start+msg_len_bytes])

	var out string
	for i := msg_start; i < msg_bytes+msg_start; i++ {
		if i >= len(msgBytes) {
			return "", errors.New("Expected a bigger message. Something's wrong with the encoding")
		}

		out = out + string(msgBytes[i])
	}

	// gotta error catch here if the message ends up being smaller than expected

	return out, nil
}

func (m *MsgBuf) EncodeMsg(msg string, version uint8) error {
	if m.MsgComplete {
		return errors.New("Outgoing message already complete.")
	}

	l := len(msg)

	m.Msg.Write([]byte("<"))
	m.Msg.Write(IntToBytes(int(version), 3))
	m.Msg.Write(IntToBytes(l, 3))
	m.Msg.Write([]byte(msg))
	m.Msg.Write([]byte(">"))
	m.MsgComplete = true

	return nil
}
