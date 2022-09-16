package main

import (
	"fmt"
	"image"
	"os"
	"unsafe"

	"github.com/lunarmobiscuit/izapple2"
	"github.com/lunarmobiscuit/izapple2/screen"

	"github.com/pkg/profile"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	a := izapple2.MainApple()
	if a != nil {
		if a.IsProfiling() {
			// See the log with:
			//    go tool pprof --pdf ~/go/bin/izapple2sdl /tmp/profile329536248/cpu.pprof > profile.pdf
			defer profile.Start().Stop()
		}

		sdlRun(a)
	}
}

func sdlRun(a *izapple2.Apple2) {

	window, renderer, err := sdl.CreateWindowAndRenderer(3*40*7+8, 3*24*8, sdl.WINDOW_SHOWN)
	if err != nil {
		panic("Failed to create window")
	}
	window.SetResizable(true)

	defer window.Destroy()
	defer renderer.Destroy()
	window.SetTitle("iz-" + a.Name)

	kp := newSDLKeyBoard(a)

	s := newSDLSpeaker()
	s.start()
	a.SetSpeakerProvider(s)

	j := newSDLJoysticks(true)
	a.SetJoysticksProvider(j)

	m := newSDLMouse()
	a.SetMouseProvider(m)

	go a.Run()
	go debugIO(a)

	paused := false
	running := true
	for running {
		select {
			case debugChar := <- a.DebugChannel:
				if a.IsPaused() == false {
					switch debugChar {
					case '`': // ` pauses the CPU
						a.SendCommand(izapple2.CommandCPUTraceOn)
						a.SendCommand(izapple2.CommandStep)
						continue
					}
				}

				switch debugChar {
				case 10: // Return steps the instruction
					if a.IsStepping() && !kp.modeSetValue {
						a.SendCommand(izapple2.CommandNextStep)
					} else if kp.modeMemory {
						kp.modeSetValue = false
						address := kp.modeValue & 0x0FFFFF0
						fmt.Printf("0x%06x :: ", address)
						for i := 0; i < 16; i++ {
							fmt.Printf("%02x ", a.Peek(address + uint32(i)))
						}
						fmt.Printf("  ")
						for i := 0; i < 16; i++ {
							ch := a.Peek(address + uint32(i))
							if (ch == 0) { ch = '.' } else if (ch < 32) { ch += 32 }
							fmt.Printf("%c", ch)
						}
						fmt.Printf("\n")
						kp.modeValue += 16
					} else if kp.modeBreakpoint {
						kp.modeSetValue = false
						kp.modeBreakpoint = false
						fmt.Printf("0x%x :: \n", kp.modeValue)
						a.SetUntilPC(kp.modeValue)
					}
				case '0','1','2','3','4','5','6','7','8','9':
					if kp.modeMemory || kp.modeBreakpoint {
						kp.modeValue = (kp.modeValue * 16) + uint32((debugChar - '0'))
					}
				case 'A','B','C','D','E','F':
					if kp.modeMemory || kp.modeBreakpoint {
						kp.modeValue = (kp.modeValue * 16) + uint32((debugChar - 'A' + 10))
					}
				case 'a','b','c','d','e','f':
					if kp.modeMemory || kp.modeBreakpoint {
						kp.modeValue = (kp.modeValue * 16) + uint32((debugChar - 'a' + 10))
					}
				case 'M','m':
					fmt.Printf("*** MEMORY: \n")
					kp.modeMemory = true
					kp.modeBreakpoint = false
					kp.modeSetValue = true
					kp.modeValue = 0
				case 'P','p':
					fmt.Printf("*** Run to PC: ")
					kp.modeMemory = false
					kp.modeBreakpoint = true
					kp.modeSetValue = true
					kp.modeValue = 0
				case '.':
					fmt.Printf("*** Run to PC (again)\n", )
					a.SetUntilPC(0xffffff)
					kp.modeSetValue = false
				case 'R', 'r', sdl.K_ESCAPE:
					kp.modeSetValue = false
					a.SendCommand(izapple2.CommandStart)
					a.SendCommand(izapple2.CommandCPUTraceOff)
				}
			default:
		       // no character to process (but if you don't have the default: then the channel is ignored?!)
		}

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				a.SendCommand(izapple2.CommandKill)
				running = false
			case *sdl.KeyboardEvent:
				kp.putKey(t)
				j.putKey(t)
			case *sdl.TextInputEvent:
				kp.putText(t.GetText())
			case *sdl.JoyAxisEvent:
				j.putAxisEvent(t)
			case *sdl.JoyButtonEvent:
				j.putButtonEvent(t)
			case *sdl.MouseMotionEvent:
				w, h := window.GetSize()
				j.putMouseMotionEvent(t, w, h)
				m.putMouseMotionEvent(t, w, h)
			case *sdl.MouseButtonEvent:
				j.putMouseButtonEvent(t)
				m.putMouseButtonEvent(t)
			}
		}

		if paused != a.IsPaused() {
			if a.IsPaused() {
				window.SetTitle("iz-" + a.Name + " - PAUSED!")
			} else {
				window.SetTitle("iz-" + a.Name)
			}
			paused = a.IsPaused()
		}

		if !a.IsPaused() {
			var img *image.RGBA
			if kp.showCharGen {
				img = screen.SnapshotCharacterGenerator(a, kp.showAltText)
				window.SetTitle(fmt.Sprintf("%v character map", a.Name))
			} else if kp.showPages {
				img = screen.SnapshotParts(a, kp.screenMode)
				window.SetTitle(fmt.Sprintf("%v - TEXT 1/2 - LORES 2/HIRES 1 - %vx%v", a.Name, img.Rect.Dx()/2, img.Rect.Dy()/2))
			} else {
				img = screen.Snapshot(a, kp.screenMode)
				window.SetTitle(a.Name)
			}
			if img != nil {
				surface, err := sdl.CreateRGBSurfaceFrom(unsafe.Pointer(&img.Pix[0]),
					int32(img.Bounds().Dx()), int32(img.Bounds().Dy()),
					32, 4*img.Bounds().Dx(),
					0x0000ff, 0x0000ff00, 0x00ff0000, 0xff000000)
				// Valid for little endian. Should we reverse for big endian?
				// 0xff000000, 0x00ff0000, 0x0000ff00, 0x000000ff)

				if err != nil {
					panic(err)
				}

				texture, err := renderer.CreateTextureFromSurface(surface)
				if err != nil {
					panic(err)
				}

				renderer.Clear()
				renderer.Copy(texture, nil, nil)
				renderer.Present()
				surface.Free()
				texture.Destroy()
			}
		}
		sdl.Delay(1000 / 30)
	}
}

// Run starts the Apple2 emulation
func debugIO(a *izapple2.Apple2) {
    var b []byte = make([]byte, 1)

	for {
        os.Stdin.Read(b)
		a.DebugChannel <- b[0]
	}

}

