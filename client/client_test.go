package client

import (
	"bytes"
	"fmt"
	"testing"
	"willofdaedalus/superluminal/utils"
)

func TestParseHeader(t *testing.T) {
	testHeader := fmt.Sprintf("%shello world", utils.HdrInfo)
	header, _, _ := bytes.Cut([]byte(testHeader), []byte(":"))
	t.Log(testHeader)

	_, err := parseServerHeader([]byte(header))
	if err != nil {
		t.Fatal(err)
	}
}
