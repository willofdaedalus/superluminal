package base

import (
	"errors"
	"reflect"
	"testing"
	"willofdaedalus/superluminal/internal/payload/common"
	err1 "willofdaedalus/superluminal/internal/payload/error"
)

func TestEncodeFunc(t *testing.T) {
	errorMsg := GenerateError(
		err1.ErrorMessage_ERROR_SERVER_FULL,
		[]byte("server_full"),
		[]byte("server is full"),
	)

	_, err := EncodePayload(common.Header_HEADER_DATA_RESEND, errorMsg)

	if !errors.Is(err, ErrHeaderPayloadMismatch) {
		t.Fatalf("expected a mismatch; got %v", err)
	}

}

func TestGenerateError(t *testing.T) {
	errorType := err1.ErrorMessage_ERROR_UNSPECIFIED

	testCases := []struct {
		name     string
		errType  err1.ErrorMessage_ErrorCode
		errMsg   []byte
		deets    []byte
		expected Payload_Error
	}{
		{
			name:    "Basic input test",
			errType: errorType,
			errMsg:  []byte("Standard error"),
			deets:   []byte("Details here"),
			expected: Payload_Error{
				Error: &err1.ErrorMessage{
					Code:    errorType,
					Message: []byte("Standard error"),
					Detail:  []byte("Details here"),
				},
			},
		},
		{
			name:    "Empty fields test",
			errType: errorType,
			errMsg:  []byte(""),
			deets:   []byte(""),
			expected: Payload_Error{
				Error: &err1.ErrorMessage{
					Code:    errorType,
					Message: []byte(""),
					Detail:  []byte(""),
				},
			},
		},
		{
			name:    "Special characters and Unicode test",
			errType: errorType,
			errMsg:  []byte("Error with emoji 😊 and special chars !@#$"),
			deets:   []byte("Details with unicode こんにちは"),
			expected: Payload_Error{
				Error: &err1.ErrorMessage{
					Code:    errorType,
					Message: []byte("Error with emoji 😊 and special chars !@#$"),
					Detail:  []byte("Details with unicode こんにちは"),
				},
			},
		},
		{
			name:    "Binary data test",
			errType: errorType,
			errMsg:  []byte{0x00, 0x01, 0xFF},
			deets:   []byte{0xAB, 0xCD},
			expected: Payload_Error{
				Error: &err1.ErrorMessage{
					Code:    errorType,
					Message: []byte{0x00, 0x01, 0xFF},
					Detail:  []byte{0xAB, 0xCD},
				},
			},
		},
		{
			name:    "Large data test",
			errType: errorType,
			errMsg:  make([]byte, 1e6), // 1 MB of zeroed bytes for simplicity
			deets:   make([]byte, 1e6),
			expected: Payload_Error{
				Error: &err1.ErrorMessage{
					Code:    errorType,
					Message: make([]byte, 1e6),
					Detail:  make([]byte, 1e6),
				},
			},
		},
		{
			name:    "Multiline message test",
			errType: errorType,
			errMsg:  []byte("Error\nLine2\nLine3"),
			deets:   []byte("Detail with\nmultilines"),
			expected: Payload_Error{
				Error: &err1.ErrorMessage{
					Code:    errorType,
					Message: []byte("Error\nLine2\nLine3"),
					Detail:  []byte("Detail with\nmultilines"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GenerateError(tc.errType, tc.errMsg, tc.deets)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Test %s failed: expected %v, got %v", tc.name, tc.expected, result)
			}
		})
	}
}

func FuzzDecodePayload(f *testing.F) {
	errorMsg := GenerateError(
		err1.ErrorMessage_ERROR_SERVER_FULL,
		[]byte("server_full"),
		[]byte("server is full"),
	)
	errPayload, _ := EncodePayload(common.Header_HEADER_ERROR, errorMsg)
	// Adding initial seed inputs for fuzzing
	f.Add(errPayload)
	f.Add([]byte{})
	f.Add([]byte("fake data"))

	f.Fuzz(func(t *testing.T, data []byte) {
		payload, err := DecodePayload(data)
		if err != nil && payload != nil {
			t.Errorf("DecodePayload returned non-nil payload with an error: %v", err)
		}
	})
}
