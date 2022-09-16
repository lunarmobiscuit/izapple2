package izapple2

/*
 See:
   https://www.apple.asimov.net/documentation/hardware/machines/APPLE%20IIe%20Auxiliary%20Memory%20Softswitches.pdf
*/

func addApple24SoftSwitches(io *ioC0Page) {
	// Reuse the CASSETTE as the switch to/from 80COL mode
	io.addSoftSwitchRW(0x60, getSoftSwitch(io, ioFlag80Col, false), "80COLOFF")
	io.addSoftSwitchRW(0x68, getSoftSwitch(io, ioFlag80Col, true), "80COLON")
}

