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


type Inline interface {
	Width() int
	Render(base Style) string
}


type Span struct {
	Text  string
	Style Style
}

func (s Span) Lines() []Span {
	lines := strings.Split(s.Text, "\n")

	out := make([]Span, len(lines))

	for i, line := range lines {
		out[i] = Span{
			Text:  line,
			Style: s.Style,
		}
	}
	return out
}

func (s Span) Width() int {
	max := 0
	for line := range strings.SplitSeq(s.Text, "\n") {
		w := runewidth.StringWidth(line)

		if w > max {
			max = w
		}
	}
	return max
}

func (s Span) Render(base Style) string {
	style := base.Combine(s.Style)
	return style.ansi() + s.Text + reset()
}


type Align int

const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)

type Segment struct {
	Items []Inline
}

type Block struct {
	Segments []Segment
	Align Align
	Style Style
}


type SizeMode int

const (
	FixedWidth SizeMode = iota
	AutoWidth
)

type BorderChars struct {
	Horizontal  rune
	Vertical    rune
	TopLeft     rune
	TopRight    rune
	BottomLeft  rune
	BottomRight rune
}

type BorderStyle struct {
	Color  Color
	Bold   bool
	Italic bool
}

func (s BorderStyle) Base() Style {
	return Style{
		Foreground: s.Color,
		Bold:       s.Bold,
		Italic:    s.Italic,
	}
}

type Border struct {
	Chars BorderChars
	Style BorderStyle
}

type BoxStyle struct {
	Background Color
}

func (s BoxStyle) Base() Style {
	return Style{Background: s.Background}
}

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
	b := NewBox(0)
	b.SizeMode = AutoWidth
	return b
}

func (box *Box) AddBlock(block Block) {
	box.Blocks = append(box.Blocks, block)
}


func lineWidth(segment Segment) int {
	w := 0

	for _, it := range segment.Items {
		w += it.Width()
	}

	return w
}

func (box *Box) computeWidth() int {
	max := 0

	for _, block := range box.Blocks {
		for _, segment := range block.Segments {
			w := lineWidth(segment)
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
		for _, segments := range block.Segments {
			out.WriteString(strings.Repeat(" ", box.MarginX))

			out.WriteString(box.renderLine(segments, block))

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

func (box *Box) renderTop(width int) string {
	style := box.Style.Base().Combine(box.Border.Style.Base())

	return style.ansi() +
		string(box.Border.Chars.TopLeft) +
		strings.Repeat(string(box.Border.Chars.Horizontal), width) +
		string(box.Border.Chars.TopRight) +
		reset()
}

func (box *Box) renderBottom(width int) string {
	style := box.Style.Base().Combine(box.Border.Style.Base())

	return style.ansi() +
		string(box.Border.Chars.BottomLeft) +
		strings.Repeat(string(box.Border.Chars.Horizontal), width) +
		string(box.Border.Chars.BottomRight) +
		reset()
}

func (box *Box) emptyLine() string {
	style := box.Style.Base().Combine(box.Border.Style.Base())
	inner := box.Width - 2

	return style.ansi() +
		string(box.Border.Chars.Vertical) +
		strings.Repeat(" ", inner) +
		string(box.Border.Chars.Vertical) +
		reset()
}

func (box *Box) renderLine(segment Segment, block Block) string {
	var lines [][]Inline
	currentLine := []Inline{}

	for _, item := range segment.Items {
		span, ok := item.(Span)

		if ok {
			for i, lineSpan := range span.Lines() {
				if i == 0 {
					currentLine = append(currentLine, lineSpan)
				} else {
					lines = append(lines, currentLine)
					currentLine = []Inline{lineSpan}
				}
			}
		} else {
			currentLine = append(currentLine, item)
		}
	}

	lines = append(lines, currentLine)

	var out strings.Builder

	for i, line := range lines {
		inner := box.Width - 2 - box.PaddingX*2
		contentWidth := 0

		for _, it := range line {
			contentWidth += it.Width()
		}
		
		left, right := getPadding(contentWidth, inner, block.Align)

		borderStyle := box.Style.Base().Combine(box.Border.Style.Base())
		baseStyle := box.Style.Base().Combine(block.Style)

		out.WriteString(borderStyle.ansi())
		out.WriteRune(box.Border.Chars.Vertical)
		out.WriteString(reset())

		out.WriteString(baseStyle.ansi())
		out.WriteString(strings.Repeat(" ", box.PaddingX+left))

		for _, item := range line {
			out.WriteString(item.Render(baseStyle))
		}

		out.WriteString(strings.Repeat(" ", right+box.PaddingX))
		out.WriteString(borderStyle.ansi())
		out.WriteRune(box.Border.Chars.Vertical)
		out.WriteString(reset())

		if i < len(lines)-1 {
			out.WriteString("\n")
		}
	}

	return out.String()
}

func getPadding(content, width int, align Align) (int, int) {
	if width <= content {
		return 0, 0
	}

	space := width - content

	switch align {
	case AlignLeft:
		return 0, space
	case AlignRight:
		return space, 0
	case AlignCenter:
		left := space / 2
		return left, space - left
	default:
		return 0, space
	}
}