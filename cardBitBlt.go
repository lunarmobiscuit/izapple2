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
	BITBLT_CHAR
	BITBLT_STRING
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

const ROMFONTS				= 0xFFE000
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
		case BITBLT_CHAR: c.BitBltChar()
		case BITBLT_STRING: c.BitBltString()
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
c.param1 = 0
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
c.param1 = 0
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
//  Load the font table into a struct
//
func (c *CardBitBlt) bltChar(ch uint8, x uint16, y uint16) (uint16, uint8) {
	// Low ASCII
	fmt.Printf("CHAR %x\n", ch)
	if (ch >= 128) {
		ch -= 128
	}

	switch c.font.height {
		case 8: return c.blt8BitChar(ch, x, y)
		case 16: return c.blt16BitChar(ch, x, y)
		case 24: return c.blt24BitChar(ch, x, y)
	}

	width := uint32(c.font.glyphs[ch].width)
	x = 0
	for ch = 0; ch < 8; ch++ {
		width = uint32(c.font.glyphs[ch].width)
		offset := uint32(c.font.glyphs[ch].offset)
		glyphsPtr := c.font.address + uint32(offset)

		for row := uint32(0); row < uint32(c.font.height); row++ {
			rAddr := uint32(GRAPHICS_BASE + ((uint32(y) + row) * GRAPHICS_ROW_BYTES))
			fromAddr := glyphsPtr + (row * ((width+7)/8))
			toAddr := rAddr + uint32(x/8);

			c.a.mmu.Poke(toAddr, uint8(c.a.mmu.Peek(fromAddr)))
			c.a.mmu.Poke(toAddr+1, uint8(c.a.mmu.Peek(fromAddr+1)))
			c.a.mmu.Poke(toAddr+1, uint8(c.a.mmu.Peek(fromAddr+2)))
			c.a.mmu.Poke(toAddr+1, uint8(c.a.mmu.Peek(fromAddr+3)))

			fromAddr += 4
			toAddr += 4
		}

		x += (uint16(width & 0xff) + 7) / 8
	}

/*
	// Lookup the details of this glyph
	width := uint32(c.font.glyphs[ch].width)
	offset := uint32(c.font.glyphs[ch].offset)
	glyphsPtr := c.font.address + uint32(offset)

	if c.trace {
		fmt.Printf("[CardBitBlt] ch='%c' (%02x)  FONT=%x  X=%d  Y=%d  width=%d  height=%d  offset=%x\n",
			ch + '@', ch, c.fontID, x, y, width, c.font.height, offset)
	}

	for row := uint32(0); row < uint32(c.font.height); row++ {
		rAddr := uint32(GRAPHICS_BASE + ((uint32(y) + row) * GRAPHICS_ROW_BYTES))
		fromAddr := glyphsPtr + (row * (width+7)/8)
		toAddr := rAddr + uint32(x/8);

		for col := uint32(0); col <= width/8; col++ {
			if c.trace && row < uint32(y+2) {
				fmt.Printf("  %d 0x%x @%x -> @%x\n", row, uint32(ch), fromAddr, toAddr)
			}

			glyphBits := uint16(c.a.mmu.Peek(fromAddr)) | uint16(c.a.mmu.Peek(fromAddr + 1)) << 8
			c.a.mmu.Poke(toAddr, uint8(glyphBits & 0xff))
			fromAddr += 1
			toAddr += 1
		}
	}
*/

/*
		// Draw the first pixels that are not aligned to x-coord mod 8
		p := 0
		ip := int(8 - (x & 0x7))
		if ip != 8 {
			fmt.Printf("  %d pixels unaligned pixels to start\n", ip)
			glyphBits := c.a.mmu.Peek(glyphsPtr + (r*uint32(c.font.height)))
			bits := c.a.mmu.Peek(rAddr + (uint32(x) + uint32(p))/8)
			bits |= glyphBits << ip
			c.a.mmu.Poke(rAddr + (uint32(x) + uint32(p))/8, bits)
			p += ip
		}

		// Draw the aligned pixels
		fmt.Printf("  %d pixels in the middle\n", width & 0xf8)
		for i := p; i <= int(width & 0xf8); i += 8 {
			glyphBits := c.a.mmu.Peek(glyphsPtr + (r*uint32(c.font.height))+(uint32(p/8)))
			bits := c.a.mmu.Peek(rAddr + (uint32(x) + uint32(p))/8)
			bits |= glyphBits
			c.a.mmu.Poke(rAddr + (uint32(x) + uint32(p))/8, bits)
			p += 8
		}

		// Draw the final pixels that are not aligned to x-coord mod 8
		if p <= int(width) {
			fmt.Printf("  %d pixels unaligned pixels to end\n", int(width) - p)
			glyphBits := c.a.mmu.Peek(glyphsPtr + (r*uint32(c.font.height))+(uint32(p/8)))
			bits := c.a.mmu.Peek(rAddr + (uint32(x) + uint32(p))/8)
			bits |= glyphBits >> (int(width) - p)
			c.a.mmu.Poke(rAddr + (uint32(x) + uint32(p))/8, bits)
		}
	}
*/

	return x + uint16(width), c.font.glyphs[ch].width
}


// @@@ Correctly draws aligned 8-pixel wide font
func (c *CardBitBlt) blt8BitChar(ch uint8, x uint16, y uint16) (uint16, uint8) {
	// Lookup the details of this glyph
	width := c.font.glyphs[ch].width
	offset := uint32(c.font.glyphs[ch].offset)
	glyphsPtr := c.font.address + uint32(offset)

	if c.trace {
		fmt.Printf("[CardBitBlt] ch='%c' (%02x)  FONT=%x  X=%d  Y=%d  width=%d  height=%d  offset=%x\n",
			ch + '@', ch, c.fontID, x, y, width, c.font.height, offset)
	}

	for r := uint32(0); r < uint32(c.font.height); r++ {
		rAddr := uint32(GRAPHICS_BASE + ((uint32(y) + r) * GRAPHICS_ROW_BYTES))
		toAddr := rAddr + uint32(x/8);

		fromAddr := glyphsPtr + r

		if c.trace && r < uint32(y+2) {
			fmt.Printf("  %d 0x%x @%x -> @%x\n", r, uint32(ch), fromAddr, toAddr)
		}

		glyphBits := c.a.mmu.Peek(fromAddr)
		c.a.mmu.Poke(toAddr, glyphBits)

		toAddr += 1
	}

	return x + uint16(width), width
}

// @@@ Correctly draws aligned 16-pixel wide font
func (c *CardBitBlt) blt16BitChar(ch uint8, x uint16, y uint16) (uint16, uint8) {
	width := uint32(c.font.glyphs[0].width)
	x = 0
	for ch = 0; ch < 32; ch++ {
		width = uint32(c.font.glyphs[ch].width)
		offset := uint32(c.font.glyphs[ch].offset)
		glyphsPtr := c.font.address + uint32(offset)

		for row := uint32(0); row < uint32(c.font.height); row++ {
			rAddr := uint32(GRAPHICS_BASE + ((uint32(y) + row) * GRAPHICS_ROW_BYTES))
			fromAddr := glyphsPtr + (row * ((width+7)/8))
			toAddr := rAddr + uint32(x/8);

			glyphBits := uint16(c.a.mmu.Peek(fromAddr)) | uint16(c.a.mmu.Peek(fromAddr + 1)) << 8
			c.a.mmu.Poke(toAddr, uint8(glyphBits & 0xff))
			c.a.mmu.Poke(toAddr+1, uint8(glyphBits >> 8 & 0xff))

			c.a.mmu.Poke(toAddr, uint8(c.a.mmu.Peek(fromAddr)))
			c.a.mmu.Poke(toAddr+1, uint8(c.a.mmu.Peek(fromAddr+1)))

			fromAddr += 2
			toAddr += 2
		}

		x += uint16(width & 0xff)
	}

	return x + uint16(width), uint8(width)
}

// @@@ Draws aligned 24-pixel wide font?
func (c *CardBitBlt) blt24BitChar(ch uint8, x uint16, y uint16) (uint16, uint8) {
	width := uint32(c.font.glyphs[0].width)
	x = 0
	for ch = 0; ch < 16; ch++ {
		width = uint32(c.font.glyphs[ch].width)
		offset := uint32(c.font.glyphs[ch].offset)
		glyphsPtr := c.font.address + uint32(offset)

		for row := uint32(0); row < uint32(c.font.height); row++ {
			rAddr := uint32(GRAPHICS_BASE + ((uint32(y) + row) * GRAPHICS_ROW_BYTES))
			fromAddr := glyphsPtr + (row * ((width+7)/8))
			toAddr := rAddr + uint32(x/8);

			glyphBits := uint32(c.a.mmu.Peek(fromAddr)) |
							uint32(c.a.mmu.Peek(fromAddr + 1)) << 8 |
							uint32(c.a.mmu.Peek(fromAddr + 2)) << 16
			c.a.mmu.Poke(toAddr, uint8(glyphBits & 0xff))
			c.a.mmu.Poke(toAddr+1, uint8(glyphBits >> 8 & 0xff))
			c.a.mmu.Poke(toAddr+2, uint8(glyphBits >> 16 & 0xff))

			c.a.mmu.Poke(toAddr, uint8(c.a.mmu.Peek(fromAddr)))
			c.a.mmu.Poke(toAddr+1, uint8(c.a.mmu.Peek(fromAddr+1)))
			c.a.mmu.Poke(toAddr+2, uint8(c.a.mmu.Peek(fromAddr+2)))

			fromAddr += 3
			toAddr += 3
		}

		x += uint16(width & 0xff)
	}

	x = 0
	for ch = 16; ch < 16+16; ch++ {
		width = uint32(c.font.glyphs[ch].width)
		offset := uint32(c.font.glyphs[ch].offset)
		glyphsPtr := c.font.address + uint32(offset)

		for row := uint32(0); row < uint32(c.font.height); row++ {
			rAddr := uint32(GRAPHICS_BASE + ((uint32(y) + (row + 30)) * GRAPHICS_ROW_BYTES))
			fromAddr := glyphsPtr + (row * ((width+7)/8))
			toAddr := rAddr + uint32(x/8);

			glyphBits := uint32(c.a.mmu.Peek(fromAddr)) |
							uint32(c.a.mmu.Peek(fromAddr + 1)) << 8 |
							uint32(c.a.mmu.Peek(fromAddr + 2)) << 16
			c.a.mmu.Poke(toAddr, uint8(glyphBits & 0xff))
			c.a.mmu.Poke(toAddr+1, uint8(glyphBits >> 8 & 0xff))
			c.a.mmu.Poke(toAddr+2, uint8(glyphBits >> 16 & 0xff))

			c.a.mmu.Poke(toAddr, uint8(c.a.mmu.Peek(fromAddr)))
			c.a.mmu.Poke(toAddr+1, uint8(c.a.mmu.Peek(fromAddr+1)))
			c.a.mmu.Poke(toAddr+2, uint8(c.a.mmu.Peek(fromAddr+2)))

			fromAddr += 3
			toAddr += 3
		}

		x += uint16(width & 0xff)
	}

	x = 0
	for ch = 90; ch < 90+16; ch++ {
		width = uint32(c.font.glyphs[ch].width)
		offset := uint32(c.font.glyphs[ch].offset)
		glyphsPtr := c.font.address + uint32(offset)

		for row := uint32(0); row < uint32(c.font.height); row++ {
			rAddr := uint32(GRAPHICS_BASE + ((uint32(y) + (row + 60)) * GRAPHICS_ROW_BYTES))
			fromAddr := glyphsPtr + (row * ((width+7)/8))
			toAddr := rAddr + uint32(x/8);

			glyphBits := uint32(c.a.mmu.Peek(fromAddr)) |
							uint32(c.a.mmu.Peek(fromAddr + 1)) << 8 |
							uint32(c.a.mmu.Peek(fromAddr + 2)) << 16
			c.a.mmu.Poke(toAddr, uint8(glyphBits & 0xff))
			c.a.mmu.Poke(toAddr+1, uint8(glyphBits >> 8 & 0xff))
			c.a.mmu.Poke(toAddr+2, uint8(glyphBits >> 16 & 0xff))

			c.a.mmu.Poke(toAddr, uint8(c.a.mmu.Peek(fromAddr)))
			c.a.mmu.Poke(toAddr+1, uint8(c.a.mmu.Peek(fromAddr+1)))
			c.a.mmu.Poke(toAddr+2, uint8(c.a.mmu.Peek(fromAddr+2)))

			fromAddr += 3
			toAddr += 3
		}

		x += uint16(width & 0xff)
	}

	return x + uint16(width), uint8(width)
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

	if c.trace {
		fmt.Printf("[CardBitBlt] LOAD_FONT(%x) @$%x\n", fontID, fontAddr)
	}

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

if c.font.height == 8 {
	for i := uint32(0); i < 8; i++ {
		addr := uint32(ROMFONTS) + uint32(c.font.glyphs[1].offset) + i
		fmt.Printf("A: $%06x %08b\n", addr, c.a.mmu.Peek(addr))
	}
}

if c.font.height == 16 {
	for i := uint32(0); i < 16; i++ {
		addr := uint32(ROMFONTS) + uint32(c.font.glyphs[1].offset) + (i*2)
		fmt.Printf("A: $%06x %08b%08b\n", addr, c.a.mmu.Peek(addr), c.a.mmu.Peek(addr+1))
	}
}

if c.font.height == 24 {
	for i := uint32(0); i < 24; i++ {
		addr := uint32(RAMFONTS) + uint32(c.font.glyphs[1].offset) + (i*3)
		fmt.Printf("A: $%06x %08b %08b %08b\n", addr, c.a.mmu.Peek(addr), c.a.mmu.Peek(addr+1), c.a.mmu.Peek(addr+2))
	}
}

if c.font.height == 32 {
	for i := uint32(0); i < 32; i++ {
		addr := uint32(RAMFONTS) + uint32(c.font.glyphs[1].offset) + (i*4)
		fmt.Printf("A: $%06x %08b %08b %08b %08b\n", addr, c.a.mmu.Peek(addr), c.a.mmu.Peek(addr+1), c.a.mmu.Peek(addr+2), c.a.mmu.Peek(addr+3))
	}
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
