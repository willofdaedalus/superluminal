package term

// normal colours
const (
	// foreground colours
	F_Black uint32 = iota + 30
	F_Red
	F_Green
	F_Yellow
	F_Blue
	F_Magenta
	F_Cyan
	F_White
	// background colours
	B_Black uint32 = iota + 32
	B_Red
	B_Green
	B_Yellow
	B_Blue
	B_Magenta
	B_Cyan
	B_White
)

// bright colours
const (
	// foreground colours
	F_BrightBlack uint32 = iota + 90
	F_BrightRed
	F_BrightGreen
	F_BrightYellow
	F_BrightBlue
	F_BrightMagenta
	F_BrightCyan
	F_BrightWhite
	// background colours
	B_BrightBlack uint32 = iota + 92
	B_BrightRed
	B_BrightGreen
	B_BrightYellow
	B_BrightBlue
	B_BrightMagenta
	B_BrightCyan
	B_BrightWhite
)
