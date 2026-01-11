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

type Span struct {
	Text  string
	Style Style
}

type Line struct {
	Spans []Span
}

type Block struct {
	Lines []Line
	Align Align
	Style Style
}


type SizeMode int

const (
	FixedWidth SizeMode = iota
	AutoWidth
)

type Box struct {
	Width    int
	SizeMode SizeMode
	MinWidth int

	PaddingX int
	PaddingY int
	MarginX  int
	MarginY  int

	Style  BoxStyle
	Border Border

	Blocks []Block
}

type BoxStyle struct {
	Background Color
	Border     BorderStyle
}

type Border struct {
	Chars BorderChars
	Style BorderStyle
}

type BorderChars struct {
	Horizontal   rune
	Vertical     rune
	TopLeft      rune
	TopRight     rune
	BottomLeft   rune
	BottomRight  rune
}

type BorderStyle struct {
	Color  Color
	Bold   bool
	Italic bool
}

func (s BoxStyle) Base() Style {
	return Style{Background: s.Background}
}

func (s BorderStyle) Base() Style {
	return Style{
		Foreground: s.Color,
		Bold:       s.Bold,
		Italic:    s.Italic,
	}
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

func (box *Box) AddBlock(b Block) {
	box.Blocks = append(box.Blocks, b)
}

func (box *Box) Render() string {
	if box.SizeMode == AutoWidth {
		box.Width = box.computeWidth()
	}

	if box.Width < 2 {
		box.Width = 2
	}

	inner := box.Width - 2
	var out strings.Builder

	out.WriteString(strings.Repeat("\n", box.MarginY))
	out.WriteString(strings.Repeat(" ", box.MarginX))
	out.WriteString(box.renderTop(inner))
	out.WriteString("\n")

	for i := 0; i < box.PaddingY; i++ {
		out.WriteString(strings.Repeat(" ", box.MarginX))

		out.WriteString(box.emptyLine())

		out.WriteString("\n")
	}

	for _, block := range box.Blocks {
		for _, line := range block.Lines {
			out.WriteString(strings.Repeat(" ", box.MarginX))

			out.WriteString(box.renderLine(block, line))

			out.WriteString("\n")
		}
	}

	for i := 0; i < box.PaddingY; i++ {
		out.WriteString(strings.Repeat(" ", box.MarginX))

		out.WriteString(box.emptyLine())

		out.WriteString("\n")
	}

	out.WriteString(strings.Repeat(" ", box.MarginX))
	out.WriteString(box.renderBottom(inner))
	out.WriteString(strings.Repeat("\n", box.MarginY))

	return out.String()
}

func (box *Box) renderTop(w int) string {
	s := box.Style.Base().Combine(box.Style.Border.Base()).ansi()
	return s + string(box.Border.Chars.TopLeft) +
		strings.Repeat(string(box.Border.Chars.Horizontal), w) +
		string(box.Border.Chars.TopRight) + reset()
}

func (box *Box) renderBottom(w int) string {
	s := box.Style.Base().Combine(box.Style.Border.Base()).ansi()
	return s + string(box.Border.Chars.BottomLeft) +
		strings.Repeat(string(box.Border.Chars.Horizontal), w) +
		string(box.Border.Chars.BottomRight) + reset()
}

func (box *Box) emptyLine() string {
	inner := box.Width - 2
	s := box.Style.Base().Combine(box.Style.Border.Base()).ansi()
	return s + string(box.Border.Chars.Vertical) +
		strings.Repeat(" ", inner) +
		s + string(box.Border.Chars.Vertical) + reset()
}

func (box *Box) renderLine(b Block, l Line) string {
	contentWidth := box.Width - 2 - box.PaddingX * 2
	textWidth := lineWidth(l)

	left, right := getPaddingWidth(textWidth, contentWidth, b.Align)

	var out strings.Builder

	border := box.Style.Base().Combine(box.Style.Border.Base()).ansi()
	out.WriteString(border + string(box.Border.Chars.Vertical) + reset())
	out.WriteString(box.Style.Base().ansi())
	out.WriteString(strings.Repeat(" ", box.PaddingX + left))

	for _, sp := range l.Spans {
		style := b.Style.Combine(sp.Style)

		out.WriteString(style.ansi())
		out.WriteString(sp.Text)
		out.WriteString(reset())
	}

	out.WriteString(box.Style.Base().ansi())
	out.WriteString(strings.Repeat(" ", box.PaddingX + right))
	out.WriteString(border + string(box.Border.Chars.Vertical) + reset())

	return out.String()
}

func (box *Box) computeWidth() int {
	max := 0

	for _, b := range box.Blocks {
		for _, l := range b.Lines {
			w := lineWidth(l)
			if w > max {
				max = w
			}
		}
	}

	inner := max + box.PaddingX * 2
	width := inner + 2

	if box.MinWidth > 0 && width < box.MinWidth {
		width = box.MinWidth
	}

	return width
}

func lineWidth(l Line) int {
	w := 0
	for _, s := range l.Spans {
		w += runewidth.StringWidth(s.Text)
	}
	return w
}

func getPaddingWidth(text, width int, align Align) (int, int) {
	if width <= text {
		return 0, 0
	}

	space := width - text
	switch align {
	case AlignRight:
		return space, 0
	case AlignCenter:
		left := space / 2
		return left, space - left
	default:
		return 0, space
	}
}
