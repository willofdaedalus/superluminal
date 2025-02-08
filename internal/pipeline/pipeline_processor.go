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

// func sanitizeRawCode(data []byte) []byte {
// 	result := make([]byte, 0, len(data))

// 	for len(data) > 0 {
// 		b := data[0]
// 		data = data[1:] // Remove the current byte from processing

// 		switch b {
// 		case OSC: // Operating System Command
// 			data = skipAheadTo(data, '\a') // Skip until BEL
// 		case ESC:
// 			if len(data) > 0 {
// 				switch data[0] {
// 				case '[': // CSI
// 					data = skipAheadToCommand(data)
// 				case ']': // OSC
// 					data = skipAheadTo(data, '\a')
// 				case 'P': // DCS
// 					data = skipAheadTo(data, '\x1B') // Skip to ST (`ESC \`)
// 				default:
// 					// Handle other `ESC` sequences as needed
// 				}
// 			}
// 		case CSI: // Control Sequence Introducer
// 			data = skipAheadToCommand(data)
// 		case DCS: // Device Control String
// 			data = skipAheadTo(data, '\x1B') // Skip to ST (`ESC \`)
// 		default:
// 			result = append(result, b) // Copy non-control bytes
// 		}
// 	}

// 	return result
// }

// func skipAheadToCommand(data []byte) []byte {
// 	// Skip until a command-ending character (letters or specific symbols)
// 	for len(data) > 0 {
// 		b := data[0]
// 		data = data[1:]
// 		if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || b == '@' || b == '~' {
// 			break
// 		}
// 	}
// 	return data
// }

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

// func makePtyGrid(data []byte, prevData [][]byte) [][]byte {
// 	gridOutput := make([][]byte, 0)
// 	var currentIdx int

// 	for i, c := range data {
// 		line := make([]byte, 0)
// 		if c == '\n' {
// 			line = data
// 		}

// 		gridOutput = append(gridOutput, line)
// 	}

// 	return gridOutput
// }
