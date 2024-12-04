package pipeline

import (
	"hash/crc32"
	"strings"
	"testing"
	"willofdaedalus/superluminal/internal/payload/base"
	"willofdaedalus/superluminal/internal/payload/common"
	"willofdaedalus/superluminal/internal/payload/term"
)

func TestANSI24BitComplexEncoding(t *testing.T) {
	tests := []struct {
		name     string
		testData string
	}{
		{"Simple ANSI", "\033[38;2;255;0;0mThis is red text\033[0m"},
		{"Multiple ANSI", "\033[38;2;255;0;0mRed\033[0m \033[38;2;0;255;0mGreen\033[0m \033[38;2;0;0;255mBlue\033[0m"},
		{"Long ANSI", "\033[38;2;255;0;0m" + strings.Repeat("Red ", 2000) + "\033[0m"},
		{"Malformed ANSI", "\033[38;2;255;0;0mRed\033[38;Invalid\033[0mText"},
		{"Special Characters", "\033[38;2;255;0;0mRed\x1b[7mInverted\x1b[0mNormal"},
		{"Plain Text", "This is plain text without any ANSI codes."},
		{"UTF-8 and ANSI", "\033[38;2;255;0;0m你好, 世界\033[0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			termContent := &base.Payload_TermContent{
				TermContent: &term.TerminalContent{
					MessageId:     "test-ansi",
					MessageLength: uint32(len(tt.testData)),
					Data:          []byte(tt.testData),
					Crc32:         crc32.ChecksumIEEE([]byte(tt.testData)),
				},
			}

			encodedData, err := base.EncodePayload(common.Header_HEADER_TERMINAL_DATA, termContent)
			if err != nil {
				t.Fatalf("Failed to marshal protobuf for %s: %v", tt.name, err)
			}

			payload, err := base.DecodePayload(encodedData)
			if err != nil {
				t.Fatalf("Failed to unmarshal protobuf for %s: %v", tt.name, err)
			}

			termPayload := payload.GetTermContent()

			if string(termPayload.GetData()) != tt.testData {
				t.Errorf("Unmarshalled data mismatch for %s.\nExpected: %s\nGot: %s", tt.name, tt.testData, string(termPayload.GetData()))
			}
		})
	}
}
