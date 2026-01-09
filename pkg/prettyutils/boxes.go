package prettyutils

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

type Color interface {
	fg() string
}

type BasicColor int

const (
	Black BasicColor = iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	BrightBlack
	BrightRed
	BrightGreen
	BrightYellow
	BrightBlue
	BrightMagenta
	BrightCyan
	BrightWhite
)

type Basic BasicColor

func (c Basic) fg() string {
	if c <= 7 {
		return "3" + c.fg()
	}
	return "9" + (c - 8).fg()
}

type ANSI256 uint8

func (c ANSI256) fg() string {
	return "38;5;" + c.fg()
}

type RGB struct {
	R, G, B uint8
}

func RGBColor(r, g, b uint8) RGB {
	return RGB{r, g, b}
}

func (c RGB) fg() string {
	return "38;2;" + 
		strconv.FormatUint(uint64(c.R), 8) + ";" + 
		strconv.FormatUint(uint64(c.B), 8) + ";" + 
		strconv.FormatUint(uint64(c.G), 8)
}

type Style struct {
	Fg     Color
	Bold   bool
	Italic bool
}

func (s Style) ansi() string {
	codes := []string{}

	if s.Bold {
		codes = append(codes, "1")
	}
	if s.Italic {
		codes = append(codes, "3")
	}
	if s.Fg != nil {
		codes = append(codes, s.Fg.fg())
	}

	if len(codes) == 0 {
		return ""
	}
	return "\033[" + strings.Join(codes, ";") + "m"
}

func reset() string {
	return "\033[0m"
}

type Align int

const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)

type Segment struct {
	Text  string
	Align Align
	Style Style
}

type SizeMode int

const (
	FixedWidth SizeMode = iota
	AutoWidth
)


type Box struct {
	Width       int
	SizeMode    SizeMode
	MinWidth    int
	PaddingX    int
	PaddingY    int
	BorderStyle Style
	Segments    []Segment
}

func NewBox(width int) *Box {
	return &Box{
		Width:    width,
		SizeMode: FixedWidth,
		PaddingX: 1,
	}
}

func NewAutoBox() *Box {
	return &Box{
		SizeMode: AutoWidth,
		PaddingX: 1,
	}
}


func (box *Box) AddSegment(s Segment) {
	box.Segments = append(box.Segments, s)
}


func (box *Box) Render() string {
	if box.SizeMode == AutoWidth {
		box.Width = box.computeWidth()
	}

	// failsafe for box smaller than 2 border
	if box.Width < 2 {
		box.Width = 2
	}

	inner := box.Width - 2

	var out strings.Builder
	border := box.BorderStyle.ansi()

	out.WriteString(border)
	out.WriteString("┌" + strings.Repeat("─", inner) + "┐")
	out.WriteString(reset() + "\n")

	for i := 0; i < box.PaddingY; i++ {
		out.WriteString(box.emptyLine())
	}

	for _, s := range box.Segments {
		out.WriteString(box.renderSegment(s))
	}

	for i := 0; i < box.PaddingY; i++ {
		out.WriteString(box.emptyLine())
	}

	out.WriteString(border)
	out.WriteString("└" + strings.Repeat("─", inner) + "┘")
	out.WriteString(reset())

	return out.String()
}

func (box *Box) emptyLine() string {
	inner := box.Width - 2
	return box.BorderStyle.ansi() + "│" +
		strings.Repeat(" ", inner) +
		box.BorderStyle.ansi() + "│" +
		reset() + "\n"
}

func (box *Box) computeWidth() int {
	max := 0

	for _, s := range box.Segments {
		l := runeLen(s.Text)
		if l > max {
			max = l
		}
	}

	inner := max + (box.PaddingX * 2)
	width := inner + 2 // offset because of borders

	if box.MinWidth > 0 && width < box.MinWidth {
		width = box.MinWidth
	}

	return width
}

func (box *Box) renderSegment(s Segment) string {
	inner := box.Width - 2 - (box.PaddingX * 2)
	textLen := runeLen(s.Text)

	var left int
	switch s.Align {
	case AlignCenter:
		left = (inner - textLen) / 2
	case AlignRight:
		left = inner - textLen
	}

	left = max(left, 0)

	right := max(inner - left - textLen, 0)

	return box.BorderStyle.ansi() + "│" +
		strings.Repeat(" ", box.PaddingX + left) +
		s.Style.ansi() + s.Text + reset() +
		strings.Repeat(" ", right + box.PaddingX) +
		box.BorderStyle.ansi() + "│" +
		reset() + "\n"
}

func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}
