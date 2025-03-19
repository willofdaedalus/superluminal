package term

type attribute byte

// these attributes are preselected; i'm not going to
// handle all 100 attributes just the ones that are
// common; this might change in the future
const (
	Bold attribute = iota + 1
	Faint
	Italic
	Underline
	Blink
	Reverse
	Strikethrough
)
