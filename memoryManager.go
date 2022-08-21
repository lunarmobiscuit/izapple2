package izapple2

import "fmt"

// See https://fabiensanglard.net/fd_proxy/prince_of_persia/Inside%20the%20Apple%20IIe.pdf
// See https://i.stack.imgur.com/yn21s.gif

type memoryManager struct {
	apple2 *Apple2

	// Main RAM area: 0x0000 to 0xbfff
	physicalMainRAM *memoryRange // 0x0000 to 0xbfff, Up to 48 Kb

	// Slots area: 0xc000 to 0xcfff
	cardsROM      [8]memoryHandler //0xcs00 to 0xcSff. 256 bytes for each card
	cardsROMExtra [8]memoryHandler // 0xc800 to 0xcfff. 2048 bytes for each card

	// Upper area ROM: 0xc000 to 0xffff (or 0xd000 to 0xffff on the II+)
	physicalROM [4]memoryHandler // 0xc000 (or 0xd000) to 0xffff, 16 (or 12) Kb. Up to four banks

	// Language card upper area RAM: 0xd000 to 0xffff. One bank for regular LC cards, up to 8 with Saturn
	physicalLangRAM    []*memoryRange // 0xd000 to 0xffff, 12KB. Up to 8 banks.
	physicalLangAltRAM []*memoryRange // 0xd000 to 0xdfff, 4KB. Up to 8 banks.

	// Extended RAM: 0x0000 to 0xffff (with 4Kb moved from 0xc000 to 0xd000 alt). One bank for extended Apple 2e card, up to 256 with RamWorks
	physicalExtRAM    []*memoryRange // 0x0000 to 0xffff. 60Kb, 0xc000 to 0xcfff not used. Up to 256 banks
	physicalExtAltRAM []*memoryRange // 0xd000 to 0xdfff, 4Kb. Up to 256 banks.

	// Extended RAM: 0x0000 to 0xffff (with 4Kb moved from 0xc000 to 0xd000 alt). One bank for extended Apple 2e card, up to 256 with RamWorks
	physical24bitRAM  *memoryRange // 0x010000 to 0xfeffff. up to 16MB. Up to 256 banks
	addressLimit24BitRAM uint32 // top-most installed address of 24-bit RAM
	physical24bitROM  *memoryRangeROM // 0xff0000 to 0xffffff. 64K

	// Configuration switches, Language cards
	lcSelectedBlock uint8 // Language card block selected. Usually, allways 0. But Saturn has 8
	lcActiveRead    bool  // Upper RAM active for read
	lcActiveWrite   bool  // Upper RAM active for write
	lcAltBank       bool  // Alternate

	// Configuration switches, Apple //e
	altZeroPage           bool          // Use extra RAM from 0x0000 to 0x01ff. And additional language card block
	altMainRAMActiveRead  bool          // Use extra RAM from 0x0200 to 0xbfff for read
	altMainRAMActiveWrite bool          // Use extra RAM from 0x0200 to 0xbfff for write
	store80Active         bool          // Special pagination for text and graphics areas
	slotC3ROMActive       bool          // Apple2e slot 3  ROM shadow
	intCxROMActive        bool          // Apple2e slots internal ROM shadow
	intC8ROMActive        bool          // C8Rom associated to the internal slot 3. Softswitch not directly accessible. See UtA2e 5-28
	activeSlot            uint8         // Active slot owner of 0xc800 to 0xcfff
	extendedRAMBlock      uint8         // Block used for entended memory for RAMWorks cards
	mainROMinhibited      memoryHandler // Alternative ROM from 0xd000 to 0xffff provided by a card with the INH signal.

	// Configuration switches, Base64A
	romPage uint8 // Active ROM page

	// Resolution cache
	lastAddressPage    uint32 // The first byte is the page. The second is zero when the cached is valid.
	lastAddressHandler memoryHandler
}

const (
	ioC8Off                uint32 = 0x0cfff
	addressLimitZero       uint32 = 0x001ff
	addressStartText       uint32 = 0x00400
	addressLimitText       uint32 = 0x007ff
	addressStartHgr        uint32 = 0x02000
	addressLimitHgr        uint32 = 0x03fff
	addressLimitMainRAM    uint32 = 0x0bfff
	addressLimitIO         uint32 = 0x0c0ff
	addressLimitSlots      uint32 = 0x0c7ff
	addressLimitSlotsExtra uint32 = 0x0cfff
	addressLimitDArea      uint32 = 0x0dfff
	address24BitROM        uint32 = 0xff0000

	invalidAddressPage uint32 = 0x00001
)

type memoryHandler interface {
	peek(uint32) uint8
	poke(uint32, uint8)
	setBase(uint32)
}

func newMemoryManager(a *Apple2) *memoryManager {
	var mmu memoryManager
	mmu.apple2 = a
	mmu.physicalMainRAM = newMemoryRange(0, make([]uint8, 0xc000), "Main RAM")

	mmu.slotC3ROMActive = true // For II+, this is the default behaviour

	return &mmu
}

func (mmu *memoryManager) add24BitMemory(size uint32) {
	mmu.physical24bitRAM = newMemoryRange(0x010000, make([]uint8, size - 0x010000), "24-bit RAM")
	mmu.addressLimit24BitRAM = size;
}

func (mmu *memoryManager) accessCArea(address uint32) memoryHandler {
	slot := uint8((address >> 8) & 0x0f)

	// Internal IIe slot 3
	if (address <= addressLimitSlots) && !mmu.slotC3ROMActive && (slot == 3) {
		mmu.intC8ROMActive = true
		return mmu.physicalROM[mmu.romPage]
	}

	// Internal IIe CxROM
	if mmu.intCxROMActive {
		return mmu.physicalROM[mmu.romPage]
	}

	// First slot area
	if slot <= 7 {
		mmu.activeSlot = slot
		mmu.intC8ROMActive = false
		return mmu.cardsROM[slot]
	}

	// Extra slot area reset
	if address == ioC8Off {
		// Reset extra slot area owner
		mmu.activeSlot = 0
		mmu.intC8ROMActive = false
	}

	// Extra slot area
	if mmu.intC8ROMActive {
		return mmu.physicalROM[mmu.romPage]
	}
	return mmu.cardsROMExtra[mmu.activeSlot]
}

func (mmu *memoryManager) accessUpperRAMArea(address uint32) memoryHandler {
	if mmu.altZeroPage {
		// Use extended RAM
		block := mmu.extendedRAMBlock
		if mmu.lcAltBank && address <= addressLimitDArea {
fmt.Printf("LC %x -> physicalExtAltRAM[%x]\n", address, block)
			return mmu.physicalExtAltRAM[block]
		}
fmt.Printf("LC %x -> physicalExtRAM[%x]\n", address, mmu.extendedRAMBlock)
		return mmu.physicalExtRAM[mmu.extendedRAMBlock]
	}

	// Use language card
	block := mmu.lcSelectedBlock
	if mmu.lcAltBank && address <= addressLimitDArea {
fmt.Printf("LC %x -> physicalLangAltRAM[%x]\n", address, block)
		return mmu.physicalLangAltRAM[block]
	}
fmt.Printf("LC %x -> physicalLangRAM[%x]\n", address, block)
	return mmu.physicalLangRAM[block]
}

func (mmu *memoryManager) getPhysicalMainRAM(ext bool) memoryHandler {
	if ext {
		return mmu.physicalExtRAM[mmu.extendedRAMBlock]
	}
	return mmu.physicalMainRAM
}

func (mmu *memoryManager) getPhysical24BitRAM(ext bool) memoryHandler {
	return mmu.physical24bitRAM
}

func (mmu *memoryManager) getVideoRAM(ext bool) *memoryRange {
	if ext {
		// The video memory uses the first extended RAM block, even with RAMWorks
		return mmu.physicalExtRAM[0]
	}
	return mmu.physicalMainRAM
}

func (mmu *memoryManager) inhibitROM(replacement memoryHandler) {
	// If a card INH the ROM, it replaces the ROM and the LC RAM
	mmu.mainROMinhibited = replacement
	mmu.lastAddressPage = invalidAddressPage // Invalidate cache
}

func (mmu *memoryManager) accessRead(address uint32) memoryHandler {
	if address <= addressLimitZero {
//fmt.Printf("READ(0x%x) ZERO-PAGE\n", address) // @@@
		return mmu.getPhysicalMainRAM(mmu.altZeroPage)
	}
	if mmu.store80Active && address <= addressLimitHgr {
		altPage := mmu.apple2.io.isSoftSwitchActive(ioFlagSecondPage) // TODO: move flag to mmu property like the store80
		if address >= addressStartText && address <= addressLimitText {
			return mmu.getPhysicalMainRAM(altPage)
		}
		hires := mmu.apple2.io.isSoftSwitchActive(ioFlagHiRes)
		if hires && address >= addressStartHgr && address <= addressLimitHgr {
			return mmu.getPhysicalMainRAM(altPage)
		}
	}
	if address <= addressLimitMainRAM {
		return mmu.getPhysicalMainRAM(mmu.altMainRAMActiveRead)
	}
	if address <= addressLimitIO {
		mmu.lastAddressPage = invalidAddressPage
		return mmu.apple2.io
	}
	if address <= addressLimitSlotsExtra {
		return mmu.accessCArea(address)
	}
	if mmu.mainROMinhibited != nil {
		return mmu.mainROMinhibited
	}
	if mmu.lcActiveRead && (address <= 0x0FFFF) {
		return mmu.accessUpperRAMArea(address)
	}
	if address <= 0x0FFFF {
//fmt.Printf("READ(0x%x) 64K RAM\n", address) // @@@
		return mmu.physicalROM[mmu.romPage]
	}
	if address < mmu.addressLimit24BitRAM {
//fmt.Printf("READ(0x%x) 24-bit RAM\n", address) // @@@
		return mmu.physical24bitRAM
	}
	if address >= address24BitROM {
//fmt.Printf("READ(0x%x) 24-bit ROM\n", address) // @@@
		return mmu.physical24bitROM
	}
fmt.Printf("READ(0x%x) INVALID MEMORY\n", address) // @@@
	return nil
}

func (mmu *memoryManager) accessWrite(address uint32) memoryHandler {
	if address <= addressLimitZero {
//fmt.Printf("WRITE(0x%x) ZERO-PAGE\n", address) // @@@
		return mmu.getPhysicalMainRAM(mmu.altZeroPage)
	}
	if address <= addressLimitHgr && mmu.store80Active {
		altPage := mmu.apple2.io.isSoftSwitchActive(ioFlagSecondPage)
		if address >= addressStartText && address <= addressLimitText {
			return mmu.getPhysicalMainRAM(altPage)
		}
		hires := mmu.apple2.io.isSoftSwitchActive(ioFlagHiRes)
		if hires && address >= addressStartHgr && address <= addressLimitHgr {
			return mmu.getPhysicalMainRAM(altPage)
		}
	}
	if address <= addressLimitMainRAM {
		return mmu.getPhysicalMainRAM(mmu.altMainRAMActiveWrite)
	}
	if address <= addressLimitIO {
		mmu.lastAddressPage = invalidAddressPage
		return mmu.apple2.io
	}
	if address <= addressLimitSlotsExtra {
		return mmu.accessCArea(address)
	}
	if mmu.mainROMinhibited != nil {
		return mmu.mainROMinhibited
	}
	if mmu.lcActiveWrite && (address <= 0x0FFFF) {
		return mmu.accessUpperRAMArea(address)
	}
	if address <= 0x0FFFF {
//fmt.Printf("WRITE(0x%x) 64K RAM\n", address) // @@@
		return mmu.physicalROM[mmu.romPage]
	}
	if address < mmu.addressLimit24BitRAM {
//fmt.Printf("WRITE(0x%x) 24-bit RAM\n", address) // @@@
		return mmu.physical24bitRAM
	}
	if address >= address24BitROM {
//fmt.Printf("WRITE(0x%x) 24-bit ROM\n", address) // @@@
		return mmu.physical24bitROM
	}
fmt.Printf("WRITE(0x%x) INVALID MEMORY\n", address) // @@@
	return nil
}

func (mmu *memoryManager) peekWord(address uint32) uint16 {
	return uint16(mmu.Peek(address)) +
		uint16(mmu.Peek(address+1))<<8

	//return uint16(mmu.Peek(address)) +
	//    0x100*uint16(mmu.Peek(address+1))

}

// Peek returns the data on the given address
func (mmu *memoryManager) Peek(address uint32) uint8 {
//fmt.Printf("\n  Peek(0x%x)", address) // @@@
	mh := mmu.accessRead(address)
	if mh == nil {
		return 0xff // Or some random number
	}
	value := mh.peek(address)
//fmt.Printf(" = %02x\n", value) // @@@
	//if address >= 0xc400 && address < 0xc500 {
	//	fmt.Printf("[MMU] Peek at %04x: %02x\n", address, value)
	//}

	return value
}

// Peek returns the data on the given address optimized for more local requests
func (mmu *memoryManager) PeekCode(address uint32) uint8 {
	page := address & 0xffff00
//fmt.Printf("  PeekCode(0x%x)  page:%x\n", address, page) // @@@
	var mh memoryHandler
	if page == mmu.lastAddressPage {
		mh = mmu.lastAddressHandler
	} else {
		mh = mmu.accessRead(address)
		if address&0xf000 != 0xc000 {
			// Do not cache 0xC area as it may reconfigure the MMU
			mmu.lastAddressPage = page
			mmu.lastAddressHandler = mh
		}
	}

	if mh == nil {
		return 0xff // Or some random number
	}

	value := mh.peek(address)
	//if address >= 0xc400 && address < 0xc500 {
	//	fmt.Printf("[MMU] PeekCode at %04x: %02x\n", address, value)
	//}

	return value
}

// Poke sets the data at the given address
func (mmu *memoryManager) Poke(address uint32, value uint8) {
//fmt.Printf("\n  Poke(0x%x) <- %02x\n", address, value) // @@@
	mh := mmu.accessWrite(address)
	if mh != nil {
		mh.poke(address, value)
	}

	//if address >= 0x0036 && address <= 0x0039 {
	//	fmt.Printf("[MMU] Poke at %04x: %02x\n", address, value)
	//}
}

// Memory initialization
func (mmu *memoryManager) setCardROM(slot int, mh memoryHandler) {
	mmu.cardsROM[slot] = mh
}

func (mmu *memoryManager) setCardROMExtra(slot int, mh memoryHandler) {
	mmu.cardsROMExtra[slot] = mh
}

func (mmu *memoryManager) initLanguageRAM(groups uint8) {
	// Apple II+ language card or Saturn (up to 8 groups)
	mmu.physicalLangRAM = make([]*memoryRange, groups)
	mmu.physicalLangAltRAM = make([]*memoryRange, groups)
	for i := uint8(0); i < groups; i++ {
		mmu.physicalLangRAM[i] = newMemoryRange(0xd000, make([]uint8, 0x3000), fmt.Sprintf("LC RAM block %v", i))
		mmu.physicalLangAltRAM[i] = newMemoryRange(0xd000, make([]uint8, 0x1000), fmt.Sprintf("LC RAM Alt block %v", i))
	}
}

func (mmu *memoryManager) initExtendedRAM(groups int) {
	// Apple IIe 80 col card with 64Kb style RAM or RAMWorks (up to 256 banks)
	mmu.physicalExtRAM = make([]*memoryRange, groups)
	mmu.physicalExtAltRAM = make([]*memoryRange, groups)
	for i := 0; i < groups; i++ {
		mmu.physicalExtRAM[i] = newMemoryRange(0, make([]uint8, 0x10000), fmt.Sprintf("Extra RAM block %v", i))
		mmu.physicalExtAltRAM[i] = newMemoryRange(0xd000, make([]uint8, 0x1000), fmt.Sprintf("Extra RAM Alt block %v", i))
	}
}

// Memory configuration
func (mmu *memoryManager) setActiveROMPage(page uint8) {
	mmu.romPage = page
}

func (mmu *memoryManager) getActiveROMPage() uint8 {
	return mmu.romPage
}

func (mmu *memoryManager) setLanguageRAM(readActive bool, writeActive bool, altBank bool) {
	mmu.lcActiveRead = readActive
	mmu.lcActiveWrite = writeActive
	mmu.lcAltBank = altBank
}

func (mmu *memoryManager) setLanguageRAMActiveBlock(block uint8) {
	block = block % uint8(len(mmu.physicalLangRAM))
	mmu.lcSelectedBlock = block
}

func (mmu *memoryManager) setExtendedRAMActiveBlock(block uint8) {
	if int(block) >= len(mmu.physicalExtRAM) {
		// How does the real hardware reacts?
		block = 0
	}
	mmu.extendedRAMBlock = block
}

func (mmu *memoryManager) reset() {
	if mmu.apple2.isApple2e {
		// MMU UtA2e 4-14, 5-22
		mmu.altZeroPage = false
		mmu.altMainRAMActiveRead = false
		mmu.altMainRAMActiveWrite = false
		mmu.store80Active = false
		mmu.slotC3ROMActive = false
		mmu.intCxROMActive = false
		mmu.intC8ROMActive = false

		// IOU UtaA2e 7-3
		// "All softswitches except KEYSTROKE, TEXT and MIXED are reset
		// when the RESET line drops low"
		mmu.apple2.io.softSwitchesData[ioFlagSecondPage] = ssOff
		mmu.apple2.io.softSwitchesData[ioFlagHiRes] = ssOff
		mmu.apple2.io.softSwitchesData[ioFlag80Col] = ssOff
		mmu.apple2.io.softSwitchesData[ioDataNewVideo] = ssOff
		// ioFlagText ?
	}
}
