package izapple2

/*
 See:
   https://www.apple.asimov.net/documentation/hardware/machines/APPLE%20IIe%20Auxiliary%20Memory%20Softswitches.pdf
*/

const (
	ioFlag64Col   uint8 = 0x6E
)

func addApple24SoftSwitches(io *ioC0Page) {
	// Reuse the CASSETTE as the switch to/from 80COL mode
	io.addSoftSwitchRW(0x60, getSoftSwitch(io, ioFlag80Col, false), "80COLOFF")
	io.addSoftSwitchRW(0x68, getSoftSwitch(io, ioFlag80Col, true), "80COLON")
	io.addSoftSwitchR(0x6E, getSoftSwitch(io, ioFlag64Col, false), "64COLOFF")
	io.addSoftSwitchR(0x6F, getSoftSwitch(io, ioFlag64Col, true), "64COLON")
}

