package base

import (
	"errors"
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

	// intentionally pass the wrong header type to the payload
	// to trigger an ErrHeaderPayloadMismatch
	_, err := EncodePayload(common.Header_HEADER_DATA_RESEND, errorMsg)

	// check to make sure we're not returning some other error
	if !errors.Is(err, ErrHeaderPayloadMismatch) {
		t.Fatalf("expected a mismatch; got %v", err)
	}
}

func FuzzDecodePayload(f *testing.F) {
	errorMsg := GenerateError(
		err1.ErrorMessage_ERROR_SERVER_FULL,
		[]byte("server_full"),
		[]byte("server is full"),
	)
	errPayload, _ := EncodePayload(common.Header_HEADER_ERROR, errorMsg)
	// adding initial seed inputs for fuzzing
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
