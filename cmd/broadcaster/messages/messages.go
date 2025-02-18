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

func DecodeMsg(msg MsgBuf, version int) (string, error) {
	if !msg.MsgComplete {
		return "", errors.New("Message incomplete")
	}

	msgBytes := msg.Msg.Bytes()
	if (msgBytes[0] != 60) || (msgBytes[len(msgBytes)-1] != 62) {
		errMsg := fmt.Sprintf("Invalid message format: '%s'", string(msgBytes))
		return "", errors.New(errMsg)
	}

	version_start := 1
	version_bytes := 3

	if BytesToInt(msgBytes[version_start:version_start+version_bytes]) != version {
		fmt.Println("version:", msgBytes)
		errMsg := fmt.Sprintf("Invalid message version: '%s'", string(msgBytes))
		return "", errors.New(errMsg)
	}

	msg_len_start := version_start + version_bytes
	msg_len_bytes := 3
	msg_start := msg_len_start + msg_len_bytes
	msg_bytes := BytesToInt(msgBytes[msg_len_start : msg_len_start+msg_len_bytes])

	var out string
	for i := msg_start; i < msg_bytes+msg_start; i++ {
		out = out + string(msgBytes[i])
	}
	return out, nil
}
