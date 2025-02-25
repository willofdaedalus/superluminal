package parser

import (
	"bytes"
	"log"
)

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
	LF  byte = '\012' //
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
			log.Print("skipped to", string(b))
			return i
		}
	}

	// skip everything in this case
	return len(data)
}

func Scanner(data []byte) []token {
	var tokens []token

	for i := 0; i < len(data); i++ {
		switch data[i] {
		case CR, LF:
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
					log.Println("after skip", string(data[i]))

				case '[': // CSI sequence
					acc, offset := accumulator(data[i+1:], map[byte]struct{}{
						'm': {}, 'h': {}, 'l': {},
						'r': {}, 'J': {}, 'K': {},
					})

					if len(acc) > 0 { // ensure there's at least one byte
						lastByte := acc[len(acc)-1]
						if tk, ok := acceptedSequences[lastByte]; ok {
							tokens = append(tokens, token{tokenType: tk, content: acc})
						}
					}
					i += offset
				case '(':
					i += skipTo(data, map[byte]struct{}{ESC: {}})
				}
			}
		default:
			acc, offset := accumulator(data[i:], map[byte]struct{}{ESC: {}})
			tokens = append(tokens, token{tokenType: normalText, content: acc})
			i += offset - 1
		}
		// i++
	}
	return tokens
}

func accumulator(data []byte, stopChars map[byte]struct{}) ([]byte, int) {
	var buf bytes.Buffer
	for i, b := range data {
		if _, ok := stopChars[b]; ok {
			// if you continue to give, temporary solutions
			// to the poorest of the poor
			if b != ESC {
				buf.WriteByte(b)
				i += 1
			}
			return buf.Bytes(), i
		}
		buf.WriteByte(b)
	}
	return buf.Bytes(), len(data)
}
