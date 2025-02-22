package parser

import "bytes"

type tokenType int

const (
	normalText tokenType = iota + 1
	ansiColour
	cursorControl
	modeControl
	newline
)

const (
	BEL byte = '\007' // ^G
	ESC byte = '\033' // ^[
	CR  byte = '\015' // ^M
)

type token struct {
	tokenType tokenType
	content   []byte
}

var acceptedSequences = map[byte]tokenType{
	'm': ansiColour,
	'K': cursorControl,
	'J': cursorControl,
	'H': cursorControl,
}

func skipTo(data []byte, ends map[byte]struct{}) int {
	for i, b := range data {
		if _, exists := ends[b]; exists {
			return i + 1
		}
	}

	// skip everything in this case
	return len(data)
}

func scanner(data []byte) []token {
	var tokens []token

	i := 0
	for i < len(data) {
		switch data[i] {
		case CR:
			tokens = append(tokens, token{
				tokenType: newline,
				content:   []byte{'\n'},
			})

		case ESC:
			if i+1 < len(data) {
				next := data[i+1]
				switch next {
				case ']': // OSC sequence
					i += skipTo(data, map[byte]struct{}{BEL: {}})

				case '[': // CSI sequence
					acc, offset := accumulator(data[i:], map[byte]struct{}{
						'm': {}, 'h': {}, 'l': {},
						'r': {}, 'J': {}, 'K': {},
					})

					if len(acc) > 0 { // ensure there's at least one byte
						lastByte := acc[len(acc)-1]
						if tk, ok := acceptedSequences[lastByte]; ok {
							tokens = append(tokens, token{tokenType: tk, content: acc})
							i += offset
						}
					}
				case '(':
					i += skipTo(data, map[byte]struct{}{ESC: {}})
				}
			}
		default:
			accumulated, offset := accumulator(data[i:], map[byte]struct{}{ESC: {}})
			tokens = append(tokens, token{tokenType: normalText, content: accumulated})
			i += offset
		}
	}
	return tokens
}

func accumulator(data []byte, stopChars map[byte]struct{}) ([]byte, int) {
	var buf bytes.Buffer
	for i, b := range data {
		if _, exists := stopChars[b]; exists {
			return buf.Bytes(), i + 1
		}
		buf.WriteByte(b)
	}
	return buf.Bytes(), len(data)
}
