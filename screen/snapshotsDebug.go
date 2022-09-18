package screen

import (
	"image"
	"image/color"
)

// SnapshotParts the currently visible screen
func SnapshotParts(vs VideoSource, screenMode int) *image.RGBA {
	snapScreen := snapshotByMode(vs, VideoText40, screenMode)
	snapPage2 := snapshotByMode(vs, VideoText40|VideoSecondPage, screenMode)
	snapPage3 := snapshotByMode(vs, VideoText80II4, screenMode)
	snapPage4 := snapshotByMode(vs, VideoII4GR, screenMode)
	snapPage5 := snapshotByMode(vs, VideoGR|VideoSecondPage, screenMode)
	snapPage6 := snapshotByMode(vs, VideoHGR, screenMode)
	return mixSixSnapshots([]*image.RGBA{snapScreen, snapPage2, snapPage3, snapPage4, snapPage5, snapPage6})

	/*
	var snapAux *image.RGBA
	videoMode := vs.GetCurrentVideoMode()
	isSecondPage := (videoMode & VideoSecondPage) != 0
	videoBase := videoMode & VideoBaseMask
	mixMode := videoMode & VideoMixTextMask
	modifiers := videoMode & VideoModifiersMask

	if videoBase == videoRGBMix {
		_, mask := snapshotDoubleHiResModeMono(a, isSecondPage, true /*isRGBMixMode*/ /*, color.White)
		snapAux = filterMask(mask)
	}

	if videoBase == VideoText40RGB {
		snapAux = snapshotText40RGBColors(vs, isSecondPage)
	} else {
		switch mixMode {
		case VideoMixText80:
			snapAux = snapshotByMode(vs, VideoText80|modifiers, screenMode)
		case VideoMixText40RGB:
			snapAux = snapshotByMode(vs, VideoText40RGB|modifiers, screenMode)
		default:
			snapAux = snapshotByMode(vs, VideoText40|modifiers, screenMode)
		}
	}

	return mixFourSnapshots([]*image.RGBA{snapScreen, snapAux, snapPage1, snapPage2})
	*/

}

// VideoModeName returns the name of the current video mode
func VideoModeName(vs VideoSource) string {
	videoMode := vs.GetCurrentVideoMode()
	videoBase := videoMode & VideoBaseMask
	mixMode := videoMode & VideoMixTextMask

	var name string

	switch videoBase {
	case VideoText40:
		name = "TEXT40COL"
	case VideoText80:
		name = "TEXT80COL"
	case VideoText40RGB:
		name = "TEXT40COLRGB"
	case VideoGR:
		name = "GR"
	case VideoDGR:
		name = "DGR"
	case VideoHGR:
		name = "HGR"
	case VideoDHGR:
		name = "DHGR"
	case VideoMono560:
		name = "Mono560"
	case VideoRGBMix:
		name = "RGBMIX"
	case VideoRGB160:
		name = "RGB160"
	case VideoSHR:
		name = "SHR"
	case VideoVidex:
		name = "VIDEX"
	case VideoText80II4:
		name = "TEXT80COL(II4)"
	case VideoII4GR:
		name = "II4GR"
	default:
		name = "Unknown video mode"
	}

	if (videoMode & VideoSecondPage) != 0 {
		name += "-PAGE2"
	}

	if (videoMode & VideoAltText) != 0 {
		name += "-ALT"
	}

	switch mixMode {
	case VideoMixText40:
		name += "-MIX40"
	case VideoMixText80:
		name += "-MIX80"
	case VideoMixText40RGB:
		name += "-MIX40RGB"
	}

	return name
}

func mixFourSnapshots(snaps []*image.RGBA) *image.RGBA {
	width := snaps[0].Rect.Dx()
	height := snaps[0].Rect.Dy()
	size := image.Rect(0, 0, width*2, height*2)
	out := image.NewRGBA(size)

	for i := 1; i < 4; i++ {
		if snaps[i].Bounds().Dx() < width {
			snaps[i] = doubleWidthFilter(snaps[i])
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			out.Set(x, y, snaps[0].At(x, y))
			out.Set(x+width, y, snaps[1].At(x, y))
			out.Set(x, y+height, snaps[2].At(x, y))
			out.Set(x+width, y+height, snaps[3].At(x, y))
		}
	}

	return out
}

func mixSixSnapshots(snaps []*image.RGBA) *image.RGBA {
	width := snaps[0].Rect.Dx()
	height := snaps[0].Rect.Dy()
	size := image.Rect(0, 0, width*3, height*2)
	out := image.NewRGBA(size)

	for i := 1; i < 4; i++ {
		if snaps[i].Bounds().Dx() < width {
			snaps[i] = doubleWidthFilter(snaps[i])
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			out.Set(x, y, snaps[0].At(x, y))
			out.Set(x+width, y, snaps[1].At(x, y))
			out.Set(x+width*2, y, snaps[2].At(x, y))
			out.Set(x, y+height, snaps[3].At(x, y))
			out.Set(x+width, y+height, snaps[4].At(x, y))
			out.Set(x+width*2, y+height, snaps[5].At(x, y))
		}
	}

	return out
}

func doubleWidthFilter(in *image.RGBA) *image.RGBA {
	b := in.Bounds()
	size := image.Rect(0, 0, 2*b.Dx(), b.Dy())
	out := image.NewRGBA(size)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := in.At(x, y)
			out.Set(2*x, y, c)
			out.Set(2*x+1, y, c)
		}
	}
	return out
}

// SnapshotCharacterGenerator shows the current character set
func SnapshotCharacterGenerator(vs VideoSource, isAltText bool) *image.RGBA {
	text := make([]uint8, textLines*text40Columns)
	for l := 0; l < textLines; l++ {
		for c := 0; c < text40Columns; c++ {
			text[text40Columns*l+c] = 0x20 + 0x80 // Space
		}
	}

	for l := 0; l < 8; l++ {
		for c := 0; c < 32; c++ {
			text[text40Columns*(2*l+4)+c+4] = uint8(l*32 + c)
		}
	}

	return renderText(vs, text, isAltText, nil, color.White)
}
