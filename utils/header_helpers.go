package utils

import (
	"bytes"
	"log"
)

func GetHeaderType(header []byte) int {
	switch string(header) {
	case "inf":
		return HdrInfoVal
	case "ack":
		return HdrAckVal
	case "err":
		return HdrErrVal
	case "res":
		return HdrResVal
	default:
		return HdrUnknownVal
	}
}

func ParseHeader(header []byte) (int, error) {
	if len(header) != 11 {
		return HdrUnknownVal, ErrInvalidHeader
	}

	split := bytes.Split(header, []byte("+"))
	if !bytes.Contains(split[0], []byte("sprlmnl")) || len(split[1]) != 3 {
		return HdrUnknownVal, ErrInvalidHeader
	}

	log.Println(string(split[1]))
	headerType := GetHeaderType(split[1])
	if headerType == HdrUnknownVal {
		return HdrUnknownVal, ErrUnknownHeader
	}

	return headerType, nil
}

func ParseIncomingMsg(msg []byte) (int, []byte, error) {
	header, message, ok := bytes.Cut(msg, []byte(":"))
	if !ok {
		return HdrUnknownVal, nil, ErrInvalidHeader
	}

	headerType, err := ParseHeader(header)
	if err != nil {
		return headerType, nil, err
	}

	return headerType, message, nil
}
