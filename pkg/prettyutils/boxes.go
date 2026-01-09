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


type Box struct {
	Width       int
	BorderStyle Style
	Segments    []Segment
	PaddingY    int
}

func NewBox(width int) *Box {
	return &Box{Width: width}
}

func (box *Box) AddSegment(s Segment) {
	box.Segments = append(box.Segments, s)
}


func (box *Box) Render() string {
	var out strings.Builder
	inner := box.Width - 2

	border := box.BorderStyle.ansi()

	out.WriteString(border + "┌" + strings.Repeat("─", inner) + "┐" + reset() + "\n")

	for i := 0; i < box.PaddingY; i++ {
		out.WriteString(box.emptyLine())
	}

	for _, s := range box.Segments {
		out.WriteString(box.renderSegment(s))
	}

	for i := 0; i < box.PaddingY; i++ {
		out.WriteString(box.emptyLine())
	}

	out.WriteString(border + "└" + strings.Repeat("─", inner) + "┘" + reset())

	return out.String()
}

func (box *Box) emptyLine() string {
	inner := box.Width - 2
	return box.BorderStyle.ansi() + "│" +
		strings.Repeat(" ", inner) +
		box.BorderStyle.ansi() + "│" +
		reset() + "\n"
}

func (box *Box) renderSegment(s Segment) string {
	inner := box.Width - 2
	textLen := runeLen(s.Text)

	var left int
	switch s.Align {
	case AlignCenter:
		left = (inner - textLen) / 2
	case AlignRight:
		left = inner - textLen
	}

	if left < 0 {
		left = 0
	}

	right := inner - left - textLen
	if right < 0 {
		right = 0
	}

	return box.BorderStyle.ansi() + "│" +
		strings.Repeat(" ", left) +
		s.Style.ansi() + s.Text + reset() +
		strings.Repeat(" ", right) +
		box.BorderStyle.ansi() + "│" +
		reset() + "\n"
}

func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}
