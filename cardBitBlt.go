package izapple2

import (
	"fmt"
	"strconv"
)

const (
	BITBLT_DMA = iota
	BITBLT_COPY
	BITBLT_TILE
	BITBLT_FILL
	BITBLT_LINE
	BITBLT_CHAR
	BITBLT_STRING
	BITBLT_CHAR_BIG
)

const (
	BITBLT_SUCCESS = iota
)

const GRAPHICS_BASE			= 0x4000
const GRAPHICS_END			= 0xB800
const GRAPHICS_WIDTH		= 512
const GRAPHICS_HEIGHT		= 384
const GRAPHICS_ROW_BYTES	= GRAPHICS_WIDTH/8 // 64
const GRAPHICS_TOTAL_BYTES	= GRAPHICS_ROW_BYTES*GRAPHICS_ROW_BYTES // 24576

const ROMFONTS				= 0xFFA000
const RAMFONTS				= 0xA000


type FontTable struct {
	nglyphs		uint8
	height		uint8
	baseline	uint8
	glyphs		[256]FontEntry
	address		uint32
}

type FontEntry struct {
	width		uint8
	offset		uint16
}


/*
 ***  Neither a hard disk nor a floppy, but a faux storage device that uses the emulator computer's hard disk
 */

// CardBitBlt represents a graphics bitblt card
type CardBitBlt struct {
	cardBase

	trace		bool

	cmd			uint8				// Which command to run
	param0		uint8
	param1		uint8
	param2		uint32
	param3		uint32
	param4		uint32
	param5		uint32
	running		bool

	c800		[0x800]uint8

	fontID		uint8
	font		FontTable
}

// NewCardBitBlt creates a new BITBLT card
func NewCardBitBlt() *CardBitBlt {
	var c CardBitBlt
	c.romC8xx = &c
	return &c
}

// GetInfo returns information about the card
func (c *CardBitBlt) GetInfo() map[string]string {
	info := make(map[string]string)
	info["trace"] = strconv.FormatBool(c.trace)
	return info
}

//
//  Set up the softswitches
//
func (c *CardBitBlt) assign(a *Apple2, slot int) {
	c.loadRom(buildBITBLTRom(slot))

	c.addCardSoftSwitchW(0, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardBitBlt] DO COMMAND $%d\n", c.cmd)
		}

		c.cmd = value
		c.running = true

		switch (c.cmd) {
		case BITBLT_DMA: c.BitBltDMA()
		case BITBLT_COPY: c.BitBltCopy()
		case BITBLT_TILE: c.BitBltTile()
		case BITBLT_FILL: c.BitBltFill()
		case BITBLT_LINE: c.BitBltLine()
		case BITBLT_CHAR: c.BitBltChar()
		case BITBLT_STRING: c.BitBltString()
		case BITBLT_CHAR_BIG: c.BitBltCharBig()
		}

		c.running = false
		return
	}, "BITBLTCMD")

	c.addCardSoftSwitchW(1, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardBitBlt] Param0 $%02x\n", value)
		}
		c.param0 = value
		return
	}, "BITBLTPARAM0")

	c.addCardSoftSwitchW(2, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardBitBlt] Param1 $%02x\n", value)
		}
		c.param1 = value
		return
	}, "BITBLTPARAM2")

	c.addCardSoftSwitchW(3, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardBitBlt] Param2.l $%02x\n", value)
		}
		c.param2 = uint32(value)
		return
	}, "BITBLTPARAM2L")
	c.addCardSoftSwitchW(4, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardBitBlt] Param2.m $%02x\n", value)
		}
		c.param2 |= uint32(value) << 8 
		return
	}, "BITBLTPARAM2M")
	c.addCardSoftSwitchW(5, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardBitBlt] Param2.h $%02x\n", value)
		}
		c.param2 |= uint32(value) << 16 
		return
	}, "BITBLTPARAM2H")

	c.addCardSoftSwitchW(6, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardBitBlt] Param3.l $%02x\n", value)
		}
		c.param3 = uint32(value)
		return
	}, "BITBLTPARAM3L")
	c.addCardSoftSwitchW(7, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardBitBlt] Param3.m $%02x\n", value)
		}
		c.param3 |= uint32(value) << 8 
		return
	}, "BITBLTPARAM3M")
	c.addCardSoftSwitchW(8, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardBitBlt] Param3.h $%02x\n", value)
		}
		c.param3 |= uint32(value) << 16 
		return
	}, "BITBLTPARAM3H")

	c.addCardSoftSwitchW(9, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardBitBlt] Param4.l $%02x\n", value)
		}
		c.param4 = uint32(value)
		return
	}, "BITBLTPARAM4L")
	c.addCardSoftSwitchW(10, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardBitBlt] Param4.m $%02x\n", value)
		}
		c.param4 |= uint32(value) << 8 
		return
	}, "BITBLTPARAM4M")
	c.addCardSoftSwitchW(11, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardBitBlt] Param4.h $%02x\n", value)
		}
		c.param4 |= uint32(value) << 16 
		return
	}, "BITBLTPARAM4H")

	c.addCardSoftSwitchW(12, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardBitBlt] Param5.l $%02x\n", value)
		}
		c.param5 = uint32(value)
		return
	}, "BITBLTPARAM5L")
	c.addCardSoftSwitchW(13, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardBitBlt] Param5.m $%02x\n", value)
		}
		c.param5 |= uint32(value) << 8 
		return
	}, "BITBLTPARAM5M")
	c.addCardSoftSwitchW(14, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardBitBlt] Param5.h $%02x\n", value)
		}
		c.param5 |= uint32(value) << 16 
		return
	}, "BITBLTPARAM5H")

	c.addCardSoftSwitchR(14, func() uint8 {
		if c.trace {
			fmt.Printf("[CardBitBlt] Running %t\n", c.running)
		}
		if c.running {
			return 0xff
		} else {
		return 0x00
		}
	}, "BITBLTRUNNING")

	c.cardBase.assign(a, slot)
}


//
//  Do a DMA memory to memory transfer
//    param0 = 0 for low to high / !0 for high to low
//    param2 = start address
//    param3 = end address
//
func (c *CardBitBlt) BitBltDMA() {
	if c.trace {
		fmt.Printf("[CardBitBlt] BITBLT_DMA\n")
	}

	start := uint32(c.param2)
	end := uint32(c.param2)

	if (c.param0 == 0) {
		for i := uint32(0); i <= end - start; i++ {
			c.a.mmu.Poke(end + i, c.a.mmu.Peek(start + i))
		}
	} else {
		for i := end - start; i >= 0; i-- {
			c.a.mmu.Poke(end + i, c.a.mmu.Peek(start + i))
		}
	}
}

//
//  Copy a rectangle on the GRAPHICS page
//    param0 = 0 for screen to screen / !0 for memory to screen
//		screen to screen:
//		  0xC800-0xC802 = from top
//		  0xC803-0xC805 = from left
//		  0xC806-0xC808 = from bottom
//		  0xC809-0xC80B = from right
//		  param2 = to top
//		  param3 = to left
//		memory to screen:
//		  param2 = to top
//		  param3 = to left
//		  param4 = memory address (sequential bytes)
//
func (c *CardBitBlt) BitBltCopy() {
	if c.trace {
		fmt.Printf("[CardBitBlt] BITBLT_COPY\n")
	}

	if (c.param0 == 0) {
		fromTop := (uint32(c.c800[2]) << 16) | (uint32(c.c800[1]) << 8) | uint32(c.c800[0])
		fromLeft := (uint32(c.c800[5]) << 16) | (uint32(c.c800[4]) << 8) | uint32(c.c800[3])
		fromBottom := (uint32(c.c800[8]) << 16) | (uint32(c.c800[7]) << 8) | uint32(c.c800[6])
		fromRight := (uint32(c.c800[9]) << 16) | (uint32(c.c800[10]) << 8) | uint32(c.c800[11])

		toTop := c.param2
		toLeft := c.param3

		// @@@ TODO - This only correctly copies left and right mod 8
		for r := uint32(0); r <= fromBottom - fromTop; r++ {
			fromAddr := uint32(GRAPHICS_BASE) + (fromTop + r) * uint32(GRAPHICS_ROW_BYTES)
			toAddr := uint32(GRAPHICS_BASE) + (toTop + r) * uint32(GRAPHICS_ROW_BYTES)
			for i := uint32(0); i <= fromRight - fromLeft; i += 8 {
				c.a.mmu.Poke(toAddr + (toLeft + i)/8, c.a.mmu.Peek(fromAddr + (fromLeft + i)/8))
			}
		}
	} else {
		// @@@ TBD
	}
}

//
//  Tile a rectangle across the GRAPHICS page
//	  0xC800-0xC802 = width
//	  0xC803-0xC805 = height
//	  param2 = to top
//	  param3 = to left
//	  param4 = memory address (sequential bytes)
//
func (c *CardBitBlt) BitBltTile() {
	if c.trace {
		fmt.Printf("[CardBitBlt] BITBLT_TILE\n")
	}

	// @@@ TBD
}

//
//  Fill a rectangle across the GRAPHICS page
//	  param0 = 0 = black | !0 = white
//	  param2 = to top
//	  param3 = to left
//	  param4 = to bottom
//	  param5 = to right
//
func (c *CardBitBlt) BitBltFill() {
	if c.trace {
		fmt.Printf("[CardBitBlt] BITBLT_FILL\n")
	}

	color := c.param0
	if (color != 0) { color = 0xff }

	toTop := c.param2
	toLeft := c.param3
	toBottom := c.param4
	toRight := c.param5

	for y := toTop; y <= toBottom; y++ {
		toAddr := GRAPHICS_BASE + y * GRAPHICS_ROW_BYTES

		// @@@ TODO - This only correctly fills left/right mod 8
		for i := uint32(0); i <= toRight - toLeft; i += 8 {
			c.a.mmu.Poke(toAddr + (toLeft + i)/8, color)
		}
	}
}

//
//  Draw a line on the GRAPHICS page
//	  param0 = 0 = black | !0 = white // @@@ TODO
//	  param2 = to top
//	  param3 = to left
//	  param4 = to bottom
//	  param5 = to right
//
func (c *CardBitBlt) BitBltLine() {
	if c.trace {
		fmt.Printf("[CardBitBlt] BITBLT_LINE\n")
	}

	xStart := c.param2
	yStart := c.param3
	xEnd := c.param4
	yEnd := c.param5

	// Vertical line
	if (xStart == xEnd) {
		pixelBits := uint8(0x01 << (xStart % 8))
		screenMask := 1 ^ pixelBits

		for y := yStart; y <= yEnd; y++ {
			toAddr := GRAPHICS_BASE + y * GRAPHICS_ROW_BYTES

			screenBits := c.a.mmu.Peek(toAddr + (xStart/8))
			screenBits &= screenMask
			screenBits |= pixelBits
			c.a.mmu.Poke(toAddr + (xStart/8), screenBits)
		}
		return
	}

	// Horizontal line
	if (yStart == yEnd) {
fmt.Printf("HORIZ: %d,%d - %d,%d\n", xStart, yStart, xEnd, yEnd)
		for x := xStart; x <= xEnd; x++ {
			toAddr := GRAPHICS_BASE + yStart * GRAPHICS_ROW_BYTES

			screenBits := c.a.mmu.Peek(toAddr + (x/8))
			screenBits |= 0xff // @@@ TODO - partial bytes
			c.a.mmu.Poke(toAddr + (x/8), screenBits)
		}
		return
	}

	if (xStart > xEnd) {
		xStart = c.param4
		xEnd = c.param2
	}
	if (yStart > yEnd) {
		yStart = c.param5
		yEnd = c.param3
	}

	// Diagonal line
fmt.Printf("DIAG: %d,%d - %d,%d\n", xStart, yStart, xEnd, yEnd)
	slope := (yEnd - yStart) / (xEnd - xStart)
	base := uint32(yEnd - (slope * xEnd))
	for x := xStart; x <= xEnd; x++ {
		y := slope * x + base
		toAddr := GRAPHICS_BASE + y * GRAPHICS_ROW_BYTES
fmt.Printf("  x:%d y:%d @ %x\n", x, y, toAddr + (x/8))

		screenBits := c.a.mmu.Peek(toAddr + (x/8))
		screenBits |= 0xff // @@@ TODO - partial bytes
		c.a.mmu.Poke(toAddr + (x/8), screenBits)
	}

}

//
//  Draw a single glyph from a font on the GRAPHICS page
//	  param0 = ascii code
//	  param1 = fontID
//	  param2 = x
//	  param3 = y
//	Returns
//	  0xC800-0xC801 = the farthest x position
//
func (c *CardBitBlt) BitBltChar() {
	if c.trace {
		fmt.Printf("[CardBitBlt] BITBLT_CHAR '%c'\n", c.param0)
	}

	ch := c.param0
	c.fontID = c.param1
	x := uint16(c.param2)
	y := uint16(c.param3)

	c.loadFontTable(c.fontID)

	x, _ = c.bltChar(ch, x, y)

	c.c800[0] = uint8(x & 0xff)
	c.c800[1] = uint8(x >> 8 & 0xff)
}

//
//  Draw a string of glyphs from a font on the GRAPHICS page
//	  0xC800 = the string chars (zero terminated)
//	  param1 = fontID
//	  param2 = x
//	  param3 = y
//	Returns
//	  0xC800-0xC801 = the farthest x position
//	  0xC802-0xC803 = the total width
//
func (c *CardBitBlt) BitBltString() {
	if c.trace {
		fmt.Printf("[CardBitBlt] BITBLT_STRING\n")
	}

	str := c.c800toString()
	c.fontID = c.param1
	x := uint16(c.param2)
	y := uint16(c.param3)

	c.loadFontTable(c.fontID)

	var w uint8
	totalWidth := uint32(0)
	for _, ch := range str {
		x, w = c.bltChar(uint8(ch), x, y)
		totalWidth += uint32(w)
	}

	c.c800[0] = uint8(x & 0xff)
	c.c800[1] = uint8(x >> 8 & 0xff)
	c.c800[2] = uint8(totalWidth & 0xff)
	c.c800[3] = uint8(totalWidth >> 8 & 0xff)
}


//
//  Draw a single glyph from a font on the GRAPHICS page at 8x scale
//	  param0 = ascii code
//	  param1 = fontID
//	  param2 = x
//	  param3 = y
//	Returns
//	  0xC800-0xC801 = the farthest x position
//
func (c *CardBitBlt) BitBltCharBig() {
	if c.trace {
		fmt.Printf("[CardBitBlt] BITBLT_CHAR_BIG '%c'\n", c.param0)
	}

	ch := c.param0
	c.fontID = c.param1
	x := uint16(c.param2)
	y := uint16(c.param3)

	c.loadFontTable(c.fontID)

	x, _ = c.bltCharBig(ch, x, y)

	c.c800[0] = uint8(x & 0xff)
	c.c800[1] = uint8(x >> 8 & 0xff)
}

//
//  Render a character from the font onto the GRAPHICS screen
//
func (c *CardBitBlt) bltChar(ch uint8, x uint16, y uint16) (uint16, uint8) {
	// Low ASCII
	if (ch >= 128) {
		ch -= 128
	}

	// Lookup the details of this glyph
	width := uint32(c.font.glyphs[ch].width)
	stride := uint32(c.font.glyphs[ch].width + 7) / 8
	offset := uint32(c.font.glyphs[ch].offset)
	glyphsPtr := c.font.address + uint32(offset)

	// X coordinate needs to fit within the screen (no clipping haflway through a glyph)
	if (x >= GRAPHICS_WIDTH) {
		return x, uint8(width)
	} else if (x + uint16(width) >= GRAPHICS_WIDTH) { 
		return x, uint8(width)
	}

	// Iterate for each row of the glyph
	for row := uint32(0); row < uint32(c.font.height); row++ {
		// Y coordinate points to baseline, but rendering starts at the top
		if (y < uint16(c.font.baseline)) {
			continue
		} else if (y >= GRAPHICS_HEIGHT) { 
			continue
		}

		rAddr := GRAPHICS_BASE + ((uint32(y - uint16(c.font.baseline)) + row) * GRAPHICS_ROW_BYTES)
		fromAddr := glyphsPtr + (row * stride)
		toAddr := rAddr + uint32(x/8);

		// Iterate over the pixels through width
		p := 0

		// Draw the first pixels that are not aligned evenly to 8 bits
		unalignedP := int(x & 0x7)
		if unalignedP != 0 {
			glyphBits := c.a.mmu.Peek(fromAddr) << unalignedP

			screenMask := uint8(0xff >> (8 - unalignedP))

			screenBits := c.a.mmu.Peek(toAddr)
			screenBits &= screenMask
			screenBits |= glyphBits
			c.a.mmu.Poke(toAddr, screenBits)

			p = (8 - unalignedP)
			toAddr += 1
		}

		// Draw the aligned pixels
		for p + 8 < int(width) {
			glyphBits := uint16(c.a.mmu.Peek(fromAddr)) | (uint16(c.a.mmu.Peek(fromAddr + 1)) << 8)
			if (unalignedP != 0) { glyphBits >>= 8 - unalignedP }
			c.a.mmu.Poke(toAddr, uint8(glyphBits & 0xff))

			p += 8
			fromAddr += 1
			toAddr += 1
		}

		// Draw more pixels
		remaining := int(width) - p
		if (remaining > 0) {
			glyphMask := uint8(0xff >> (8-remaining))
			glyphBits := uint16(c.a.mmu.Peek(fromAddr)) | (uint16(c.a.mmu.Peek(fromAddr + 1)) << 8)
			if (unalignedP != 0) { glyphBits >>= 8 - unalignedP }

			screenMask := uint8(0xff)
			if (remaining == 8) { screenMask = 0
			} else { screenMask <<= remaining }

			screenBits := c.a.mmu.Peek(toAddr)
			screenBits &= screenMask
			screenBits |= uint8(glyphBits) & glyphMask
			c.a.mmu.Poke(toAddr, screenBits)

			p += 8
		}
	}

	return x + uint16(width), uint8(width)
}


//
//  Render a character from the font onto the GRAPHICS screen at 8x scale
//  (using y as the top-left corner, ignoring the baseline)
//  (blasting over any other pixels in the way, not aligning to rounded up x % 8)
//
func (c *CardBitBlt) bltCharBig(ch uint8, x uint16, y uint16) (uint16, uint8) {
	// Low ASCII
	if (ch >= 128) {
		ch -= 128
	}

	// Lookup the details of this glyph
	width := uint32(c.font.glyphs[ch].width)
	stride := uint32(c.font.glyphs[ch].width + 7) / 8
	offset := uint32(c.font.glyphs[ch].offset)
	glyphsPtr := c.font.address + uint32(offset)

	// Round the x coordinate up to the next nearest aligned pixel
	x = (x + 7) >> 3

	// X coordinate needs to fit within the screen (no clipping haflway through a glyph)
	if (x >= GRAPHICS_WIDTH) {
		return x, uint8(width)
	} else if (x + uint16(width * 8) >= GRAPHICS_WIDTH) { 
		return x, uint8(width * 8)
	}

	// Iterate for each row of the glyph
	for row := uint32(0); row < uint32(c.font.height * 8); row++ {
		// Stop if the glyph goes off the bottom of the screen
		if (y >= GRAPHICS_HEIGHT) { 
			break
		}

		rAddr := GRAPHICS_BASE + ((uint32(y) + row) * GRAPHICS_ROW_BYTES)
		fromAddr := glyphsPtr + (row/8 * stride)
		toAddr := rAddr + uint32(x);

		// Iterate over the pixels from 0 through width
		for p := 0; p < int(width); p++ {
			if (p != 0) && ((p % 8) == 0) {
				fromAddr += 1
			}

			glyphBits := c.a.mmu.Peek(fromAddr)
			if ((glyphBits >> (p % 8)) & 0x01 == 0) {
				c.a.mmu.Poke(toAddr, 0x00)
			} else {
				c.a.mmu.Poke(toAddr, 0xff)
			}

			toAddr += 1
		}
	}

	return x + uint16(width*8), uint8(width)
}


//
//  Load the font table into a struct
//
func (c *CardBitBlt) loadFontTable(fontID uint8) {
	// Lookup the address in the tables at ROMFONTS or RAMFONTS
	// each are arrays of 24 bits, ROM indexed from 0xFF down and RAM indexed up
	var fontAddr uint32
	if (fontID > 0x80) {
		idx := uint32(0xff - fontID)
		fontAddr = uint32(c.a.mmu.Peek(ROMFONTS + (idx*3)+2)) << 16 |
						uint32(c.a.mmu.Peek(ROMFONTS + (idx*3)+1)) << 8 |
						uint32(c.a.mmu.Peek(ROMFONTS + (idx*3)))
		if c.trace {
			fmt.Printf("[CardBitBlt] idx=%d @$%x+idx = %x\n", idx, ROMFONTS, ROMFONTS + (idx*3))
		}
	} else {
		idx := uint32(fontID)
		fontAddr = uint32(c.a.mmu.Peek(RAMFONTS + (idx*3)+2)) << 16 |
						uint32(c.a.mmu.Peek(RAMFONTS + (idx*3)+1)) << 8 |
						uint32(c.a.mmu.Peek(RAMFONTS + (idx*3)))
	}

	/*if c.trace {
		fmt.Printf("[CardBitBlt] LOAD_FONT(%x) @$%x\n", fontID, fontAddr)
	}*/

	// 2-byte header with the number of glyphs (up to 256) and the height of the glyphs
	c.font.nglyphs = c.a.mmu.Peek(fontAddr)
	c.font.height = c.a.mmu.Peek(fontAddr+1)
	c.font.baseline = c.a.mmu.Peek(fontAddr+2)
	c.font.address = fontAddr

	// 3-byte array of widths and offsets to the glyph pixels
	for i := uint32(0); i < uint32(c.font.nglyphs); i++ {
		c.font.glyphs[i].width = c.a.mmu.Peek(fontAddr+3+(i*3))
		c.font.glyphs[i].offset = uint16(c.a.mmu.Peek(fontAddr+3+(i*3)+2)) << 8 |
									uint16(c.a.mmu.Peek(fontAddr+3+(i*3)+1))

		/* if c.trace{
			fmt.Printf("[CardBitBlt] font[%d] @%x: %d %x\n",
				i, fontAddr+3+(i*3), c.font.glyphs[i].width, c.font.glyphs[i].offset)
		} */
	}
}


//
//  Copy $C800 to a string
//
func (c *CardBitBlt) c800toString() string {
	name := make([]uint8, 32)
	for i := 0; i < 32; i++ {
		name[i] = c.c800[i]
		if c.c800[i] == 0x00 {
			name = name[:i]
			break
		}
		// @@@ TODO - Add wait to simulate clock cycles
	}

	return string(name)
}


//
//  Read and write this card's $C800 RAM
//
func (c *CardBitBlt) peek(address uint32) uint8 {
//fmt.Printf("CardBitBlt PEEK(%x) = %x\n", address, c.c800[address-0xC800])
	return c.c800[address-0xC800]
}
func (c *CardBitBlt) poke(address uint32, value uint8) {
//fmt.Printf("CardBitBlt POKE(%x) = %x\n", address, value)
	c.c800[address-0xC800] = value
}
func (c *CardBitBlt) setBase(address uint32) {
}


//
//  This card's ROM ($C700-$C7FF)
//
func buildBITBLTRom(slot int) []uint8 {
	data := make([]uint8, 256)

	copy(data, []uint8{
		// Preamble bytes to comply with the expectation in $Cn01, 3, 5 and 7
		0xa9, 0x20, // LDA #$20
		0xa9, 0x00, // LDA #$00
		0xa9, 0x03, // LDA #$03
		0xa9, 0x00, // LDA #$00

		0x60, // RTS
	})

	data[0xfc] = 0
	data[0xfd] = 0
	data[0xfe] = 3    // Status and Read. No write, no format. Single volume
	data[0xff] = 0x08 // Driver entry point

	return data
}
