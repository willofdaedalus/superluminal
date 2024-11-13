package base

import (
	"fmt"
	"hash/crc32"
	"time"
	"willofdaedalus/superluminal/internal/payload/common"
	err1 "willofdaedalus/superluminal/internal/payload/error"
	"willofdaedalus/superluminal/internal/payload/term"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

// EncodePayload creates a payload with the provided arguments and using proto, marshalls
// it to bytes which it then returns ready to be sent across the wire
func EncodePayload(header common.Header, content isPayload_Content) ([]byte, error) {
	switch header {
	case common.Header_HEADER_ERROR:
		_, ok := content.(*Payload_Error)
		if !ok {
			return nil, ErrHeaderPayloadMismatch
		}
	// case common.Header_HEADER_AUTH_REQUEST:
	// case common.Header_HEADER_AUTH_RESPONSE:
	// case common.Header_HEADER_CLIENT_SHUTDOWN:
	// case common.Header_HEADER_DATA_RESEND:
	// case common.Header_HEADER_HEARTBEAT_REQUEST:
	// case common.Header_HEADER_HEARTBEAT_RESPONSE:
	// case common.Header_HEADER_TERMINAL_DATA:
	// case common.Header_HEADER_TYPE_SERVER_SHUTDOWN:
	// case common.Header_HEADER_UNSPECIFIED:
	default:
		fmt.Printf("unexpected common.Header: %#v", header)
	}

	payload := Payload{
		Version:   1,
		Header:    header,
		Timestamp: uint64(time.Now().Unix()),
		Content:   content,
	}

	data, err := proto.Marshal(&payload)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// DecodePayload takes the slice of bytes which was received through the wire, unmarshalls
// it with proto into a new Payload variable and returns the Payload and an error.
// Using the Payload, we can then view the contents of the Payload including the HeaderType,
// ActualContent and more
func DecodePayload(data []byte) (*Payload, error) {
	var payload Payload

	err := proto.Unmarshal(data, &payload)
	if err != nil {
		return nil, err
	}

	return &payload, nil
}

func GenerateError(errType err1.ErrorMessage_ErrorCode, errMsg []byte, deets []byte) *Payload_Error {
	return &Payload_Error{
		Error: &err1.ErrorMessage{
			Code:    errType,
			Message: errMsg,
			Detail:  deets,
		},
	}
}

// GenerateTermContent generates a new Payload of type term content which is passed to the Encoder
// to transform into bytes to be sent over the wire. Upon receiving the content, it is then appended
// to the last sent content
func GenerateTermContent(msgId string, msgLen int32, data []byte) Payload_TermContent {
	return Payload_TermContent{
		TermContent: &term.TerminalContent{
			MessageId:     uuid.NewString(),
			MessageLength: uint32(len(data)),
			Data:          data,
			Crc32:         crc32.ChecksumIEEE(data),
		},
	}
}
