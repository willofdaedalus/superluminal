package term

type Screen struct {
	Cursor    *Cursor
	CellGrid  [][]Cell
	MaxWidth  uint16
	MaxHeight uint16
}

func InitializeScreen(width, height uint16) *Screen {
	grid := make([][]Cell, height)
	for i := range grid {
		row := make([]Cell, width)
		for j := range row {
			row[j] = Cell{' ', F_Default, B_Default, 0, false, 0}
		}
		grid[i] = row
	}

	return &Screen{
		CellGrid:  grid,
		Cursor:    NewCursor(width, height),
		MaxWidth:  width,
		MaxHeight: height,
	}
}

func (s *Screen) SetHeight(h uint16) {
	s.MaxHeight = h
}

func (s *Screen) SetWidth(w uint16) {
	s.MaxWidth = w
}

func (s *Screen) ResetScreen() {
	for i := range s.CellGrid {
		for j := range s.CellGrid[i] {
			s.CellGrid[i][j].Reset()
		}
	}
	s.Cursor.MoveTo(0, 0)
}
