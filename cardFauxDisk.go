package izapple2

import (
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"os"
	"time"

	//"github.com/lunarmobiscuit/izapple2/storage"
)

const (
	FAUX_VOLUME_NAME = iota
	FAUX_CATALOG
	FAUX_CATALOG_NEXT
	FAUX_EXISTS
	FAUX_CREATE
	FAUX_OPEN
	FAUX_READ
	FAUX_READ_DMA
	FAUX_WRITE
	FAUX_WRITE_DMA
	FAUX_CLOSE
	FAUX_SEEK
)

const (
	FAUX_SUCCESS = iota
	FAUX_END_OF_CATALOG
	FAUX_END_OF_FILE
	FAUX_ERR_EXISTS
	FAUX_ERR_NOT_FOUND
	FAUX_ERR_READ_ERROR
	FAUX_ERR_WRITE_ERROR
)

/*
 ***  Neither a hard disk nor a floppy, but a faux storage device that uses the emulator computer's hard disk
 */

// CardFauxDisk represents a faux storage disk but is just the emulator's hard disk
type CardFauxDisk struct {
	cardBase

	trace		bool

	rootName	string
	root		[]fauxFile
	dirIdx		int

	files		[8]*os.File

	cmd			uint8				// Which command to run
	param0		uint32
	param1		uint32
	ret0		uint32
	ret1		uint32
	ret2		uint16
	retErr		uint8

	c800		[0x800]uint8

	catDir		[]fauxFile
}

// CardFauxDisk represents a faux storage disk but is just the emulator's hard disk
type fauxFile struct {
	filename	string
	name		string
	ftype		string
	size		int64
	isdir		bool
	subdir		[]fauxFile
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

	return c.reloadRoot()
}
func (c *CardFauxDisk) reloadRoot() error {
	// Load the root directory
	dir, err := os.ReadDir(c.rootName)
	if err != nil {
		return err
	}
	root, err := c.processDirectory(dir, 0)
	if (err != nil) {
		return err
	}
	c.root = root

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
			fmt.Printf("[CardFauxDisk] Param0.l $%02x\n", value)
		}
		c.param0 = uint32(value)
		return
	}, "FAUXDISKPARAM1L")
	c.addCardSoftSwitchW(2, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Param0.m $%02x\n", value)
		}
		c.param0 |= uint32(value) << 8 
		return
	}, "FAUXDISKPARAM1M")
	c.addCardSoftSwitchW(3, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Param0.h $%02x\n", value)
		}
		c.param0 |= uint32(value) << 16 
		return
	}, "FAUXDISKPARAM1H")

	c.addCardSoftSwitchW(4, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Param1.l $%02x\n", value)
		}
		c.param1 = uint32(value)
		return
	}, "FAUXDISKPARAM1L")
	c.addCardSoftSwitchW(5, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Param1.m $%02x\n", value)
		}
		c.param1 |= uint32(value) << 8 
		return
	}, "FAUXDISKPARAM1M")
	c.addCardSoftSwitchW(6, func(value uint8) {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Param1.h $%02x\n", value)
		}
		c.param1 |= uint32(value) << 16 
		return
	}, "FAUXDISKPARAM1H")

	c.addCardSoftSwitchR(7, func() uint8 {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Ret0.l $%02x\n", c.ret0 & 0xff)
		}
		return uint8(c.ret0 & 0xff)
	}, "FAUXDISKRET0L")
	c.addCardSoftSwitchR(8, func() uint8 {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Ret0.m $%02x\n", (c.ret0 >> 8) & 0xff)
		}
		return uint8((c.ret0 >> 8) & 0xff)
	}, "FAUXDISKRET0M")
	c.addCardSoftSwitchR(9, func() uint8 {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Ret0.h $%02x\n", (c.ret0 >> 16) & 0xff)
		}
		return uint8((c.ret0 >> 16) & 0xff)
	}, "FAUXDISKRET0H")

	c.addCardSoftSwitchR(10, func() uint8 {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Ret1.l $%02x\n", c.ret1 & 0xff)
		}
		return uint8((c.ret1 >> 8) & 0xff)
	}, "FAUXDISKRET1L")
	c.addCardSoftSwitchR(11, func() uint8 {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Ret1.m $%02x\n", (c.ret1 >> 8) & 0xff)
		}
		return uint8(c.ret1 & 0xff)
	}, "FAUXDISKRET1M")
	c.addCardSoftSwitchR(12, func() uint8 {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Ret1.h $%02x\n", (c.ret1 >> 16) & 0xff)
		}
		return uint8((c.ret1 >> 16) & 0xff)
	}, "FAUXDISKRET1H")

	c.addCardSoftSwitchR(13, func() uint8 {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Ret2.l $%02x\n", c.ret2 & 0xff)
		}
		return uint8(c.ret2 & 0xff)
	}, "FAUXDISKRET1L")
	c.addCardSoftSwitchR(14, func() uint8 {
		if c.trace {
			fmt.Printf("[CardFauxDisk] Ret2.h $%02x\n", (c.ret2 >> 8) & 0xff)
		}
		return uint8((c.ret2 >> 8) & 0xff)
	}, "FAUXDISKRET1M")

	c.addCardSoftSwitchRW(15, func() uint8 {
		if c.trace {
			fmt.Printf("[CardFauxDisk] DO COMMAND $%d\n", c.cmd)
		}
		switch (c.cmd) {
		case FAUX_VOLUME_NAME: return c.fauxDiskName()
		case FAUX_CATALOG: return c.fauxDiskCatalog(true)
		case FAUX_CATALOG_NEXT: return c.fauxDiskCatalog(false)
		case FAUX_EXISTS: return c.fauxDiskExists()
		case FAUX_CREATE: return c.fauxDiskCreate()
		case FAUX_OPEN: return c.fauxDiskOpen()
		case FAUX_READ: return c.fauxDiskRead()
		case FAUX_READ_DMA: return c.fauxDiskReadDMA()
		case FAUX_WRITE: return c.fauxDiskWrite()
		case FAUX_WRITE_DMA: return c.fauxDiskWriteDMA()
		case FAUX_CLOSE: return c.fauxDiskClose()
		}
		return c.retErr
	}, "FAUXDISKCMD")

	c.cardBase.assign(a, slot)
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
	// Optional subdir name
	catSubdir := false
	dirname := strings.ToUpper(c.c800toName())
	if dirname != "" {
		catSubdir = true
		if c.trace {
			fmt.Printf("[CardFauxDisk] FAUX_CATALOG %s %t\n", dirname, firstCall)
		}
	} else {
		if c.trace {
			fmt.Printf("[CardFauxDisk] FAUX_CATALOG %t\n", firstCall)
		}
	}
	
	// Begin the catalog process
	if (firstCall) {
		// Load the root directory
		c.LoadRoot(c.rootName)

		// Reset the directory index
		c.dirIdx = 0

		if catSubdir {
			subdir := c.findSubdir(dirname)
			if (subdir == nil) {
				fmt.Printf("[CardFauxDisk] FILE NOT FOUND\n")
				return FAUX_ERR_NOT_FOUND
			}
			c.catDir = subdir

			// Return the number of items in subdr
			c.ret0 = uint32(len(subdir) & 0x0FFFFFF)
		} else {
			// Return the number of items in directory
			c.catDir = c.root
			c.ret0 = uint32(len(c.root) & 0x0FFFFFF)
		}
	} else if (c.dirIdx >= len(c.catDir)) {
		// No more items
		return FAUX_END_OF_CATALOG
	} else {
		f := c.catDir[c.dirIdx]

		c.c800[0] = uint8(f.ftype[0]) | 0x80
		c.c800[1] = uint8(f.ftype[1]) | 0x80
		c.c800[2] = uint8(f.ftype[2]) | 0x80

		if (f.isdir) {
			c.c800[3] = '-' | 0x80
			c.c800[4] = '-' | 0x80
			c.c800[5] = '-' | 0x80
		} else {
			c.c800[3] = uint8(f.size % 1000 / 100) + ('0' | 0x80)
			c.c800[4] = uint8(f.size % 100 / 10) + ('0' | 0x80)
			c.c800[5] = uint8(f.size % 10) + ('0' | 0x80)
		}

		addr := uint32(0xc806)
		for i := 0; i < len(f.name) && i < 16; i++ {
			c.c800[addr-0xc800] = uint8(f.name[i]) | 0x80
			addr += 1
		}
		c.c800[addr-0xc800] = 0x00

		c.dirIdx += 1
	}

	c.retErr = FAUX_SUCCESS
	return FAUX_SUCCESS
}

//
//  Process the items from the OS directory
//
func (c *CardFauxDisk) processDirectory(dir []fs.DirEntry, level int) ([]fauxFile, error) {
	nFiles := len(dir)
	files := make([]fauxFile, nFiles)
	n := 0
	for i := 0; i < nFiles; i++ {
		e := dir[i]

		// Skip the files that start with '.'
		if (e.Name()[0] == '.') {
			files = files[:len(files)-1] // shrink the array
			continue
		}

		// Copy the basic information
		finfo, _ := e.Info()
		f := files[n]
		f.filename = e.Name()
		f.ftype = "   "
		f.isdir = finfo.IsDir()
		f.size = finfo.Size()

		// Extract and return a 3-byte type (from the filename .suffix)
		if (f.isdir) {
			// Only allow one level of subdirectories
			if (level > 0) {
				continue
			}

			f.ftype = ":::"
			f.name = f.filename
			f.size = 0

			// Load the sub directory
			subdir, err := os.ReadDir(c.rootName + "/" + f.filename)
			if err != nil {
				return files, err
			}

			f.subdir , err = c.processDirectory(subdir, level+1)
			if err != nil {
				return files, err
			}
		} else {
			dot := strings.LastIndexByte(f.filename, '.')
			if (dot != -1) {
				f.ftype = (strings.ToUpper(f.filename[dot+1:] + "   "))[0:3]
				f.name = f.filename[:dot]
			}
		}

		// Set the size to be in K (rounded up)
		if (f.size > 0) && (f.size < 1024) {
			f.size = 1
		} else {
			f.size = (f.size + 512) / 1024
		}

		/*if c.trace {
			fmt.Printf("[CardFauxDisk] %d %s %d %s %s\n", n, f.ftype, f.size, f.name, f.filename)
		}*/

		files[n] = f
		n += 1
	}

	return files, nil
}


//
//  Find a subdir in root
//
func (c *CardFauxDisk) findSubdir(dirname string) []fauxFile {
	// Find the matching subdir
	for i := 0; i < len(c.root); i++ {
		f := c.root[i]
		if (dirname == strings.ToUpper(f.name)) {
			return f.subdir
		}
	}

	return nil
}


//
//  Does the file exist?
//    0xC800 = zero-terminated name (high ASCII)
//	Returns
//	  0xC800-0xC802 = file type
//	  0xC803-0xC805 = file size (in ASCII)
//	  0xC806-0xC808 = file size (as 24-bit value)
//
func (c *CardFauxDisk) fauxDiskExists() uint8 {
	if c.trace {
		fmt.Printf("[CardFauxDisk] FAUX_EXISTS\n")
	}

	fname := strings.ToUpper(c.c800toName())
	if c.trace {
		fmt.Printf("[CardFauxDisk] - EXISTS '%s'\n", fname)
	}

	// Subdir is specified by a colon 
	colon := strings.IndexByte(fname, ':')
	if (colon == -1) {
		return c.fauxDiskExists__(fname, c.root, "")
	} else {
		dirname := fname[:colon]
		subdir := c.findSubdir(dirname)
		if (subdir == nil) {
			fmt.Printf("[CardFauxDisk] FILE NOT FOUND\n")
			return FAUX_ERR_NOT_FOUND
		}
		return c.fauxDiskExists__(fname[colon+1:], subdir, dirname + "/")
	}
}
func (c *CardFauxDisk) fauxDiskExists__(filename string, dir []fauxFile, dirName string) uint8 {
	// Find the matching file
	for i := 0; i < len(dir); i++ {
		f := dir[i]
		if (filename == strings.ToUpper(f.name)) {
			c.c800[0] = uint8(f.ftype[0]) | 0x80
			c.c800[1] = uint8(f.ftype[1]) | 0x80
			c.c800[2] = uint8(f.ftype[2]) | 0x80

			if (f.isdir) {
				c.c800[3] = '-' | 0x80
				c.c800[4] = '-' | 0x80
				c.c800[5] = '-' | 0x80
			} else {
				c.c800[3] = uint8(f.size % 1000 / 100) + ('0' | 0x80)
				c.c800[4] = uint8(f.size % 100 / 10) + ('0' | 0x80)
				c.c800[5] = uint8(f.size % 10) + ('0' | 0x80)
			}

			c.c800[6] = uint8(f.size & 0x0ff)
			c.c800[7] = uint8((f.size >> 8) & 0x0ff)
			c.c800[8] = uint8((f.size >> 16) & 0x0ff)

			return FAUX_SUCCESS
		}
	}

	// Not found, so look in subdirectories
	if (dirName == "") {
		for i := 0; i < len(c.root); i++ {
			d := c.root[i]
			if (d.isdir) {
				dir = d.subdir
				for j := 0; j < len(dir); j++ {
					f := dir[j]
					if (filename == strings.ToUpper(f.name)) {
						return c.fauxDiskExists__(filename, dir, d.filename + "/")
					}
				}
			}
		}
	}

	fmt.Printf("[CardFauxDisk] FILE NOT FOUND\n")
	return FAUX_ERR_NOT_FOUND
}


//
//  Create a file
//    0xC800 = zero-terminated name (high ASCII)
//    PARAM1 = file type (3 chars)
//	Returns
//	  RET0 = file number
//
func (c *CardFauxDisk) fauxDiskCreate() uint8 {
	if c.trace {
		fmt.Printf("[CardFauxDisk] FAUX_CREATE\n")
	}

	fname := c.c800toName()
	fnameUC := strings.ToUpper(fname)
	ftype := fmt.Sprintf("%c%c%c",
		uint8(c.param0 & 0xff), uint8((c.param0 >> 8) & 0xff), uint8((c.param0 >> 16) & 0xff))
	if c.trace {
		fmt.Printf("[CardFauxDisk] - CREATE '%s.%s'\n", fname, ftype)
	}

	// Subdir is specified by a colon 
	colon := strings.IndexByte(fname, ':')
	if (colon == -1) {
		return c.fauxDiskCreate__(fname, fnameUC, ftype, c.root, "")
	} else {
		dirname := fname[:colon]
		subdir := c.findSubdir(dirname)
		if (subdir == nil) {
			fmt.Printf("[CardFauxDisk] FILE NOT FOUND\n")
			return FAUX_ERR_NOT_FOUND
		}
		return c.fauxDiskCreate__(fname[colon+1:], fnameUC[colon+1:], ftype, subdir, dirname + "/")
	}
}
func (c *CardFauxDisk) fauxDiskCreate__(fname string, fnameUC string, ftype string, dir []fauxFile, dirName string) uint8 {
	// Check if the file already exists
	for i := 0; i < len(dir); i++ {
		f := dir[i]
		if (fnameUC == strings.ToUpper(f.name)) {
			if c.trace {
				fmt.Printf("[CardFauxDisk] ERROR: '%s' already exists\n", fname)
			}
			return FAUX_ERR_EXISTS
		}
	}

	file, err := os.Create(string(c.rootName + "/" + dirName + fname + "." + ftype))
	if (err != nil) {
		fmt.Printf("[CardFauxDisk] ERROR: %s\n", err)
		return FAUX_ERR_NOT_FOUND
	}

	c.reloadRoot()

	c.files[0] = file
	c.ret0 = 0
	return FAUX_SUCCESS
}


//
//  Open a file
//    0xC800 = zero-terminated name (high ASCII)
//	Returns
//	  RET0 = file number
//
func (c *CardFauxDisk) fauxDiskOpen() uint8 {
	if c.trace {
		fmt.Printf("[CardFauxDisk] FAUX_OPEN\n")
	}

	fname := strings.ToUpper(c.c800toName())
	if c.trace {
		fmt.Printf("[CardFauxDisk] - OPEN '%s'\n", fname)
	}

	// Subdir is specified by a colon 
	colon := strings.IndexByte(fname, ':')
	if (colon == -1) {
		return c.fauxDiskOpen__(fname, c.root, "")
	} else {
		dirname := fname[:colon]
		subdir := c.findSubdir(dirname)
		if (subdir == nil) {
			fmt.Printf("[CardFauxDisk] FILE NOT FOUND\n")
			return FAUX_ERR_NOT_FOUND
		}
		return c.fauxDiskOpen__(fname[colon+1:], subdir, dirname + "/")
	}

}
func (c *CardFauxDisk) fauxDiskOpen__(filename string, dir []fauxFile, dirName string) uint8 {
	// Find the matching file
	hasMatch := false
	catIdx := 0
	for i := 0; i < len(dir); i++ {
		f := dir[i]
		if (filename == strings.ToUpper(f.name)) {
			catIdx = i
			hasMatch = true
			break
		}
	}
	if hasMatch == false {
		// Not found, so look in subdirectories
		if (dirName == "") {
			for i := 0; i < len(c.root); i++ {
				d := c.root[i]
				if (d.isdir) {
					dir = d.subdir
					for j := 0; j < len(dir); j++ {
						f := dir[j]
						if (filename == strings.ToUpper(f.name)) {
							return c.fauxDiskOpen__(filename, dir, d.filename + "/")
						}
					}
				}
			}
		}

		fmt.Printf("[CardFauxDisk] FILE NOT FOUND\n")
		return FAUX_ERR_NOT_FOUND
	}

fmt.Printf("  OPEN '%s'\n", c.rootName + "/" + dirName + dir[catIdx].filename)
	file, err := os.Open(string(c.rootName + "/" + dirName + dir[catIdx].filename))
	if (err != nil) {
		fmt.Printf("[CardFauxDisk] ERROR: %s\n", err)
		return FAUX_ERR_NOT_FOUND
	}

	c.files[0] = file
	c.ret0 = 0
	return FAUX_SUCCESS
}


//
//  Read from a file
//	  PARAM1 = FN
//    0xC800 = bytes read
//
func (c *CardFauxDisk) fauxDiskRead() uint8 {
	if c.trace {
		fmt.Printf("[CardFauxDisk] FAUX_READ\n")
	}

	buffer := make([]byte, 1024)
	actual, err := c.files[int(c.param0)].Read(buffer)
	if (actual == 0) {
		fmt.Printf("[CardFauxDisk] END OF FILE\n")
		return FAUX_END_OF_FILE
	}
	if (err != nil) {
		fmt.Printf("[CardFauxDisk] ERROR: %s\n", err)
		return FAUX_ERR_READ_ERROR
	}
	if c.trace {
		fmt.Printf("[CardFauxDisk] READ %d/0x%x bytes\n", actual, actual)
	}

	// Special case to reformat 960 byte LORES files
	/*
	if (actual == 960) {
		if c.trace {
			fmt.Printf("[CardFauxDisk] -  960 byte LORES FILE\n")
		}

		fixed := make([]byte, 1024)
		i := 0
		for r := 0; r < 24; r++ {
			var base int
			switch r {
				case 0: base = 0
				case 1: base = 0x80
				case 2: base = 0x100
				case 3: base = 0x180
				case 4: base = 0x200
				case 5: base = 0x280
				case 6: base = 0x300
				case 7: base = 0x380
				case 8: base = 0x28
				case 9: base = 0xA8
				case 10: base = 0x128
				case 11: base = 0x1A8
				case 12: base = 0x228
				case 13: base = 0x2A8
				case 14: base = 0x328
				case 15: base = 0x3A8
				case 16: base = 0x050
				case 17: base = 0x0D0
				case 18: base = 0x150
				case 19: base = 0x1D0
				case 20: base = 0x250
				case 21: base = 0x2D0
				case 22: base = 0x350
				case 23: base = 0x3D0
			}
			
			for c := 0; c < 40; c++ {
				fixed[base + c] = buffer[i]
				i += 1
			}
		}
		buffer = fixed
		actual = 1024
	}
	*/

	c.ret0 = uint32(actual)
	addr := uint32(0xc800)
	for i := 0; i < actual; i++ {
		c.c800[addr-0xc800] = buffer[i]
		addr += 1

		// @@@ TODO - Add wait to simulate clock cycles
		time.Sleep(10 * time.Microsecond)
	}

	return FAUX_SUCCESS
}


//
//  Read from a file and DMA the contents to a specified address @@@ TBD
//
func (c *CardFauxDisk) fauxDiskReadDMA() uint8 {
	if c.trace {
		fmt.Printf("[CardFauxDisk] FAUX_READ_DMA\n")
	}

	dest := c.param0
	addr := dest
	buffer := make([]byte, 256*1024)
	for {
		actual, err := c.files[int(c.param0)].Read(buffer)
		if (err != nil) {
			fmt.Printf("[CardFauxDisk] ERROR: %s\n", err)
			return FAUX_ERR_NOT_FOUND
		}
		if (actual == 0) {
			break
		}

		// Copy directly to CPU RAM
		for i := 0; i < actual; i++ {
			c.a.mmu.Poke(addr, buffer[i])
			addr += 1

			// @@@ TODO - Add wait to simulate clock cycles
			time.Sleep(10 * time.Microsecond)
		}
	}

	return FAUX_SUCCESS
}


//
//  Write to a file
//	  PARAM1 = FN
//    PARAM2 = number of bytes to write
//    0xC800 = bytes to write
//
func (c *CardFauxDisk) fauxDiskWrite() uint8 {
	if c.trace {
		fmt.Printf("[CardFauxDisk] FAUX_WRITE\n")
	}

	buffer := make([]uint8, c.param1)
	for i := 0; i < int(c.param1); i++ {
		buffer[i] = c.c800[i]
	}
	actual, err := c.files[int(c.param0)].Write(buffer)
	if (err != nil) {
		fmt.Printf("[CardFauxDisk] ERROR: %s\n", err)
		return FAUX_ERR_WRITE_ERROR
	}
	if (actual != int(c.param1)) {
		fmt.Printf("[CardFauxDisk] ERROR: %d bytes written instead of %d\n", actual, c.param1)
		return FAUX_ERR_WRITE_ERROR
	}

	// @@@ TODO - Add wait to simulate clock cycles
	time.Sleep(1 * time.Microsecond)

	return FAUX_SUCCESS
}


//
//  Write to a file with DMA reading from memory
//
func (c *CardFauxDisk) fauxDiskWriteDMA() uint8 {
	if c.trace {
		fmt.Printf("[CardFauxDisk] FAUX_WRITE_DMA\n")
	}

/*
*/

	return FAUX_SUCCESS
}


//
//  Close a file @@@ FD
//
func (c *CardFauxDisk) fauxDiskClose() uint8 {
	if c.trace {
		fmt.Printf("[CardFauxDisk] FAUX_CLOSE\n")
	}

	c.files[c.param0].Close()
	return FAUX_SUCCESS
}


//
//  Copy the filename from $C800 to a string
//
func (c *CardFauxDisk) c800toName() string {
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


//
//  This card's ROM ($C700-$C7FF)
//
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
