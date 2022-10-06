/*
 *  go run . ii4_font Apple\ II4\ Charset.rom ii
 *  ./ii4_font Macintosh/Times_10.dfont Times10
 *  ./ii4_font PNGs/Chicago.png Chicago; cp Chicago.fnt ../../Apple2four/src/fonts/
 */

package main

import (
	"fmt"
	"image"
	"image/png"
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
	column		uint16
	offset		uint16
}

// Structure to hold the parsed data
type dfontParser struct {
	b 			[]uint8			// file buffer
	end			int				// buffer length-1
	i			int				// index into the file

	rsrcData		uint32
	rsrcMap			uint32
	rsrcDataLength	uint32
	rsrcMapLength	uint32
}


func main() {
	fmt.Printf("ii4_font\n")

	if len(os.Args) < 1 {
		fmt.Printf("*** The input FILENAME is missing\n")
		return
	}
	if len(os.Args) < 2 {
		fmt.Printf("*** The output FILENAME is missing\n")
		return
	}

	infile, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Printf("*** The font INPUT FILE isn't found '%s' - %s\n", os.Args[2], err)
		return
	}
	defer infile.Close()

	if (os.Args[1][len(os.Args[1])-4:] == ".rom") {

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

	} else if (os.Args[1][len(os.Args[1])-6:] == ".dfont") {

		fmt.Printf("  Read the DFONT\n")
		dfont := new(dfontParser)
		dfont.b = make([]uint8, 64*1024)
		dfont.end, err = infile.Read(dfont.b)
		if (err != nil) {
			fmt.Printf("*** Problem reading the DFONT FILE - %s\n", err)
			return
		}
		fmt.Printf("  DFONT FILE is %d bytes long\n", dfont.end)

		convertDFONT(dfont)

	} else if (os.Args[1][len(os.Args[1])-4:] == ".png") {

		fmt.Printf("  Read the PNG\n")

		// Decode will figure out what type of image is in the file on its own
		img, err := png.Decode(infile)
		if err != nil {
			fmt.Printf("*** The FILE isn't a valid image - %s\n", err)
			return
		}

		convertPNG(img)

	} else {

		fmt.Printf("*** UNKNOWN font format\n")
		return

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
// Convert a PNG to the AppleII4 format
//
func convertPNG(img image.Image) {

	// Encode the font to the output file
	outfile, err := os.Create(os.Args[2] + ".fnt")
	if err != nil {
		fmt.Printf("*** The FNT output FILE could not be created - %s\n", err)
		return
	}
	defer outfile.Close()

	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	spacing := h/9
	fmt.Printf(" %d x %d pixels PNG\n", w, h)

	fmt.Printf("  Fill the FONT table\n")
	var font FontTable
	font.nglyphs = 128
	font.height = uint8(h & 0xff) - 2
	font.baseline = font.height
	for y := 0; y < h; y++ {
		px := img.At(0, y)
		shade, _, _, _ := px.RGBA()
		if (shade < 0x80) {
			font.baseline = uint8(y)
			break
		}
	}
	fmt.Printf(" %d baseline -- %d pixels between glyphs\n", font.baseline, spacing)

	fmt.Printf("  Find the GLYPH widths\n")
	g := 32 // from space (32) to apple (128)
	offset := (128*3)+3 // starting after the 3-byte header and 128 width/offset values
	for x := 1; x < w; x++ {
		px := img.At(x, 0)
		shade, _, _, _ := px.RGBA()
		if (shade < 0x80) {
			font.glyphs[g].column = uint16(x)
			for x2 := x+1; x < w; x2++ {
				px2 := img.At(x2, 1)
				shade2, _, _, _ := px2.RGBA()
				if (shade2 > 0x80) {
					font.glyphs[g].width = uint8(x2 - x)
					font.glyphs[g].offset = uint16(offset)
					glyphBytes := (((x2 - x + spacing) + 7) >> 3) * (h-2)	// bytes per row * (h-2)
					if (g == 128) {
						break
					}
					g += 1
					offset += glyphBytes
					x = x2
					break
				}
			}
		}
	}

	// The first 0-31th ASCII codes point to the 64-95th codes
	for c := 0; c < 32; c++ {
		font.glyphs[c].width = font.glyphs[c+64].width
		font.glyphs[c].offset = font.glyphs[c+64].offset
	}

	fmt.Printf("  Create the FNT header\n")
	headerBytes := make([]uint8, 128*3+3)
	headerBytes[0] = font.nglyphs
	headerBytes[1] = font.height
	headerBytes[2] = font.baseline
	for c := 0; c < 128; c++ {
		headerBytes[(c*3)+3] = font.glyphs[c].width + uint8(spacing)
		headerBytes[(c*3)+4] = uint8(font.glyphs[c].offset & 0xff)
		headerBytes[(c*3)+5] = uint8(font.glyphs[c].offset >> 8 & 0xff)
	}
	actualW, errW := outfile.Write(headerBytes)
	if errW != nil {
		fmt.Printf("*** Error writing the FNT FILE - %s\n", err)
		return
	}
	if actualW != len(headerBytes) {
		fmt.Printf("*** Wrote %d bytes to the FNT FILE\n", actualW)
		return
	}

	totalW := 3
	pixel := make([]uint8, 1)
	for g := 32; g <= 128; g++ {
		for y := 2; y < h; y++ {
			x2 := font.glyphs[g].column + uint16(font.glyphs[g].width)
			for x := font.glyphs[g].column; x < x2 + uint16(spacing); x += 8 {
				pixel[0] = 0
				for p := 0; p < 8; p ++ {
					px := img.At(int(x)+p, y)
					shade, _, _, _ := px.RGBA()
					if (int(x)+p < int(x2)) && (shade < 0x80) {
						pixel[0] |= 0x01 << p
					}
				}
				actual, errW := outfile.Write(pixel)
				if errW != nil {
					fmt.Printf("*** Error writing the FNT FILE - %s\n", err)
					return
				}
				if actual != 1 {
					fmt.Printf("*** Wrote %d bytes to the FNT FILE\n", actual)
					return
				}
				totalW += actual
			}
		}
	}
}

func reversebits(b uint8) uint8 {
	r := uint8(0)
	for p := 0; p < 8; p++ {
		if (b << p) & 0x80 != 0 {
			r |= 0x01 << p
		}
	}
	return r
}


//
// 8x8 pixel verson of charset font
//
func generate7x8(charset []uint8) {	
	fmt.Printf("  Create '%s8.fnt'\n", os.Args[2])
	outfile, err := os.Create(os.Args[2] + "8.fnt")
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
	fmt.Printf("  Create '%s16.fnt'\n", os.Args[2])
	outfile, err := os.Create(os.Args[2] + "16.fnt")
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
	fmt.Printf("  Create '%s24.fnt'\n", os.Args[2])
	outfile, err := os.Create(os.Args[2] + "24.fnt")
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
	fmt.Printf("  Create '%s32.fnt'\n", os.Args[2])
	outfile, err := os.Create(os.Args[2] + "32.fnt")
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


//
// Convert a Macintosh .dfont font to the AppleII4 format
//
func convertDFONT(dfont *dfontParser) {

	// Start of file:
	//	long	Offset to start of resource data (always 0x100)
	//	long	Offset to start of resource map
	//	long	Length of resource data (map_offset-data_offset)
	//	long	Length of resource map
	//  pad with 0s to 0x100
	dfont.rsrcData = dfont.nextLong()
	dfont.rsrcMap = dfont.nextLong()
	dfont.rsrcDataLength = dfont.nextLong()
	dfont.rsrcMapLength = dfont.nextLong()
	dfont.i = 0x100

	fmt.Printf("  DATA:%x DLEN:%x  MAP:%x MLEN:%x\n", dfont.rsrcData, dfont.rsrcDataLength, dfont.rsrcMap, dfont.rsrcMapLength)

	// Resource map
	//	(repeat the initial 16 bytes for normal resource files, or 16 bytes of 0 for dfonts)
	//	long	0
	//	short	0
	//	short	0
	//	short	Offset from start of map to start of resource types (28?)
	//	short	Offset from start of map to start of resource names
	dfont.i = int(dfont.rsrcMap)
	for i := 0; i < 16; i++ {
		skip := dfont.nextByte()
		if (skip != 0) {
			fmt.Printf("  MAP[%d] == %x, not 0\n", i, skip)
		}
	}
	skip0 := dfont.nextLong()
	skip1 := dfont.nextWord()
	skip2 := dfont.nextWord()
	if (skip0 != 0) || (skip1 != 0) || (skip2 != 0) {
		fmt.Printf("  skip0:%x skip1:%x skip2:%x\n", skip0, skip1, skip2)
	}
	typesOffset := dfont.nextWord()
	namesOffset := dfont.nextWord()
	fmt.Printf("  TYPES:%x  NAMES:%x\n", typesOffset, namesOffset)

	// Resource Types
	//	short		Number of different types-1
	//	for each type:
	//	  long		tag
	//	  short		number of resources of this type-1
	//	  short		offset to resource list
	//	end
	dfont.i = int(dfont.rsrcMap) + int(typesOffset)
	numTypes := dfont.nextWord() + 1
	fmt.Printf("  #TYPES:%d\n", numTypes)
	for i := 0; i < int(numTypes); i++ {
		rsrcTag := dfont.nextLong()
		tag := fmt.Sprintf("%c%c%c%c", uint8((rsrcTag >> 24) & 0xff), uint8((rsrcTag >> 16) & 0xff),
											uint8((rsrcTag >> 8) & 0xff), uint8(rsrcTag & 0xff))
		numberOfThese := dfont.nextWord() + 1
		rsrcOffset := uint32(dfont.nextWord())
		fmt.Printf("	TAG:%x (%s)  #:%d  OFFSET:%x\n", rsrcTag, tag, numberOfThese, rsrcOffset)

		for i := 0; i < int(numberOfThese); i++ {
			// Resource lists
			//	for each resource of the given type:
			//	  short	resource id
			//	  short	offset to name in resource name list (0xffff for none)
			//	  byte	flags
			//	  byte*3	offset from start of resource data section to this resource's data
			//	  long	0
			//	end

			remember := dfont.i
			dfont.i = int(dfont.rsrcMap) + int(typesOffset) + int(rsrcOffset)
			rsrcId := dfont.nextWord()
			rsrcNameOffset := dfont.nextWord()
			rsrcFlags := dfont.nextByte()
			rsrcStart := dfont.next24()
			rsrcZero := dfont.nextLong()
			fmt.Printf("	  ID:%x  N:%x  F:%x  S:%x  0:%x\n", rsrcId, rsrcNameOffset, rsrcFlags, rsrcStart, rsrcZero)
			dfont.i = remember

			switch tag {
			case "sfnt":
				dfont.parseSFNT(rsrcStart)
			case "NFNT":
				dfont.parseNFNT(rsrcStart)
			case "FOND":
				dfont.parseFOND(rsrcStart)
			}
		}
	}
}


//
// Parse the 'sfnt' resource
//
func (dfont *dfontParser) parseSFNT(offset uint32) {
	fmt.Printf("		SFNT\n")
}

//
// Parse the 'NFNT' resource
//
func (dfont *dfontParser) parseNFNT(offset uint32) {

	fmt.Printf("		NFNT @ %x\n", dfont.rsrcData + uint32(offset))
	remember := dfont.i
	dfont.i = int(dfont.rsrcData) + int(offset)

	nfntLength := dfont.nextLong()
	fontType := dfont.nextWord()
	firstChar := dfont.nextWord()
	lastChar := dfont.nextWord()
	widthMax := dfont.nextWord()
	kernMax := dfont.nextWord()
	descent := dfont.nextWord()
	if descent >= 0 {
		descent = 0xFFFF - descent + 1
	}
	fRectWidth := dfont.nextWord()
	fRectHeight := dfont.nextWord()
	fmt.Printf("		  LEN:0x%x/%d TYPE:%x %d-%d W:%d K:%d D:%d R:%dx%d\n",
							nfntLength, nfntLength, fontType, firstChar, lastChar, widthMax, kernMax,
							descent, fRectWidth, fRectHeight)

	glyphBaseOW := dfont.i
	glphOW := dfont.nextWord()
	fmt.Printf("		  B-OW:%x OW:%x\n", glyphBaseOW, glphOW)

	fontAscent := dfont.nextWord()
	fontDescent := dfont.nextWord()
	fontLeading := dfont.nextWord()
	fmt.Printf("		  A:%d D:%d L:%d\n", fontAscent, fontDescent, fontLeading)
	fmt.Printf("			@%x\n", dfont.i)

	fontRowWords := dfont.nextWord()
	fmt.Printf("		  #WORDS:%d  #CHARS:%d\n", fontRowWords, lastChar - firstChar + 3)
	fontImage := make([]uint16, fontRowWords * fRectHeight) // font->fontImage = calloc(font->rowWords*font->fRectHeight,sizeof(short));
	fontLocs := make([]uint16, lastChar-firstChar+3) // font->locs = calloc(font->lastChar-font->firstChar+3,sizeof(short));
	fontWidths := make([]uint16, lastChar-firstChar+3) // font->offsetWidths = calloc(font->lastChar-font->firstChar+3,sizeof(short));
	
	for i := 0; i < int(fontRowWords * fRectHeight); i++ {
		fontImage[i] = dfont.nextWord()
	}

	for i := 0; i < int(lastChar - firstChar + 3); i++ {
		fontLocs[i] = dfont.nextWord()
		//fmt.Printf("			#%02x.loc = %04x\n", i+int(firstChar), fontLocs[i])
	}

	dfont.i = int(dfont.rsrcData) + int(glyphBaseOW) + 2 * int(glphOW)
	/*fmt.Printf("		  WIDTHS @ %x = %x + %x + 2 * %x\n",
								int(dfont.rsrcData) + int(glyphBaseOW) + 2 * int(glphOW),
								int(dfont.rsrcData), int(glyphBaseOW), int(glphOW))*/
	for i := 0; i < int(lastChar - firstChar + 3); i++ {
		fontWidths[i] = dfont.nextWord()
		//fmt.Printf("			#%02x.width = %d\n", i+int(firstChar), fontWidths[i])
	}

	for i := 0; i < int(lastChar - firstChar + 3); i++ {
		//fmt.Printf("			[%02x] = %04x / %02x.%dw\n", i+int(firstChar), fontLocs[i], fontWidths[i] >> 16, fontWidths[i] & 0xff)
	}

	dfont.i = remember
}

//
// Parse the 'FOND' resource
//
func (dfont *dfontParser) parseFOND(offset uint32) {
	// Resource lists
	//	for each resource of the given type:
	//	  short	resource id
	//	  short	offset to name in resource name list (0xffff for none)
	//	  byte	flags
	//	  byte*3	offset from start of resource data section to this resource's data
	//	  long	0
	//	end

	fmt.Printf("		FOND\n")
}

func (dfont *dfontParser) todo() {
	// Resource data
	//	for each resource
	//	  long		length of this resource
	//	  byte*n	resource data
	//	end
	dfont.i = 0x100
	for dfont.i < int(dfont.rsrcMap) {
		rsrcLength := dfont.nextLong()
		fmt.Printf("  RLEN:%x\n", rsrcLength)
		startI := dfont.i

		// Resource lists
		// for each resource of the given type:
		//	  short		resource id
		//	  short		offset to name in resource name list (0xffff for none)
		//	  byte		flags
		//	  byte*3	offset from start of resource data section to this resource's data
		//	  long		0
		// end
		rsrcId := dfont.nextWord()
		rsrcNameOffset := dfont.nextWord()
		rsrcFlags := dfont.nextByte()
		rsrcStart := dfont.next24()
		rsrcZero := dfont.nextLong()

		fmt.Printf("		ID:%x NAME:%x  FLAGS:%x START:%x  ZERO:%x\n", rsrcId, rsrcNameOffset, rsrcFlags, rsrcStart, rsrcZero)


		dfont.i = startI + int(rsrcLength)
	}

	fmt.Printf("  Create '%s.fnt'\n", os.Args[2])
	outfile, err := os.Create(os.Args[2] + ".fnt")
	if err != nil {
		fmt.Printf("*** The FNT output FILE could not be created - %s\n", err)
		return
	}
	defer outfile.Close()
}

// Return values from the dfont file (big endian)
func (dfont *dfontParser) nextLong() uint32 {
	value := uint32(dfont.b[dfont.i]) << 24 | uint32(dfont.b[dfont.i+1]) << 16 |
				uint32(dfont.b[dfont.i+2]) << 8 | uint32(dfont.b[dfont.i+3])
	dfont.i += 4
	return value
}
func (dfont *dfontParser) next24() uint32 {
	value := uint32(dfont.b[dfont.i]) << 16 | uint32(dfont.b[dfont.i+1]) << 8 | uint32(dfont.b[dfont.i+2])
	dfont.i += 3
	return value
}
func (dfont *dfontParser) nextWord() uint16 {
	value := uint16(dfont.b[dfont.i]) << 8 | uint16(dfont.b[dfont.i+1])
	dfont.i += 2
	return value
}
func (dfont *dfontParser) nextByte() uint8 {
	value := dfont.b[dfont.i]
	dfont.i += 1
	return value
}

