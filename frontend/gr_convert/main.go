package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"

	"github.com/esimov/dithergo"
)


var ditherers []dither.Dither = []dither.Dither {
	dither.Dither {
		"Atkinson",
		dither.Settings{
			[][]float32{
				[]float32{ 0.0, 0.0, 1.0 / 8.0, 1.0 / 8.0 },
				[]float32{ 1.0 / 8.0, 1.0 / 8.0, 1.0 / 8.0, 0.0 },
				[]float32{ 0.0, 1.0 / 8.0, 0.0, 0.0 },
			},
		},
	},
}

func mainD() {
    fmt.Printf("gr_convert\n")
    dither.Process(ditherers)
}



func main() {
    fmt.Printf("gr_convert\n")

    if len(os.Args) < 1 {
    	fmt.Printf("*** The THRESHOLD is missing\n")
    	return
    }
    if len(os.Args) < 2 {
    	fmt.Printf("*** The output FILENAME is missing\n")
    	return
    }
    if len(os.Args) < 3 {
    	fmt.Printf("*** The output FILENAME is missing\n")
    	return
    }

    threshholdV, errV := strconv.ParseInt(os.Args[1], 0, 16)
    if errV != nil {
        fmt.Printf("*** The THRESHOLD isn't a valid number - %s\n", errV)
        return
    }
    threshhold := uint8(threshholdV & 0xff)

    infile, err := os.Open(os.Args[2])
    if err != nil {
        fmt.Printf("*** The FILE isn't found - %s\n", err)
        return
    }
    defer infile.Close()

    // Encode the grayscale image to the output file
    outfile, err := os.Create(os.Args[3] + ".png")
    if err != nil {
        fmt.Printf("*** The PNG output FILE could not be created - %s\n", err)
        return
    }
    defer outfile.Close()

    // Encode the II4 graphics image to the output file
    grfile, err := os.Create(os.Args[3] + ".gr_")
    if err != nil {
        fmt.Printf("*** The GR_ output FILE could not be created - %s\n", err)
        return
    }
    defer grfile.Close()


    // Decode will figure out what type of image is in the file on its own
    src, _, err := image.Decode(infile)
    if err != nil {
        fmt.Printf("*** The FILE isn't a valid image - %s\n", err)
        return
    }

    // Create a new grayscale image
    bounds := src.Bounds()
    w, h := bounds.Max.X, bounds.Max.Y
    fmt.Printf(" %d x %d pixels\n", w, h)
    gray := image.NewGray(image.Rect(0, 0, w, h))
	pixel := make([]uint8, 1)
    for y := 0; y < h; y++ {
	    for x := 0; x < w; x += 8 {
	    	pixel[0] = uint8(0)
	    	for p := 0; p < 8; p ++ {
	            oldColor := src.At(x+p, y)
	            grayColor := color.GrayModel.Convert(oldColor)
	            shade := grayColor.(color.Gray).Y
	            if shade < threshhold {
	            	gray.Set(x+p, y, color.Gray{0x00})
	            } else {
	            	gray.Set(x+p, y, color.Gray{0xff})
	            	pixel[0] |= 0x01 << p
	            }
	         }
	         actual, errW := grfile.Write(pixel)
	         if errW != nil {
			        fmt.Printf("*** Error writing the GR_ FILE - %s\n", err)
			        return
	         }
	         if actual != 1 {
			        fmt.Printf("*** Wrote %d bytes to the GR_ FILE\n", actual)
			        return
	         }
	    }
    }

    png.Encode(outfile, gray)
}


/*
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
*/