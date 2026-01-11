package pretty

import (
	"io"
	"strings"

	"github.com/spakin/netpbm"
)

type PBMImage struct {
	Img netpbm.Image
}

func (img *PBMImage) Width() int {
	return img.Img.Bounds().Dx()
}

func (img *PBMImage) Height() int {
	return img.Img.Bounds().Dy()
}

func (img *PBMImage) RenderRow(row int, base Style) string {
	var out strings.Builder
	b := img.Img.Bounds()

	for x := b.Min.X; x < b.Max.X; x++ {
		r, g, b, _ := img.Img.At(x, row).RGBA()

		out.WriteString(
			base.Combine(Style{
				Foreground: RGBColor(
					uint8(r>>8),
					uint8(g>>8),
					uint8(b>>8),
				),
			}).ansi(),
		)
		out.WriteRune('█')
		out.WriteString(reset())
	}

	return out.String()
}


func LoadPBM(r io.Reader) (Image, error) {
	img, err := netpbm.Decode(r, nil)
	if err != nil {
		return nil, err
	}
	return &PBMImage{Img: img}, nil
}

func LoadPGM(r io.Reader) (Image, error) {
	img, err := netpbm.Decode(r, nil)
	if err != nil {
		return nil, err
	}
	return &PBMImage{Img: img}, nil
}

func LoadPPM(r io.Reader) (Image, error) {
	img, err := netpbm.Decode(r, nil)
	if err != nil {
		return nil, err
	}
	return &PBMImage{Img: img}, nil
}
