package screen

import (
	//"fmt"
	"image"
	"image/color"
)

const (
	ii4ResWidth       = 640
	ii4ResLineBytes   = ii4ResWidth / 8
	ii4ResHeight      = 384
)

func snapshotII4GR(vs VideoSource, light color.Color) *image.RGBA {
	data := vs.GetII4VideoMemory()
	return renderiII4GR(data, light)
}

func renderiII4GR(data []uint8, light color.Color) *image.RGBA {
	// No Woz-esque/II-esque interlacing, and square(ish) pixels
	size := image.Rect(0, 0, ii4ResWidth, ii4ResHeight)
	img := image.NewRGBA(size)

	for y := 0; y < ii4ResHeight; y++ {
		offset := y * ii4ResLineBytes
		for x := 0; x < ii4ResWidth; x += 8 {	// 8 pixels per byte
			pixels := data[offset + (x/8)]
			for p := 0; p < 8; p++ {
				bit := (pixels >> p) & 0x01
				if bit == 0 {
					img.Set(x+p, y, color.Black)
				} else {
					img.Set(x+p, y, light)
				}
			}
		}
	}

	return img
}

func SizeOfII4GR() uint32 {
	return ii4ResLineBytes * ii4ResHeight
}

