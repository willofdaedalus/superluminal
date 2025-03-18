package parser

import (
	"bytes"
	"strconv"
	"willofdaedalus/superluminal/internal/utils"
)

func EncodeData(data []byte) []byte {
	tokens := scanner(data)
	if len(tokens) < 0 {
		return []byte{0}
	}

	return encode(tokens)
}

func encode(tks []token) []byte {
	encoded := make([]byte, 0)

	for _, tk := range tks {
		switch tk.tokenType {
		case ansiColour:
			final := parseAnsiColours(tk.content)
			encoded = append(encoded, final...)
		case newline:
			encoded = append(encoded, '\n')
		case normalText:
			// if it can't be rlencoded then it's most likely not needed
			if rlencoded := utils.RLEncode(tk.content); rlencoded != nil {
				encoded = append(encoded, rlencoded...)
				continue
			}
			encoded = append(encoded, tk.content...)
		}
	}

	return encoded
}

func parseAnsiColours(content []byte) []byte {
	var attr byte = 0
	var fgBright, bgBright byte = 0, 0
	var fgColor, bgColor byte = 0, 0

	if len(content) == 0 { // Reset case
		return []byte{ESC, 0, 0, 0} // Just return zeroed attributes
	}

	tks := bytes.Split(content, []byte{';'})
	for i := 0; i < len(tks); i++ {
		num, err := strconv.Atoi(string(tks[i]))
		if err != nil {
			continue
		}
		switch num {
		case 0:
			attr, fgBright, bgBright, fgColor, bgColor = 0, 0, 0, 0, 0
		case 1:
			attr |= 1 << 0 // bold
		case 2:
			attr |= 1 << 1 // faint
		case 3:
			attr |= 1 << 2 // italic
		case 4:
			attr |= 1 << 3 // underline
		case 5, 6:
			attr |= 1 << 4 // blink
		case 7:
			attr |= 1 << 5 // reverse
		case 9:
			attr |= 1 << 6 // strikethrough

		// foreground colour
		case 30, 31, 32, 33, 34, 35, 36, 37:
			fgColor = byte(num - 30)
		case 90, 91, 92, 93, 94, 95, 96, 97:
			// colour codes within the 90s range are reserved for bright
			// so that instead of writing something like [1;31 for bright red,
			// it can instead be something like [91 for the same bright red
			fgColor = byte(num - 90)
			fgBright = 1

		// background colour handling
		case 40, 41, 42, 43, 44, 45, 46, 47:
			bgColor = byte(num - 40)
		case 100, 101, 102, 103, 104, 105, 106, 107:
			// same reason for the 90s range, numbers between 100 and 107 are
			// reserved for bright background colours
			bgColor = byte(num - 100)
			bgBright = 1
		}
	}

	// Return 3-byte format: [attr][(fgBright<<3)|fgColor][(bgBright<<3)|bgColor]
	return []byte{
		ESC, attr,
		(fgBright << 3) | fgColor,
		(bgBright << 3) | bgColor,
	}
}
