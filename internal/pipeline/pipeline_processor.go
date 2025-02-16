package pipeline

import "bytes"

// control sequences
const (
	BEL byte = '\007' // ^G
	ESC byte = '\033' // ^[
	CR  byte = '\015' // ^M
)

func skipAheadTo(data []byte, b []byte) []byte {
	result := data
	for i, curByte := range data {
		if bytes.Contains(b, []byte{curByte}) {
			result = data[i+1:]
			break
		}
	}

	return result
}

// returns a sanitized and clean pty output
func sanitizeRawCode(data []byte) []byte {
	result := make([]byte, 0, len(data))

	for len(data) > 0 {
		b := data[0]
		data = data[1:]

		switch b {
		case ESC:
			delimiter := data[0]
			data = handleDelimiter(delimiter, data)
		default:
			result = append(result, b)
		}
	}

	return result
}

func handleDelimiter(del byte, data []byte) []byte {
	switch del {
	case ']':
		data = skipAheadTo(data, []byte{BEL})
	case '[':
		skipTo := make([]byte, 0)
		switch data[1] {
		case '?':
			skipTo = append(skipTo, 'h', 'l')
		}

		data = skipAheadTo(data, []byte{'h', 'l'})
	}

	return data
}
