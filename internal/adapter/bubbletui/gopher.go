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
