package base

import (
	"fmt"
	"hash/crc32"
	"log"
	"time"
	"willofdaedalus/superluminal/internal/payload/auth"
	"willofdaedalus/superluminal/internal/payload/common"
	err1 "willofdaedalus/superluminal/internal/payload/error"
	"willofdaedalus/superluminal/internal/payload/heartbeat"
	"willofdaedalus/superluminal/internal/payload/info"
	"willofdaedalus/superluminal/internal/payload/term"
	"willofdaedalus/superluminal/internal/utils"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

type PayloadType int

const (
	PayloadUnknown = iota + 1
	PayloadAuth
	PayloadTermContent
	PayloadHeartbeat
	PayloadError
	PayloadInfo
)

// EncodePayload creates a payload with the provided arguments and using proto, marshalls
// it to bytes which it then returns ready to be sent across the wire
func EncodePayload(header common.Header, content isPayload_Content) ([]byte, error) {
	switch header {
	case common.Header_HEADER_ERROR:
		if GetPayloadType(content) != PayloadError {
			return nil, utils.ErrPayloadHeaderMismatch
		}
	case common.Header_HEADER_AUTH:
		if GetPayloadType(content) != PayloadAuth {
			return nil, utils.ErrPayloadHeaderMismatch
		}
	case common.Header_HEADER_INFO:
		if GetPayloadType(content) != PayloadInfo {
			return nil, utils.ErrPayloadHeaderMismatch
		}
	case common.Header_HEADER_TERMINAL_DATA:
		if GetPayloadType(content) != PayloadTermContent {
			return nil, utils.ErrPayloadHeaderMismatch
		}

	default:
		return nil, utils.ErrPayloadHeaderMismatch
	}

	payload := Payload{
		Version:   1,
		Header:    header,
		Timestamp: uint64(time.Now().Unix()),
		Content:   content,
	}

	data, err := proto.Marshal(&payload)
	if err != nil {
		return nil, fmt.Errorf("encode err: %v", err)
	}

	return data, nil
}

func GetPayloadType(payload isPayload_Content) PayloadType {
	switch payload.(type) {
	case *Payload_Auth:
		return PayloadAuth
	case *Payload_TermContent:
		return PayloadTermContent
	case *Payload_Info:
		return PayloadInfo
	case *Payload_Heartbeat:
		return PayloadHeartbeat
	default:
		return PayloadUnknown
	}
}

// GenerateInfo generates an info payload that contains something for its
// destination
func GenerateInfo(infoType info.Info_InfoType, message string) *Payload_Info {
	return &Payload_Info{
		Info: &info.Info{
			InfoType: infoType,
			Message:  message,
		},
	}
}

// DecodePayload takes the slice of bytes which was received through the wire, unmarshalls
// it with proto into a new Payload variable and returns the Payload and an error.
// Using the Payload, we can then view the contents of the Payload including the HeaderType,
// ActualContent and more
func DecodePayload(data []byte) (*Payload, error) {
	var payload Payload

	if data == nil {
		// return nil, fmt.Errorf("can't decode nil/empty data bytes")
		log.Fatal("can't decode nil data bytes")
	}

	if len(data) == 0 {
		// return nil, fmt.Errorf("can't decode nil/empty data bytes")
		log.Fatal("can't decode empty data bytes")
	}

	err := proto.Unmarshal(data, &payload)
	if err != nil {
		return nil, fmt.Errorf("decode err: %v", err)
	}

	return &payload, nil
}

// GenerateError creates and passes a Payload of type Error which satisfies the interface requirement
// of EncodePayload. Used to share error messages
func GenerateError(errType err1.ErrorMessage_ErrorCode, errMsg []byte, deets []byte) *Payload_Error {
	return &Payload_Error{
		Error: &err1.ErrorMessage{
			Code:    errType,
			Message: errMsg,
			Detail:  deets,
		},
	}
}

// GenerateAuthResp generates an auth response payload comprised of the client's name and passphrase
// and is sent over the wire to the session to act as a simple barrier for authentication
func GenerateAuthResp(name, pass string) *Payload_Auth {
	return &Payload_Auth{
		Auth: &auth.Authentication{
			Auth: auth.Authentication_AUTH_TYPE_RESPONSE,
			AuthType: &auth.Authentication_Response{
				Response: &auth.AuthResponse{
					Username:   name,
					Passphrase: pass,
				},
			},
		},
	}
}

func GenerateAuthReq() *Payload_Auth {
	// authentication request doesn't need any information for now as all the client needs to know
	// from the succeeding payload header is that it's an auth request
	// this is different from an auth response which actually sends a password and a name to the
	// session
	return &Payload_Auth{
		Auth: &auth.Authentication{
			Auth: auth.Authentication_AUTH_TYPE_REQUEST,
			// AuthType: auth.Authentication_Request{},
		},
	}
}

// GenerateTermContent generates a new Payload of type term content which is passed to the Encoder
// to transform into bytes to be sent over the wire. Upon receiving the content, it is then appended
// to the last sent content
func GenerateTermContent(msgId string, dataLen uint32, data []byte) Payload_TermContent {
	return Payload_TermContent{
		TermContent: &term.TerminalContent{
			MessageId:     uuid.NewString(),
			MessageLength: dataLen,
			Data:          data,
			Crc32:         crc32.ChecksumIEEE(data),
		},
	}
}

func GenerateHeartbeatReq() Payload_Heartbeat {
	return Payload_Heartbeat{
		Heartbeat: &heartbeat.Heartbeat{
			Type:    heartbeat.Heartbeat_HEARTBEAT_TYPE_PING,
			Payload: "ping",
		},
	}
}

func GenerateHeartbeatResp() Payload_Heartbeat {
	return Payload_Heartbeat{
		Heartbeat: &heartbeat.Heartbeat{
			Type:    heartbeat.Heartbeat_HEARTBEAT_TYPE_PONG,
			Payload: "pong",
		},
	}
}

// func FixPayloadHeader(payload isPayload_Content) {

// }
