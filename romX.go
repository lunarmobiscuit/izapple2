package izapple2

import (
	"fmt"

	"github.com/lunarmobiscuit/iz6502"
)

/*
RomX from https://theromexchange.com/
This complement uses the RomX API spec to switch main ROM and character generator ROM

Only the font switch is implemented

See:
	https://theromexchange.com/documentation/ROM%20X%20API%20Reference.pdf
	https://theromexchange.com/downloads/ROM%20X%2020-10-22.zip
	https://theromexchange.com/documentation/romxce/ROMXce%20API%20Reference.pdf

For romX:
It is not enough to intercept the ROM accesses. RomX intercept the 74LS138 in
position F12, that has access to the full 0xc000-0xf000 on the Apple II+

Firmware:
	- It first copies $D000-$DFFF to $6000 and runs there.

go run *.go -rom ROMX.FIRM.dump -disk ROM\ X\ 20-10-22.dsk


*/

type romX struct {
	a              *Apple2
	memory         iz6502.Memory
	activationStep int
	systemBank     uint8
	mainBank       uint8
	tempBank       uint8
	textBank       uint8
	debug          bool
}

var romXActivationSequence = []uint32{0x0caca, 0x0caca, 0x0cafe}
var romXceActivationSequence = []uint32{0x0faca, 0x0faca, 0x0fafe}

const (
	romxSetupBank                    = uint8(0)
	romXPlusSetSystemBankBaseAddress = uint32(0x0cef0)
	romXPlusSetTextBankBaseAddress   = uint32(0x0cfd0)

	// Unknown
	//romXFirmwareMark0Address = uint32(0x0dffe)
	//romXFirmwareMark0Value   = uint8(0x4a)
	//romXFirmwareMark1Address = uint32(0x0dfff)
	//romXFirmwareMark1Value   = uint8(0xcd)

	romXceSelectTempBank  = uint32(0x0f850)
	romXceSelectMainBank  = uint32(0x0f851)
	romXceSetTempBank     = uint32(0x0f830) // 16 positions
	romXceSetMainBank     = uint32(0x0f800) // 16 positions
	romXcePresetTextBank  = uint32(0x0f810) // 16 positions
	romXceMCP7940SDC      = uint32(0x0f860) // 16 positions
	romXceLowerUpperBanks = uint32(0x0f820) // 16 positions

	romXGetDefaultSystemBank = uint32(0x0d034) // $00 to $0f
	romXGetDefaultTextBank   = uint32(0x0d02e) // $10 to $1f
	romXGetCurrentBootDelay  = uint32(0x0deca) // $00 to $0f

	/*
		romXceEntryPointSetClock      = uint32(0x0c803)
		romXceEntryPointReadClock     = uint32(0x0c803)
		romXceEntryPointLauncherToRam = uint32(0x0dfd9)
		romXceEntryPointLauncher      = uint32(0x0dfd0)
	*/
)

func newRomX(a *Apple2, memory iz6502.Memory) (*romX, error) {
	var rx romX
	rx.a = a
	rx.memory = memory
	rx.systemBank = 1
	rx.mainBank = 1
	rx.tempBank = 1
	rx.textBank = 0
	rx.debug = true

	if a.isApple2e {
		err := a.cg.load("<internal>/ROMXce Production 1Mb Text ROM V5.bin")
		if err != nil {
			return nil, err
		}
	}

	return &rx, nil
}

func (rx *romX) Peek(address uint32) uint8 {
	intercepted, value := rx.interceptAccess(address)
	if intercepted {
		return value
	}
	return rx.memory.Peek(address)
}

func (rx *romX) PeekCode(address uint32) uint8 {
	intercepted, value := rx.interceptAccess(address)
	if intercepted {
		return value
	}
	return rx.memory.PeekCode(address)
}

func (rx *romX) Poke(address uint32, value uint8) {
	rx.interceptAccess(address)
	rx.memory.Poke(address, value)
}

func (rx *romX) logf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("[romX]%s\n", msg)
}

func (rx *romX) interceptAccess(address uint32) (bool, uint8) {
	// Intercept only $C080 to $FFFF as seen by the F12 chip
	if address < 0xc080 {
		return false, 0
	}

	// Setup mode when the setup bank is active
	if rx.systemBank == romxSetupBank {

		// Range commands
		nibble := uint8(address & 0xf)
		switch address & 0xfff0 {
		case romXceSetMainBank:
			rx.mainBank = nibble
			rx.logf("Main bank set to $%x", nibble)
		case romXcePresetTextBank:
			textBank := int(nibble)
			rx.a.cg.setPage(textBank)
			rx.logf("[romX]Text bank set to $%x", nibble)
		case romXceLowerUpperBanks:
			rx.logf("Configure lower upper banks $%x", address)
		case romXceSetTempBank:
			rx.tempBank = nibble
			rx.logf("Temp bank set to $%x", nibble)
		case romXceMCP7940SDC:
			rx.logf("Configure MCP7940 $%x", address)
		}

		// More commands
		switch address {
		case romXceSelectTempBank:
			rx.systemBank = rx.tempBank
			rx.logf("System bank set to temp bank $%x", rx.systemBank)
		case romXceSelectMainBank:
			rx.systemBank = rx.mainBank
			rx.logf("System bank set to main bank $%x", rx.systemBank)
		}

		// Queries
		switch address {
		case romXGetDefaultSystemBank:
			bank := rx.systemBank
			rx.logf("Peek in $%04x, current system bank %v", address, bank)
			return true, bank
		case romXGetDefaultTextBank:
			page := uint8(rx.a.cg.getPage() & 0xf)
			rx.logf("PeeK in $%04x, current text bank %v", address, page)
			return true, 0x10 + page
		case romXGetCurrentBootDelay:
			delay := uint8(5) // We don't care
			rx.logf("PeeK in $%04x, current boot delay %v", address, delay)
			return true, delay
		}

		return false, 0
	}

	// Activation sequence detection
	if address == romXceActivationSequence[rx.activationStep] {
		rx.activationStep++
		rx.logf("Activation step %v", rx.activationStep)
		if rx.activationStep == len(romXActivationSequence) {
			// Activation sequence completed
			rx.systemBank = romxSetupBank
			rx.activationStep = 0
			rx.logf("System bank set to 0, %v", rx.systemBank)
		}
	} else {
		rx.activationStep = 0
	}

	return false, 0
}
