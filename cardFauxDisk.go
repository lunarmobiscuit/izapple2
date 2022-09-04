package izapple2

import (
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"os"

	//"github.com/lunarmobiscuit/izapple2/storage"
)

const (
	FAUX_RESET = iota
	FAUX_VOLUME_NAME
	FAUX_CATALOG
	FAUX_CATALOG_NEXT
	FAUX_EXISTS
	FAUX_CREATE
	FAUX_OPEN
	FAUX_READ
	FAUX_READ_NEXT
	FAUX_READ_DMA
	FAUX_WRITE
	FAUX_WRITE_NEXT
	FAUX_WRITE_DMA
	FAUX_CLOSE
	FAUX_CHDIR
	FAUX_CHDIR_UP
)

const (
	FAUX_SUCCESS = iota
	FAUX_END_OF_CATALOG
	FAUX_END_OF_FILE
	FAUX_ERR_NOT_FOUND
	FAUX_ERR_READ_ERROR
)

/*
 ***  Neither a hard disk nor a floppy, but a faux storage device that uses the emulator computer's hard disk
 */

// CardFauxDisk represents a faux storage disk but is just the emulator's hard disk
type CardFauxDisk struct {
	cardBase

	trace		bool

	rootName	string
	root		[]fs.DirEntry
	dirIdx		int

	cmd			uint8				// Which command to run
	arg0		uint32
	arg1		uint32
	ret0		uint32
	ret1		uint32
	ret2		uint16
	retErr		uint8

	c800		[0x800]uint8
}

// NewCardFauxDisk creates a new FauxDisk card
func NewCardFauxDisk() *CardFauxDisk {
	var c CardFauxDisk
	c.name = "Disk][4"
	c.romC8xx = &c
	return &c
}

// GetInfo returns disk info
func (c *CardFauxDisk) GetInfo() map[string]string {
	info := make(map[string]string)
	info["dirname"] = c.rootName
	info["trace"] = strconv.FormatBool(c.trace)
	return info
}

// LoadRoot loads the root directory
func (c *CardFauxDisk) LoadRoot(rootDirName string) error {
	c.rootName = rootDirName

	if c.trace {
		fmt.Printf("[CardFauxDisk] Faux root directory: '%s'\n", c.rootName)
	}

	// Load the root directory
	var err error
	c.root, err = os.ReadDir(c.rootName)
	if err != nil {
		return err
	}

	return nil
}

//
//  Set up the softswitches
//
func (c *CardFauxDisk) assign(a *Apple2, slot int) {
	c.loadRom(buildFauxDiskRom(slot))

	c.addCardSoftSwitchW(0, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Command $%d\n", value)
		}
		c.cmd = value
		return
	}, "FAUXDISKCMD")

	c.addCardSoftSwitchW(1, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Arg0.l $%x\n", value)
		}
		c.arg0 = uint32(value)
		return
	}, "FAUXDISKARG0L")
	c.addCardSoftSwitchW(2, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Arg0.m $%x\n", value)
		}
		c.arg0 |= uint32(value) << 8 
		return
	}, "FAUXDISKARG0M")
	c.addCardSoftSwitchW(3, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Arg0.h $%x\n", value)
		}
		c.arg0 |= uint32(value) << 16 
		return
	}, "FAUXDISKARG0H")

	c.addCardSoftSwitchW(4, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Arg1.l $%x\n", value)
		}
		c.arg1 = uint32(value)
		return
	}, "FAUXDISKARG0L")
	c.addCardSoftSwitchW(5, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Arg1.m $%x\n", value)
		}
		c.arg1 |= uint32(value) << 8 
		return
	}, "FAUXDISKARG0M")
	c.addCardSoftSwitchW(6, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Arg1.h $%x\n", value)
		}
		c.arg1 |= uint32(value) << 16 
		return
	}, "FAUXDISKARG0H")

	c.addCardSoftSwitchR(7, func() uint8 {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Ret0.l $%x\n", c.ret0 & 0xff)
		}
		return uint8((c.ret0 >> 8) & 0xff)
	}, "FAUXDISKRET0L")
	c.addCardSoftSwitchR(8, func() uint8 {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Ret0.m $%x\n", (c.ret0 >> 8) & 0xff)
		}
		return uint8(c.ret0 & 0xff)
	}, "FAUXDISKRET0M")
	c.addCardSoftSwitchR(9, func() uint8 {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Ret0.h $%x\n", (c.ret0 >> 16) & 0xff)
		}
		return uint8((c.ret0 >> 16) & 0xff)
	}, "FAUXDISKRET0H")

	c.addCardSoftSwitchR(10, func() uint8 {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Ret1.l $%x\n", c.ret1 & 0xff)
		}
		return uint8((c.ret1 >> 8) & 0xff)
	}, "FAUXDISKRET1L")
	c.addCardSoftSwitchR(11, func() uint8 {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Ret1.m $%x\n", (c.ret1 >> 8) & 0xff)
		}
		return uint8(c.ret1 & 0xff)
	}, "FAUXDISKRET1M")
	c.addCardSoftSwitchR(12, func() uint8 {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Ret1.h $%x\n", (c.ret1 >> 16) & 0xff)
		}
		return uint8((c.ret1 >> 16) & 0xff)
	}, "FAUXDISKRET1H")

	c.addCardSoftSwitchR(13, func() uint8 {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Ret2.l $%x\n", c.ret2 & 0xff)
		}
		return uint8(c.ret2 & 0xff)
	}, "FAUXDISKRET1L")
	c.addCardSoftSwitchR(14, func() uint8 {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Ret2.h $%x\n", (c.ret2 >> 8) & 0xff)
		}
		return uint8((c.ret2 >> 8) & 0xff)
	}, "FAUXDISKRET1M")

	c.addCardSoftSwitchRW(15, func() uint8 {
		if c.trace {
			fmt.Printf("[CardFauxDisk] DO COMMAND $%d\n", c.cmd)
		}
		switch (c.cmd) {
		case FAUX_RESET: return c.fauxDiskReset()
		case FAUX_VOLUME_NAME: return c.fauxDiskName()
		case FAUX_CATALOG: return c.fauxDiskCatalog(true)
		case FAUX_CATALOG_NEXT: return c.fauxDiskCatalog(false)
		case FAUX_OPEN: return c.fauxDiskOpen()
		case FAUX_READ: return c.fauxDiskRead(true)
		case FAUX_READ_NEXT: return c.fauxDiskRead(false)
		case FAUX_CLOSE: return c.fauxDiskClose()
		case FAUX_CHDIR: return c.fauxDiskChdir()
		case FAUX_CHDIR_UP: return c.fauxDiskChdirUp()
		}
		return c.retErr
	}, "FAUXDISKCMD")

	c.cardBase.assign(a, slot)
}

//
//  Reset the Faux disk
//
func (c *CardFauxDisk) fauxDiskReset() uint8 {
	if c.trace {
		fmt.Printf("[CardFauxDisk] FAUX_RESET\n")
	}

	c.LoadRoot(c.rootName)
	c.dirIdx = 0

	c.retErr = FAUX_SUCCESS
	return FAUX_SUCCESS
}

//
//  Return the volume name
//    0xC800 = zero-terminated name (high ASCII)
//
func (c *CardFauxDisk) fauxDiskName() uint8 {
	if c.trace {
		fmt.Printf("[CardFauxDisk] FAUX_VOLUME_NAME\n")
	}

	// Copy the name to peripheral RAM
	addr := uint32(0xc800)
	for i := 0; i < len(c.name) && i < 16; i++ {
		c.c800[addr-0xc800] = uint8(c.name[i]) | 0x80
		addr += 1
	}
	c.c800[addr-0xc800] = 0x00

	c.retErr = FAUX_SUCCESS
	return FAUX_SUCCESS
}

//
//  Return an entry in the current directory
//    0xC800-0xC802 = type (3-char high ASCII)
//    0xC803-0xC805 = length
//    0xC806-0xC8nn = zero-terminated name (high ASCII)
//
func (c *CardFauxDisk) fauxDiskCatalog(firstCall bool) uint8 {
	if c.trace {
		fmt.Printf("[CardFauxDisk] FAUX_CATALOG %t\n", firstCall)
	}

	// Begint the catalog process
	if (firstCall) {
		// Reset the directory index
		c.dirIdx = 0

		// Return the number of items in directory
		c.ret0 = uint32(len(c.root) & 0x0FFFFFF)
	} else if (c.dirIdx >= len(c.root)) {
		// No more items
		return FAUX_END_OF_CATALOG
	} else {
		e := c.root[c.dirIdx]
		name := e.Name()

		// Skip the . files
		for (name[0] == '.') {
			if (c.dirIdx >= len(c.root)) {
				return FAUX_END_OF_CATALOG
			}
			c.dirIdx += 1
			e = c.root[c.dirIdx]
			name = e.Name()
		}
		ftype := "   "
		finfo, _ := e.Info()
		size := finfo.Size()  // size in Kbytes
		isdir := e.IsDir()

		// Extract and return a 3-byte type (from the filename .suffix)
		if (isdir) {
			ftype = ":::"
			size = 0
		} else {
			dot := strings.LastIndexByte(name, '.')
			if (dot != -1) {
				ftype = strings.ToUpper(name[dot+1:] + "   ")
				name = name[:dot]
			}
		}
		c.c800[0] = uint8(ftype[0]) | 0x80
		c.c800[1] = uint8(ftype[1]) | 0x80
		c.c800[2] = uint8(ftype[2]) | 0x80

		// Return the dize (little endian)
		if (size > 0) && (size < 1024) {
			size = 1
		} else {
			size /= 1024
		}
		if (isdir) {
			c.c800[3] = '-' | 0x80
			c.c800[4] = '-' | 0x80
			c.c800[5] = '-' | 0x80
		} else {
			c.c800[3] = uint8(size % 1000 / 100) + ('0' | 0x80)
			c.c800[4] = uint8(size % 100 / 10) + ('0' | 0x80)
			c.c800[5] = uint8(size % 10) + ('0' | 0x80)
		}

		// Copy the name to peripheral RAM
		addr := uint32(0xc806)
		for i := 0; i < len(name) && i < 16; i++ {
			c.c800[addr-0xc800] = uint8(name[i]) | 0x80
			addr += 1
		}
		c.c800[addr-0xc800] = 0x00

		c.dirIdx += 1
	}

	c.retErr = FAUX_SUCCESS
	return FAUX_SUCCESS
}

func (c *CardFauxDisk) fauxDiskOpen() uint8 {
	if c.trace {
		fmt.Printf("[CardFauxDisk] FAUX_OPEN\n")
	}

	return FAUX_SUCCESS
}

func (c *CardFauxDisk) fauxDiskRead(firstCall bool) uint8 {
	if c.trace {
		fmt.Printf("[CardFauxDisk] FAUX_READ\n")
	}

	return FAUX_SUCCESS
}

func (c *CardFauxDisk) fauxDiskClose() uint8 {
	if c.trace {
		fmt.Printf("[CardFauxDisk] FAUX_CLOSE\n")
	}

	return FAUX_SUCCESS
}

func (c *CardFauxDisk) fauxDiskChdir() uint8 {
	if c.trace {
		fmt.Printf("[CardFauxDisk] FAUX_CHDIR\n")
	}

	return FAUX_SUCCESS
}

func (c *CardFauxDisk) fauxDiskChdirUp() uint8 {
	if c.trace {
		fmt.Printf("[CardFauxDisk] FAUX_CHDIR_UP\n")
	}

	return FAUX_SUCCESS
}


func (c *CardFauxDisk) peek(address uint32) uint8 {
//fmt.Printf("CardFauxDisk PEEK(%x) = %x\n", address, c.c800[address-0xC800])
	return c.c800[address-0xC800]
}
func (c *CardFauxDisk) poke(address uint32, value uint8) {
//fmt.Printf("CardFauxDisk POKE(%x) = %x\n", address, value)
	c.c800[address-0xC800] = value
}
func (c *CardFauxDisk) setBase(address uint32) {
}


func buildFauxDiskRom(slot int) []uint8 {
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
