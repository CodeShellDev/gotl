package pretty

import (
	"fmt"
	"regexp"
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
	R, G, B	uint8
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
	Foreground		Color
	Background		Color
	Bold   			bool
	Italic			bool
	Underline		bool
	Strikethrough	bool
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
	if s.Underline {
		codes = append(codes, "4")
	}
	if s.Strikethrough {
		codes = append(codes, "9")
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
	Text	string
	Style	Style
}

func (s Span) Width() int { return runewidth.StringWidth(s.Text) }

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


type Segment interface {
	Width() int
	Render(base Style, width int) []string
}


type InlineSegment struct{ Items []Inline }

func (seg InlineSegment) Width() int {
	w := 0

	for _, item := range seg.Items {
		w += item.Width()
	}

	return w
}

func (seg InlineSegment) Render(base Style, width int) []string {
	var sb strings.Builder

	for _, item := range seg.Items {
		sb.WriteString(item.Render(base))
	}

	return []string{sb.String()}
}


type TextBlockSegment struct {
	Text	string
	Style	Style
}

func (s TextBlockSegment) Width() int {
	max := 0
	for line := range strings.SplitSeq(s.Text, "\n") {
		w := runewidth.StringWidth(line)
		if w > max {
			max = w
		}
	}
	return max
}

func (s TextBlockSegment) Render(base Style, width int) []string {
	lines := strings.Split(s.Text, "\n")

	result := make([]string, len(lines))

	st := base.Combine(s.Style).ansi()

	for i, l := range lines {
		result[i] = st + l + reset()
	}

	return result
}


type StyledTextBlockSegment struct {
	Raw	string
}

var tagRe = regexp.MustCompile(`\{/?[^\}]*\}`)

func (s StyledTextBlockSegment) Width() int {
	m := 0
	lines := strings.SplitSeq(s.Raw, "\n")
	for line := range lines {
		// strip all { } tags
		cleaned := tagRe.ReplaceAllString(line, "")

		w := visibleWidth(cleaned)

		m = max(m, w)
	}

	return m
}

var NamedColors = map[string]Color{
	// Standard 8
	"black":   Basic(Black),
	"red":     Basic(Red),
	"green":   Basic(Green),
	"yellow":  Basic(Yellow),
	"blue":    Basic(Blue),
	"magenta": Basic(Magenta),
	"cyan":    Basic(Cyan),
	"white":   Basic(White),

	// Bright 8
	"bright_black":   Basic(BrightBlack),
	"bright_red":     Basic(BrightRed),
	"bright_green":   Basic(BrightGreen),
	"bright_yellow":  Basic(BrightYellow),
	"bright_blue":    Basic(BrightBlue),
	"bright_magenta": Basic(BrightMagenta),
	"bright_cyan":    Basic(BrightCyan),
	"bright_white":   Basic(BrightWhite),

	// Extended colors (RGB)
	"orange":  RGBColor(255, 165, 0),
	"pink":    RGBColor(255, 192, 203),
	"purple":  RGBColor(128, 0, 128),
	"teal":    RGBColor(0, 128, 128),
	"lime":    RGBColor(0, 255, 0),
	"navy":    RGBColor(0, 0, 128),
	"maroon":  RGBColor(128, 0, 0),
	"olive":   RGBColor(128, 128, 0),
	"silver":  RGBColor(192, 192, 192),
	"gray":    RGBColor(128, 128, 128),
	"fuchsia": RGBColor(255, 0, 255),
	"aqua":    RGBColor(0, 255, 255),
	"gold":    RGBColor(255, 215, 0),
}

func parseStyleColor(color string) Color {
	switch {
	case strings.HasPrefix(color, "rgb(") && strings.HasSuffix(color, ")"):
		var r, g, b uint8

		fmt.Sscanf(color, "rgb(%d,%d,%d)", &r, &g, &b)

		return RGBColor(r, g, b)
	case strings.HasPrefix(color, "ansi(") && strings.HasSuffix(color, ")"):
		var n int

		fmt.Sscanf(color, "ansi(%d)", &n)

		return ANSI256(n)
	default:
		return NamedColors[color]
	}
}

func parseStyleTag(tag string) Style {
	s := Style{}

	parts := strings.Split(tag, ",")

	for _, p := range parts {
		p = strings.TrimSpace(p)
		switch p {
		case "/":
			return s
		case "b", "bold":
			s.Bold = true
		case "i", "italic":
			s.Italic = true
		case "u", "underline":
			s.Underline = true
		case "s", "strikethrough":
			s.Strikethrough = true
		default:
			switch {
			case strings.HasPrefix(p, "fg="):
				s.Foreground = parseStyleColor(p[3:])
			case strings.HasPrefix(p, "bg="):
				s.Background = parseStyleColor(p[3:])
			}
		}
	}

	return s
}

func renderStyledText(input string) string {
	var out strings.Builder
	stack := []Style{}

	for i := 0; i < len(input); {
		if input[i] == '{' {
			if i + 2 < len(input) && input[i + 1] == '/' && input[i + 2] == '}' {
				if len(stack) > 0 {
					stack = stack[:len(stack)-1]
				}

				out.WriteString(reset())

				for _, s := range stack {
					out.WriteString(s.ansi())
				}

				i += 3
				continue
			}

			closingIndex := strings.IndexByte(input[i:], '}')
			if closingIndex == -1 {
				// invalid tag -> take literal {
				out.WriteByte(input[i])

				i++
				continue
			}

			tag := input[i + 1 : i + closingIndex]
			
			style := parseStyleTag(tag)
			stack = append(stack, style)

			out.WriteString(style.ansi())

			i += closingIndex + 1
		} else {
			out.WriteByte(input[i])
			i++
		}
	}

	out.WriteString(reset())
	return out.String()
}


func (s StyledTextBlockSegment) Render(base Style, width int) []string {
	styled := renderStyledText(s.Raw)

	lines := strings.Split(styled, "\n")

	return lines
}


type Block struct {
	Segments []Segment
	Align    Align
	Style    Style
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
		Bold: s.Bold, 
		Italic: s.Italic,
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
	return Style{ Background: s.Background } 
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

func (box *Box) computeWidth() int {
	max := 0
	for _, block := range box.Blocks {
		for _, seg := range block.Segments {
			w := seg.Width()
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
		for _, seg := range block.Segments {
			lines := seg.Render(box.Style.Base().Combine(block.Style), inner)

			for _, line := range lines {
				out.WriteString(strings.Repeat(" ", box.MarginX))

				out.WriteString(box.renderLine(line, inner, block.Align))

				out.WriteString("\n")
			}
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
	
	return style.ansi() + string(box.Border.Chars.TopLeft) + strings.Repeat(string(box.Border.Chars.Horizontal), width) + string(box.Border.Chars.TopRight) + reset()
}

func (box *Box) renderBottom(width int) string {
	style := box.Style.Base().Combine(box.Border.Style.Base())

	return style.ansi() + string(box.Border.Chars.BottomLeft) + strings.Repeat(string(box.Border.Chars.Horizontal), width) + string(box.Border.Chars.BottomRight) + reset()
}

func (box *Box) emptyLine() string {
	style := box.Style.Base().Combine(box.Border.Style.Base())
	inner := box.Width - 2

	return style.ansi() + string(box.Border.Chars.Vertical) + strings.Repeat(" ", inner) + string(box.Border.Chars.Vertical) + reset()
}

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)

func visibleWidth(s string) int {
	clean := ansiRegexp.ReplaceAllString(s, "") 
	return runewidth.StringWidth(clean)
}

func (box *Box) renderLine(content string, width int, align Align) string {
    innerWidth := width - box.PaddingX * 2
    left, right := getPadding(visibleWidth(content), innerWidth, align)

    borderStyle := box.Style.Base().Combine(box.Border.Style.Base())

    return borderStyle.ansi() + string(box.Border.Chars.Vertical) + reset() +
        box.Style.Base().ansi() +
        strings.Repeat(" ", box.PaddingX) +
        strings.Repeat(" ", left) +
        content +
        strings.Repeat(" ", right) +
        strings.Repeat(" ", box.PaddingX) +
        borderStyle.ansi() + string(box.Border.Chars.Vertical) + reset()
}
