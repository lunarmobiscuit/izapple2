/*
 *  go run . ii4_font Apple\ II4\ Charset.rom ii
 */

package main

import (
	"fmt"
	"image"
	"image/color"
	//"image/png"
	"os"
)

type FontTable struct {
	nglyphs		uint8
	height		uint8
	baseline	uint8
	glyphs		[256]FontEntry
}

type FontEntry struct {
	width		uint8
	offset		uint16
}

func main() {
    fmt.Printf("ii4_font\n")

    if len(os.Args) < 2 {
    	fmt.Printf("*** The input FILENAME is missing\n")
    	return
    }
    if len(os.Args) < 3 {
    	fmt.Printf("*** The output FILENAME is missing\n")
    	return
    }

    infile, err := os.Open(os.Args[2])
    if err != nil {
        fmt.Printf("*** The font INPUT FILE isn't found - %s\n", err)
        return
    }
    defer infile.Close()

    if (os.Args[2][len(os.Args[2])-4:] == ".rom") {
    	fmt.Printf("  Read the ROM\n")
    	charset := make([]uint8, 4096)
    	actual, errR := infile.Read(charset)
    	if (errR != nil) {
	        fmt.Printf("*** Problem reading the ROM FILE - %s\n", err)
	        return
    	}
    	if (actual != 4096) {
	        fmt.Printf("*** Problem reading the ROM FILE - %d bytes read\n", actual)
	        return
    	}

    	generate7x8(charset)
    	generate14x16(charset)
    	generate21x16(charset)
    	generate28x32(charset)

    } else {
	    // Encode the font to the output file
	    outfile, err := os.Create(os.Args[3] + ".fnt")
	    if err != nil {
	        fmt.Printf("*** The FNT output FILE could not be created - %s\n", err)
	        return
	    }
	    defer outfile.Close()

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
		            if shade < 0x80 {
		            	gray.Set(x+p, y, color.Gray{0x00})
		            } else {
		            	gray.Set(x+p, y, color.Gray{0xff})
		            	pixel[0] |= 0x01 << p
		            }
		         }
		         actual, errW := outfile.Write(pixel)
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
	}
}


/*
	'@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', // 0
	'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', // 8
	'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', // 16
	'X', 'Y', 'Z', '[', '\', ']', '^', '_', // 24
	' ', '!', '"', '#', '$', '%', '&', '\', // 32
	'(', ')', '*', '+', ',', '-', '.', '\\', // 40
	'0', '1', '2', '3', '4', '5', '6', '7', // 48
	'8', '9', ':', ';', '<', '=', '>', '?', // 56
	'@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', // 64 (64-95 = 0-31)
	'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', // 72
	'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', // 80
	'X', 'Y', 'Z', '[', '\', ']', '^', '_', // 88
	'`', 'a', 'b', 'c', 'd', 'e', 'f', 'g', // 96
	'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', // 104
	'p', 'q', 'r', 's', 't', 'u', 'v', 'w', // 112
	'x', 'y', 'z', '{', '|', '}', '~', 255, // 120
*/


//
// 8x8 pixel verson of charset font
//
func generate7x8(charset []uint8) {	
	fmt.Printf("  Create '%s8.fnt'\n", os.Args[3])
	outfile, err := os.Create(os.Args[3] + "8.fnt")
	if err != nil {
	    fmt.Printf("*** The FNT output FILE could not be created - %s\n", err)
	    return
	}
	defer outfile.Close()

	fmt.Printf("  Fill the FONT table\n")
	var font FontTable
	font.nglyphs = 128
	font.height = 8
	font. baseline = 7
    for c := 128; c < 256; c++ {
    	font.glyphs[c-128].width = 7
    	if (c-128 < 64) {
	    	font.glyphs[c-128].offset = (128*3)+3 + uint16((c-128)*8)
    	} else if (c-128 < 96) {
	    	font.glyphs[c-128].offset = font.glyphs[c-128-64].offset
    	} else {
	    	font.glyphs[c-128].offset = (128*3)+3 + uint16((c-128-32)*8)
    	}
    }

	fmt.Printf("  Create the x8 FNT header\n")
	headerBytes := make([]uint8, 128*3+3)
    headerBytes[0] = font.nglyphs
    headerBytes[1] = font.height
    headerBytes[2] = font. baseline
    for c := 128; c < 256; c++ {
    	headerBytes[((c-128)*3)+3] = font.glyphs[c-128].width
    	headerBytes[((c-128)*3)+4] = uint8(font.glyphs[c-128].offset & 0xff)
    	headerBytes[((c-128)*3)+5] = uint8(font.glyphs[c-128].offset >> 8 & 0xff)
    }
	actualW, errW := outfile.Write(headerBytes)
	if errW != nil {
	    fmt.Printf("*** Error writing the x8 FNT FILE - %s\n", err)
	    return
	}
	if actualW != len(headerBytes) {
	    fmt.Printf("*** Wrote %d bytes to the x8 FNT FILE\n", actualW)
	    return
	}

	fmt.Printf("  Translate the x8 pixels\n")
	pixel := make([]uint8, 1)
    for c := 128; c < 256; c++ {
    	if (c-128 < 64) ||  (c-128 >= 96) { // skip repeat chars
		    for row := 0; row < 8; row++ {
		    	pixel[0] = uint8(0)
				bits := charset[c*8+row]
		    	for p := 0; p < 8; p ++ {
		    		if (bits >> p & 0x01) != 0 {
						pixel[0] |= 0x01 << p
					}
				}
		        actualW, errW = outfile.Write(pixel)
				if errW != nil {
				    fmt.Printf("*** Error writing pixels to the FNT FILE - %s\n", err)
				    return
				}
				if actualW != 1 {
				    fmt.Printf("*** Wrote %d pixel bytes to the FNT FILE\n", actualW)
				    return
				}
			}
		}
	}
}


//
// 16x16 pixel verson of charset font
//
func generate14x16(charset []uint8) {
	fmt.Printf("  Create '%s16.fnt'\n", os.Args[3])
	outfile, err := os.Create(os.Args[3] + "16.fnt")
	if err != nil {
	    fmt.Printf("*** The FNT output FILE could not be created - %s\n", err)
	    return
	}
	defer outfile.Close()

	fmt.Printf("  Fill the FONT table\n")
	var font FontTable
	font.nglyphs = 128
	font.height = 16
	font. baseline = 14
    for c := 128; c < 256; c++ {
    	font.glyphs[c-128].width = 13
    	if (c-128 < 64) {
	    	font.glyphs[c-128].offset = (128*3)+3 + uint16((c-128)*(16*2))
    	} else if (c-128 < 96) {
	    	font.glyphs[c-128].offset = font.glyphs[c-128-64].offset
    	} else {
	    	font.glyphs[c-128].offset = (128*3)+3 + uint16((c-128-32)*(16*2))
    	}
    }

	fmt.Printf("  Create the x16 FNT header\n")
	headerBytes := make([]uint8, 128*3+3)
    headerBytes[0] = font.nglyphs
    headerBytes[1] = font.height
    headerBytes[2] = font. baseline
    for c := 128; c < 256; c++ {
    	headerBytes[((c-128)*3)+3] = font.glyphs[c-128].width
    	headerBytes[((c-128)*3)+4] = uint8(font.glyphs[c-128].offset & 0xff)
    	headerBytes[((c-128)*3)+5] = uint8(font.glyphs[c-128].offset >> 8 & 0xff)
    }
	actualW, errW := outfile.Write(headerBytes)
	if errW != nil {
	    fmt.Printf("*** Error writing the x16 FNT FILE - %s\n", err)
	    return
	}
	if actualW != len(headerBytes) {
	    fmt.Printf("*** Wrote %d bytes to the x16 FNT FILE\n", actualW)
	    return
	}

	fmt.Printf("  Translate the x16 pixels\n")
	pixels := make([]uint8, 2)
	for c := 128; c < 256; c++ {
    	if (c-128 < 64) ||  (c-128 >= 96) { // skip repeat chars
		    for row := 0; row < 8; row++ {
				word := uint16(0)
				bits := charset[c*8+row]
		    	for p := 0; p < 8; p ++ {
		    		if (bits >> p & 0x01) != 0 {
						word |= 0x03 << (p * 2)
					}
				}
				// two bytes per row
				pixels[0] = uint8(word & 0xff)
				pixels[1] = uint8((word >> 8) & 0xff)

		        actualW, errW = outfile.Write(pixels)
		        actualW, errW = outfile.Write(pixels)
				if errW != nil {
				    fmt.Printf("*** Error writing pixels to the FNT FILE - %s\n", err)
				    return
				}
				if actualW != 2 {
				    fmt.Printf("*** Wrote %d pixel bytes to the FNT FILE\n", actualW)
				    return
				}
			}
		}
	}
}


//
// 24x24 pixel verson of charset font
//
func generate21x16(charset []uint8) {
	fmt.Printf("  Create '%s24.fnt'\n", os.Args[3])
	outfile, err := os.Create(os.Args[3] + "24.fnt")
	if err != nil {
	    fmt.Printf("*** The FNT output FILE could not be created - %s\n", err)
	    return
	}
	defer outfile.Close()

	fmt.Printf("  Fill the FONT table\n")
	var font FontTable
	font.nglyphs = 128
	font.height = 24
	font. baseline = 21
    for c := 128; c < 256; c++ {
    	font.glyphs[c-128].width = 21
    	if (c-128 < 64) {
	    	font.glyphs[c-128].offset = (128*3)+3 + uint16((c-128)*(24*3))
    	} else if (c-128 < 96) {
	    	font.glyphs[c-128].offset = font.glyphs[c-128-64].offset
    	} else {
	    	font.glyphs[c-128].offset = (128*3)+3 + uint16((c-128-32)*(24*3))
    	}
    }

	fmt.Printf("  Create the x24 FNT header\n")
	headerBytes := make([]uint8, 128*3+3)
    headerBytes[0] = font.nglyphs
    headerBytes[1] = font.height
    headerBytes[2] = font. baseline
    for c := 128; c < 256; c++ {
    	headerBytes[((c-128)*3)+3] = font.glyphs[c-128].width
    	headerBytes[((c-128)*3)+4] = uint8(font.glyphs[c-128].offset & 0xff)
    	headerBytes[((c-128)*3)+5] = uint8(font.glyphs[c-128].offset >> 8 & 0xff)
    }
	actualW, errW := outfile.Write(headerBytes)
	if errW != nil {
	    fmt.Printf("*** Error writing the x24 FNT FILE - %s\n", err)
	    return
	}
	if actualW != len(headerBytes) {
	    fmt.Printf("*** Wrote %d bytes to the x24 FNT FILE\n", actualW)
	    return
	}

	fmt.Printf("  Translate the x24 pixels\n")
	pixels := make([]uint8, 3)
	for c := 128; c < 256; c++ {
    	if (c-128 < 64) ||  (c-128 >= 96) { // skip repeat chars
		    for row := 0; row < 8; row++ {
				word := uint32(0)
				bits := charset[c*8+row]
		    	for p := 0; p < 8; p ++ {
		    		if (bits >> p & 0x01) != 0 {
						word |= 0x07 << (p * 3)
					}
				}
				// three bytes per row
				pixels[0] = uint8(word & 0xff)
				pixels[1] = uint8((word >> 8) & 0xff)
				pixels[2] = uint8((word >> 16) & 0xff)

				// three identical rows as we blow up the original pixels to 3x3
		        actualW, errW = outfile.Write(pixels)
		        actualW, errW = outfile.Write(pixels)
		        actualW, errW = outfile.Write(pixels)
				if errW != nil {
				    fmt.Printf("*** Error writing pixels to the FNT FILE - %s\n", err)
				    return
				}
				if actualW != 3 {
				    fmt.Printf("*** Wrote %d pixel bytes to the FNT FILE\n", actualW)
				    return
				}
			}
		}
	}
}


//
// 32x32 pixel verson of charset font
//
func generate28x32(charset []uint8) {
	fmt.Printf("  Create '%s32.fnt'\n", os.Args[3])
	outfile, err := os.Create(os.Args[3] + "32.fnt")
	if err != nil {
	    fmt.Printf("*** The FNT output FILE could not be created - %s\n", err)
	    return
	}
	defer outfile.Close()

	fmt.Printf("  Fill the FONT table\n")
	var font FontTable
	font.nglyphs = 128
	font.height = 24
	font. baseline = 21
    for c := 128; c < 256; c++ {
    	font.glyphs[c-128].width = 28
    	if (c-128 < 64) {
	    	font.glyphs[c-128].offset = (128*3)+3 + uint16((c-128)*(32*4))
    	} else if (c-128 < 96) {
	    	font.glyphs[c-128].offset = font.glyphs[c-128-64].offset
    	} else {
	    	font.glyphs[c-128].offset = (128*3)+3 + uint16((c-128-32)*(32*4))
    	}
    }

	fmt.Printf("  Create the x32 FNT header\n")
	headerBytes := make([]uint8, 128*3+3)
    headerBytes[0] = font.nglyphs
    headerBytes[1] = font.height
    headerBytes[2] = font. baseline
    for c := 128; c < 256; c++ {
    	headerBytes[((c-128)*3)+3] = font.glyphs[c-128].width
    	headerBytes[((c-128)*3)+4] = uint8(font.glyphs[c-128].offset & 0xff)
    	headerBytes[((c-128)*3)+5] = uint8(font.glyphs[c-128].offset >> 8 & 0xff)
    }
	actualW, errW := outfile.Write(headerBytes)
	if errW != nil {
	    fmt.Printf("*** Error writing the x32 FNT FILE - %s\n", err)
	    return
	}
	if actualW != len(headerBytes) {
	    fmt.Printf("*** Wrote %d bytes to the x32 FNT FILE\n", actualW)
	    return
	}

	fmt.Printf("  Translate the pixels\n")
	pixels := make([]uint8, 4)
	for c := 128; c < 256; c++ {
    	if (c-128 < 64) ||  (c-128 >= 96) { // skip repeat chars
		    for row := 0; row < 8; row++ {
				word := uint32(0)
				bits := charset[c*8+row]
		    	for p := 0; p < 8; p ++ {
		    		if (bits >> p & 0x01) != 0 {
						word |= 0x0f << (p * 4)
					}
				}
				// four bytes per row
				pixels[0] = uint8(word & 0xff)
				pixels[1] = uint8((word >> 8) & 0xff)
				pixels[2] = uint8((word >> 16) & 0xff)
				pixels[3] = uint8((word >> 24) & 0xff)

				// four identical rows as we blow up the original pixels to 4x4
		        actualW, errW = outfile.Write(pixels)
		        actualW, errW = outfile.Write(pixels)
		        actualW, errW = outfile.Write(pixels)
		        actualW, errW = outfile.Write(pixels)
				if errW != nil {
				    fmt.Printf("*** Error writing pixels to the FNT FILE - %s\n", err)
				    return
				}
				if actualW != 4 {
				    fmt.Printf("*** Wrote %d pixel bytes to the FNT FILE\n", actualW)
				    return
				}
			}
		}
	}
}
