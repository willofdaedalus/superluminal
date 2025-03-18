package term

type Cell struct {
	// the character this cell holds
	Character rune
	// both foreground and background colors of the cell
	FgCode     termColor
	BgCode     termColor
	Attributes attribute
	// has this cell changes since the last redraw
	Dirty bool
	// width for double-width characters like emojis and such
	Width uint8
}

func CreateCell(char rune, fg, bg termColor, attrs attribute, width uint8) *Cell {
	return &Cell{
		Character:  char,
		FgCode:     fg,
		BgCode:     bg,
		Attributes: attrs,
		Dirty:      false,
		Width:      width,
	}
}

func (c *Cell) SetCharacter(char rune) {
	c.Character = char
}

func (c *Cell) SetAttribute(attr attribute) {
	c.Attributes = attr
}

func (c *Cell) HasAttribute(attr attribute) bool {
	// bit shift to the location and then & it to check it's set
	return c.Attributes&(1<<attr) != 0
}

func (c *Cell) ClearAttribute(attrToClear attribute) {
	c.Attributes &^= (1 << attrToClear)
}

func (c *Cell) SetColour(fg, bg termColor) {
	c.FgCode = fg
	c.BgCode = bg
}

func (c *Cell) SetForegroundColour(fg termColor) {
	c.FgCode = fg
}

func (c *Cell) SetBackgroundColour(bg termColor) {
	c.BgCode = bg
}

func (c *Cell) SetDirty() {
	c.Dirty = true
}

func (c *Cell) ClearDirty() {
	c.Dirty = false
}

func (c *Cell) IsDirty() bool {
	return c.Dirty
}

func (c *Cell) Reset() {
	c.Character = ' '
	c.FgCode = F_Default
	c.BgCode = B_Default
	c.Attributes = 0
	c.Dirty = true
	c.Width = 1
}
