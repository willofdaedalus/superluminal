package term

type Cell struct {
	// the character this cell holds
	Character rune
	// both foreground and background colors of the cell
	FgCode uint32
	BgCode uint32
	// these attributes are preselected; i'm not going to
	// handle all 100 attributes just the ones that are
	// common; this might change in the future
	Attributes byte
	// has this cell changes since the last redraw
	Dirty bool
	// width for double-width characters like emojis and such
	Width uint8
}

func CreateCell(char rune, fg, bg uint32, attrs byte, width uint8) *Cell {
	return &Cell{
		Character:  char,
		FgCode:     fg,
		BgCode:     bg,
		Attributes: attrs,
		Dirty:      false,
		Width:      width,
	}
}

func (c *Cell) SetColour(fg, bg uint32) {
	c.FgCode = fg
	c.BgCode = bg
}
