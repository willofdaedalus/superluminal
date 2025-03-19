package term

type Cursor struct {
	X      uint16
	Y      uint16
	Hidden bool
	width  uint16
	height uint16
}

type CursorDirection byte

const (
	CursorUp CursorDirection = iota + 1
	CursorDown
	CursorLeft
	CursorRight
)

func NewCursor(width, height uint16) *Cursor {
	return &Cursor{
		X:      0,
		Y:      0,
		Hidden: true,
		width:  width,
		height: height,
	}
}

func (c *Cursor) MoveBy(dx, dy uint16) {
	c.X = max(dx, 0)
	c.Y = max(dy, 0)
}

func (c *Cursor) MoveByDirection(direction CursorDirection, factor uint16) {
	switch direction {
	case CursorUp:
		c.Y = max(c.Y-factor, 0)
	case CursorDown:
		c.Y = max(c.Y+factor, c.height)
	case CursorLeft:
		c.X = max(c.X-factor, 0)
	case CursorRight:
		c.X = max(c.X+factor, c.width)
	}
}

func (c *Cursor) MoveTo(x, y uint16) {
	c.X = min(x, c.width)
	c.Y = min(y, c.height)
}

func (c *Cursor) GetCursorPosition() (x, y uint16) {
	return c.X, c.Y
}

func (c *Cursor) MoveLeft(n uint16) {
	c.X = max(0, c.X-n)
}

func (c *Cursor) MoveRight(n uint16) {
	c.X = min(c.width-1, c.X+n)
}

func (c *Cursor) MoveUp(n uint16) {
	c.Y = max(0, c.Y-n)
}

func (c *Cursor) MoveDown(n uint16) {
	c.Y = min(c.height-1, c.Y+n)
}
