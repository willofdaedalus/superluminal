package utils

import (
	"bytes"
	"fmt"
	"strconv"
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

// func ParseHeader(header []byte) (int, error) {
// 	if len(header) != 11 {
// 		return HdrUnknownVal, ErrInvalidHeader
// 	}

// 	split := bytes.Split(header, []byte("+"))
// 	if !bytes.Contains(split[0], []byte("sprlmnl")) || len(split[1]) != 3 {
// 		return HdrUnknownVal, ErrInvalidHeader
// 	}

// 	log.Println(string(split[1]))
// 	headerType := GetHeaderType(split[1])
// 	if headerType == HdrUnknownVal {
// 		return HdrUnknownVal, ErrUnknownHeader
// 	}

// 	return headerType, nil
// }

// typical header looks like this
// 300+301
// header above is an ack with a self report message
func ParseHeader(header []byte) (int, int, error) {
	if len(header) != 7 {
		return HdrUnknownVal, -1, ErrInvalidHeader
	}

	split := bytes.Split(header, []byte("+"))
	hdrVal, err := strconv.Atoi(string(split[0]))
	if err != nil {
		return HdrUnknownVal, -1, ErrInvalidHeader
	}
	hdrReq, err := strconv.Atoi(string(split[1]))
	if err != nil {
		return HdrUnknownVal, -1, ErrInvalidHeader
	}

	// check for cases where the request is not
	// from the correct header family
	if hdrReq < hdrVal || hdrReq > (hdrReq+100) {
		return HdrUnknownVal, -1, fmt.Errorf("sprlmnl: msg header and sub do not match; resend")
	}

	return hdrVal, hdrReq, nil
}

func ParseIncomingMsg(msg []byte) (int, int, []byte, error) {
	header, message, ok := bytes.Cut(msg, []byte(":"))
	if !ok {
		return HdrUnknownVal, -1, nil, ErrInvalidHeader
	}

	hdrVal, hdrMsg, err := ParseHeader(header)
	if err != nil {
		return HdrUnknownVal, -1, nil, err
	}

	return hdrVal, hdrMsg, message, nil
}
