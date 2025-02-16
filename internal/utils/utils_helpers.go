package utils

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"time"
)

// logBytes logs the length of received bytes with a timestamp to a file.
func LogBytes(action, filePath string, receivedBytes []byte) {
	// Open or create the log file
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer f.Close()

	// log the length of bytes and timestamp
	logEntry := fmt.Sprintf("%s: %s %d bytes\n", action, time.Now().Format(time.RFC3339), len(receivedBytes))
	if _, err := f.WriteString(logEntry); err != nil {
		fmt.Printf("Error writing to log file: %v\n", err)
	}
}

// utility function to safely close channels
func SafeClose[T any](ch chan T) {
	select {
	case <-ch:
		// channel is already closed
	default:
		close(ch)
	}
}

func PrependLength(payload []byte) []byte {
	pLen := len(payload)
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(pLen))
	message := append(header, payload...)
	return message
}

func RLEncode(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}

	var encoded []byte
	currentByte := data[0]
	count := 1

	for _, b := range data[1:] {
		if b == currentByte {
			count += 1
		} else {
			encoded = append(encoded, currentByte)
			encoded = append(encoded, []byte(strconv.Itoa(count))...)
			currentByte = b
			count = 1
		}
	}
	encoded = append(encoded, currentByte)
	encoded = append(encoded, []byte(strconv.Itoa(count))...)

	return encoded
}
