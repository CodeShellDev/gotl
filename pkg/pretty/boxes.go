package pretty

import (
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
)

type Color interface {
	foreground() string
	background() string
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

func (c Basic) foreground() string {
	if c <= 7 {
		return "3" + strconv.Itoa(int(c))
	}
	return "9" + strconv.Itoa(int(c) - 8)
}

func (c Basic) background() string {
	if c <= 7 {
		return "4" + strconv.Itoa(int(c))
	}
	return "10" + strconv.Itoa(int(c) - 8)
}

type ANSI256 uint8

func (c ANSI256) foreground() string {
	return "38;5;" + strconv.Itoa(int(c))
}

func (c ANSI256) background() string {
	return "48;5;" + strconv.Itoa(int(c))
}

type RGB struct {
	R, G, B uint8
}

func RGBColor(r, g, b uint8) RGB {
	return RGB{r, g, b}
}

func (c RGB) foreground() string {
	return "38;2;" +
		strconv.Itoa(int(c.R)) + ";" +
		strconv.Itoa(int(c.G)) + ";" +
		strconv.Itoa(int(c.B))
}

func (c RGB) background() string {
	return "48;2;" +
		strconv.Itoa(int(c.R)) + ";" +
		strconv.Itoa(int(c.G)) + ";" +
		strconv.Itoa(int(c.B))
}

type Style struct {
	Foreground	Color
	Background  Color
	Bold   		bool
	Italic		bool
}

func (s1 Style) Combine(s2 Style) Style {
	result := s1

	if s2.Foreground != nil {
		result.Foreground = s2.Foreground
	}
	if s2.Background != nil {
		result.Background = s2.Background
	}

	result.Bold = s2.Bold
	result.Italic = s2.Italic

	return result
}

func (s Style) ansi() string {
	codes := []string{}

	if s.Bold {
		codes = append(codes, "1")
	}
	if s.Italic {
		codes = append(codes, "3")
	}
	if s.Foreground != nil {
		codes = append(codes, s.Foreground.foreground())
	}
	if s.Background != nil {
		codes = append(codes, s.Background.background())
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
	MarginX		int
	MarginY		int
	Style		BoxStyle
	Border		Border
	Segments    []Segment
}

type BoxStyle struct {
	Background	Color
	Border		BorderStyle
}

type Border struct {
	Chars		BorderChars
	Style		BorderStyle
}

type BorderChars struct {
	Horizontal	rune
	Vertical	rune
	TopLeft		rune
	TopRight	rune
	BottomLeft 	rune
	BottomRight	rune
}

type BorderStyle struct {
	Color		Color
	Bold		bool
	Italic		bool
}

func (s BoxStyle) Base() Style {
	return Style{
		Background: s.Background,
	}
}

func (s BoxStyle) ansi() string {
	return s.Base().ansi()
}

func (s BorderStyle) Base() Style {
	return Style{
		Foreground: s.Color,
		Bold: s.Bold,
		Italic: s.Italic,
	}
}

func (s BorderStyle) ansi() string {
	return s.Base().ansi()
}

func NewBox(width int) *Box {
	return &Box{
		Width:    width,
		SizeMode: FixedWidth,
		PaddingX: 1,
		Border: Border{
			Chars: BorderChars{
				Horizontal:  '─',
				Vertical:    '│',
				TopLeft:     '┌',
				TopRight:    '┐',
				BottomLeft:  '└',
				BottomRight: '┘',
			},
		},
	}
}

func NewAutoBox() *Box {
	box := NewBox(0)

	box.SizeMode = AutoWidth

	return box
}


func (box *Box) AddSegment(s Segment) {
	box.Segments = append(box.Segments, s)
}

func (box *Box) renderTop(width int) string {
	return box.Style.Base().Combine(box.Style.Border.Base()).ansi() +
		string(box.Border.Chars.TopLeft) + strings.Repeat(string(box.Border.Chars.Horizontal), width) + string(box.Border.Chars.TopRight) +
		reset()
}

func (box *Box) renderBottom(width int) string {
	return box.Style.Base().Combine(box.Style.Border.Base()).ansi() +
		string(box.Border.Chars.BottomLeft) + strings.Repeat(string(box.Border.Chars.Horizontal), width) + string(box.Border.Chars.BottomRight) +
		reset()
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

	out.WriteString(strings.Repeat("\n", box.MarginY))
	out.WriteString(strings.Repeat(" ", box.MarginX))

	out.WriteString(box.renderTop(inner))

	out.WriteString(strings.Repeat(" ", box.MarginX))
	out.WriteString(strings.Repeat("\n", box.MarginY))

	for i := 0; i < box.PaddingY; i++ {
		out.WriteString(strings.Repeat(" ", box.MarginX))

		out.WriteString(box.emptyLine())

		out.WriteString(strings.Repeat(" ", box.MarginX) + "\n")
	}

	for _, s := range box.Segments {
		out.WriteString(strings.Repeat(" ", box.MarginX))

		out.WriteString(box.renderSegment(s))

		out.WriteString(strings.Repeat(" ", box.MarginX) + "\n")
	}

	for i := 0; i < box.PaddingY; i++ {
		out.WriteString(strings.Repeat(" ", box.MarginX))

		out.WriteString(box.emptyLine())

		out.WriteString(strings.Repeat(" ", box.MarginX) + "\n")
	}

	out.WriteString(strings.Repeat(" ", box.MarginX))

	out.WriteString(box.renderBottom(inner))

	out.WriteString(strings.Repeat(" ", box.MarginX))
	out.WriteString(strings.Repeat("\n", box.MarginY))

	return out.String()
}

func (box *Box) emptyLine() string {
	inner := box.Width - 2
	return box.Style.Base().Combine(box.Style.Border.Base()).ansi() + "│" +
		strings.Repeat(" ", inner) +
		box.Style.Base().Combine(box.Style.Border.Base()).ansi() + "│" +
		reset()
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

func getPadding(text string, width int, align Align) (int, int) {
	textLen := runeLen(text)
	if width <= textLen {
		return 0, 0
	}

	space := width - textLen
	var left, right int

	switch align {
	case AlignLeft:
		left = 0
		right = space
	case AlignRight:
		left = space
		right = 0
	case AlignCenter:
		left = space / 2
		right = space - left
	}

	return left, right
}

func (box *Box) renderSegment(s Segment) string {
	inner := box.Width - 2 - (box.PaddingX * 2)
	paddingLeft, paddingRight := getPadding(s.Text, inner, s.Align)

	return box.Style.Base().Combine(box.Style.Border.Base()).ansi() + string(box.Border.Chars.Vertical) +
		reset() +
		box.Style.ansi() +
		strings.Repeat(" ", box.PaddingX) +
		strings.Repeat(" ", paddingLeft) +
		box.Style.Base().Combine(s.Style).ansi() + s.Text +
		reset() +
		box.Style.ansi() +
		strings.Repeat(" ", paddingRight) +
		strings.Repeat(" ", box.PaddingX) +
		box.Style.Base().Combine(box.Style.Border.Base()).ansi() + string(box.Border.Chars.Vertical) +
		reset()
	}

func runeLen(s string) int {
	return runewidth.StringWidth(s)
}