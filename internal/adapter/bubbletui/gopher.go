package bubbletui

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

//go:embed gopher.jpg
var gopherJPG []byte

var (
	gopherOnce sync.Once
	gopherImg  image.Image
)

// gopherImage decodes the embedded gopher photo once and caches it. Returns nil
// if decoding fails (should not happen with the committed asset).
func gopherImage() image.Image {
	gopherOnce.Do(func() {
		img, err := jpeg.Decode(bytes.NewReader(gopherJPG))
		if err == nil {
			gopherImg = img
		}
	})
	return gopherImg
}

// fitDims returns the largest width/height (in pixels) that fits inside
// maxCols x maxRowsPx while preserving src aspect ratio. Height is forced even
// so it splits cleanly into half-block row pairs.
func fitDims(srcW, srcH, maxCols, maxRowsPx int) (w, h int) {
	if srcW <= 0 || srcH <= 0 {
		return 0, 0
	}
	s := float64(maxCols) / float64(srcW)
	if sh := float64(maxRowsPx) / float64(srcH); sh < s {
		s = sh
	}
	w = int(float64(srcW) * s)
	h = int(float64(srcH) * s)
	if w < 1 {
		w = 1
	}
	if h < 2 {
		h = 2
	}
	if h%2 == 1 {
		h--
	}
	return w, h
}

// scaleNearest resamples src to exactly w x h using nearest-neighbor sampling.
func scaleNearest(src image.Image, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	sb := src.Bounds()
	for y := 0; y < h; y++ {
		sy := sb.Min.Y + y*sb.Dy()/h
		for x := 0; x < w; x++ {
			sx := sb.Min.X + x*sb.Dx()/w
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

// halfBlockCell renders one terminal cell as an upper half-block whose
// foreground is the top pixel and background is the bottom pixel.
func halfBlockCell(top, bottom color.Color) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(hexOf(top))).
		Background(lipgloss.Color(hexOf(bottom))).
		Render("▀")
}

func hexOf(c color.Color) string {
	r, g, b, _ := c.RGBA() // 16-bit per channel
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}

// halfBlocks converts img into rows of half-block cells. Each cell packs two
// vertical pixels; source rows are consumed in pairs. A dangling final row
// (odd height) is dropped.
func halfBlocks(img image.Image) string {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	var rows []string
	for y := 0; y+1 < h; y += 2 {
		var sb strings.Builder
		for x := 0; x < w; x++ {
			top := img.At(b.Min.X+x, b.Min.Y+y)
			bottom := img.At(b.Min.X+x, b.Min.Y+y+1)
			sb.WriteString(halfBlockCell(top, bottom))
		}
		rows = append(rows, sb.String())
	}
	return strings.Join(rows, "\n")
}
