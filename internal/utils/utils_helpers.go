package utils

import (
	"encoding/binary"
	"fmt"
	"os"
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
