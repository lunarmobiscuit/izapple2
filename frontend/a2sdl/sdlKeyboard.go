package main

import (
	"fmt"

	"github.com/lunarmobiscuit/izapple2"
	"github.com/lunarmobiscuit/izapple2/screen"
	"github.com/veandco/go-sdl2/sdl"
)

type sdlKeyboard struct {
	a          *izapple2.Apple2
	keyChannel *izapple2.KeyboardChannel

	showPages   bool
	showCharGen bool
	showAltText bool
	screenMode  int

	modeMemory		bool
	modeBreakpoint	bool
	modeValue		uint32
}

func newSDLKeyBoard(a *izapple2.Apple2) *sdlKeyboard {
	var k sdlKeyboard
	k.a = a
	k.keyChannel = izapple2.NewKeyboardChannel(a)

	k.screenMode = screen.ScreenModePlain
	return &k
}

func (k *sdlKeyboard) putText(text string) {
	k.keyChannel.PutText(text)
}

func (k *sdlKeyboard) putKey(keyEvent *sdl.KeyboardEvent) {
	/*
		See "Apple II reference manual", page 5

		To get keys as understood by the Apple2 hardware run:
		10 A=PEEK(49152)
		20 PRINT A, A - 128
		30 GOTO 10
	*/
	if keyEvent.Type != sdl.KEYDOWN {
		// Process only key pushes
		return
	}

	key := keyEvent.Keysym
	ctrl := key.Mod&sdl.KMOD_CTRL != 0
	shift := key.Mod&sdl.KMOD_SHIFT != 0

	// Step-by-step debugging mode
	if k.a.IsPaused() {
		switch key.Sym {
		case ' ':
			k.a.SendCommand(izapple2.CommandNextStep)
		case sdl.K_RETURN:
			if k.modeMemory {
				address := k.modeValue & 0x0FFFFF0
				fmt.Printf("0x%06x :: ", address)
				for i := 0; i < 16; i++ {
					fmt.Printf("%02x ", k.a.Peek(address + uint32(i)))
				}
				fmt.Printf("  ")
				for i := 0; i < 16; i++ {
					ch := k.a.Peek(address + uint32(i))
					if (ch == 0) { ch = '.' } else if (ch < 32) { ch += 32 }
					fmt.Printf("%c", ch)
				}
				fmt.Printf("\n")
				k.modeValue += 16
			} else if k.modeBreakpoint {
				k.modeBreakpoint = false
				fmt.Printf("0x%x :: \n", k.modeValue)
				k.a.SetUntilPC(k.modeValue)
			}
		case '0','1','2','3','4','5','6','7','8','9':
			if k.modeMemory || k.modeBreakpoint {
				k.modeValue = (k.modeValue * 16) + uint32((key.Sym - '0'))
			}
		case 'A','B','C','D','E','F':
			if k.modeMemory || k.modeBreakpoint {
				k.modeValue = (k.modeValue * 16) + uint32((key.Sym - 'A' + 10))
			}
		case 'a','b','c','d','e','f':
			if k.modeMemory || k.modeBreakpoint {
				k.modeValue = (k.modeValue * 16) + uint32((key.Sym - 'a' + 10))
			}
		case 'M','m':
			fmt.Printf("*** MEMORY: \n")
			k.modeMemory = true
			k.modeBreakpoint = false
			k.modeValue = 0
		case 'P','p':
			fmt.Printf("*** Run to PC: ")
			k.modeMemory = false
			k.modeBreakpoint = true
			k.modeValue = 0
		case '.':
			fmt.Printf("*** Run to PC (again)\n", )
			k.a.SetUntilPC(0xffffff)
		case 'R', 'r', sdl.K_ESCAPE:
			k.a.SendCommand(izapple2.CommandStart)
			k.a.SendCommand(izapple2.CommandCPUTraceOff)
		}

		return
	}

	if ctrl {
		if key.Sym >= 'a' && key.Sym <= 'z' {
			k.keyChannel.PutChar(uint8(key.Sym) - 97 + 1)
			return
		}
	}

	result := uint8(0)

	switch key.Sym {
	case sdl.K_ESCAPE:
		if (ctrl) {
			k.a.SendCommand(izapple2.CommandReset)
		} else if (shift) {
			k.a.SendCommand(izapple2.CommandCPUTraceOn)
			k.a.SendCommand(izapple2.CommandStep)
		} else {
			result = 27
		}
	case sdl.K_BACKSPACE:
		if (ctrl && shift) {
			k.a.SendCommand(izapple2.CommandReset)
		} else {
			result = 127 // was 8 = LEFT, but those two keys should be different behaviors
		}
	case sdl.K_RETURN:
		result = 13
	case sdl.K_RETURN2:
		result = 13
	case sdl.K_LEFT:
		if ctrl {
			result = 31 // Base64A
		} else {
			result = 8
		}
	case sdl.K_RIGHT:
		result = 21

	// Apple //e
	case sdl.K_UP:
		result = 11 // 31 in the Base64A
	case sdl.K_DOWN:
		result = 10
	case sdl.K_TAB:
		result = 9
	case sdl.K_DELETE:
		result = 127 // 24 in the Base64A

	// Base64A clone particularities
	case sdl.K_F2:
		result = 127 // Base64A

	// Control of the emulator
	case sdl.K_F1:
		if ctrl {
			k.a.SendCommand(izapple2.CommandReset)
		}
	case sdl.K_F5:
		if ctrl {
			k.a.SendCommand(izapple2.CommandShowSpeed)
		} else {
			k.a.SendCommand(izapple2.CommandToggleSpeed)
		}
	case sdl.K_F6:
		if k.screenMode == screen.ScreenModeNTSC {
			k.screenMode = screen.ScreenModeGreen
		} else {
			k.screenMode = screen.ScreenModeNTSC
		}
	case sdl.K_F7:
		k.showPages = !k.showPages
	case sdl.K_F9:
		k.a.SendCommand(izapple2.CommandDumpDebugInfo)
	case sdl.K_F10:
		if ctrl {
			k.showCharGen = !k.showCharGen
		} else if shift {
			k.showAltText = !k.showAltText
		} else {
			k.a.SendCommand(izapple2.CommandNextCharGenPage)
		}
	case sdl.K_F11:
		k.a.SendCommand(izapple2.CommandToggleCPUTrace)
	case sdl.K_F12:
		fallthrough
	case sdl.K_PRINTSCREEN:
		err := screen.SaveSnapshot(k.a, screen.ScreenModeNTSC, "snapshot.png")
		if err != nil {
			fmt.Printf("Error saving snapshoot: %v.\n.", err)
		} else {
			fmt.Println("Saving snapshot 'snapshot.png'")
		}
	case sdl.K_PAUSE:
		k.a.SendCommand(izapple2.CommandPauseUnpause)
	}

	// Missing values 91 to 95. Usually control for [\]^_
	// On the Base64A it's control for \]./

	if result != 0 {
		k.keyChannel.PutChar(result)
	}
}
