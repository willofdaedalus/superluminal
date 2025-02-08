package base

import (
	"reflect"
	"testing"
	"willofdaedalus/superluminal/internal/payload/common"
	err1 "willofdaedalus/superluminal/internal/payload/error"
	"willofdaedalus/superluminal/internal/payload/info"
)

func TestEncodePayload(t *testing.T) {
	type args struct {
		header  common.Header
		content isPayload_Content
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodePayload(tt.args.header, tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodePayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EncodePayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateInfo(t *testing.T) {
	type args struct {
		infoType info.Info_InfoType
		message  string
	}
	tests := []struct {
		name string
		args args
		want *Payload_Info
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateInfo(tt.args.infoType, tt.args.message); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodePayload(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Payload
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodePayload(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodePayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecodePayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateError(t *testing.T) {
	type args struct {
		errType err1.ErrorMessage_ErrorCode
		errMsg  []byte
		deets   []byte
	}
	tests := []struct {
		name string
		args args
		want *Payload_Error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateError(tt.args.errType, tt.args.errMsg, tt.args.deets); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateAuthResp(t *testing.T) {
	type args struct {
		name string
		pass string
	}
	tests := []struct {
		name string
		args args
		want *Payload_Auth
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateAuthResp(tt.args.name, tt.args.pass); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateAuthResp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateAuthReq(t *testing.T) {
	tests := []struct {
		name string
		want *Payload_Auth
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateAuthReq(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateAuthReq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateTermContent(t *testing.T) {
	type args struct {
		msgId  string
		msgLen int32
		data   []byte
	}
	tests := []struct {
		name string
		args args
		want Payload_TermContent
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateTermContent(tt.args.msgId, tt.args.msgLen, tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateTermContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateHeartbeatReq(t *testing.T) {
	tests := []struct {
		name string
		want Payload_Heartbeat
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateHeartbeatReq(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateHeartbeatReq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateHeartbeatResp(t *testing.T) {
	tests := []struct {
		name string
		want Payload_Heartbeat
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateHeartbeatResp(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateHeartbeatResp() = %v, want %v", got, tt.want)
			}
		})
	}
}
