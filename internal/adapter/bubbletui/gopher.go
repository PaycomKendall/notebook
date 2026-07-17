package bubbletui

import (
	"bytes"
	_ "embed"
	"image"
	"image/jpeg"
	"sync"
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
