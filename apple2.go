package izapple2

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/lunarmobiscuit/iz6502"
)

// Apple2 represents all the components and state of the emulated machine
type Apple2 struct {
	Name                string
	cpu                 *iz6502.State
	mmu                 *memoryManager
	io                  *ioC0Page
	cg                  *CharacterGenerator
	cards               [8]Card
	softVideoSwitch     *SoftVideoSwitch
	isApple2e           bool
	commandChannel      chan int
	cycleDurationNs     float64 // Current speed. Inverse of the cpu clock in Ghz
	fastMode            bool
	fastRequestsCounter int32
	cycleBreakpoint     uint64
	breakPoint          bool
	profile             bool
	showSpeed           bool
	paused              bool
	stepping            bool
	nextStep            bool
	stopAtPC			bool
	stopAtAddr			uint32
	tracers             []executionTracer
	forceCaps           bool
}

type executionTracer interface {
	inspect()
}

const (
	// CPUClockMhz is the actual Apple II clock speed
	CPUClockMhz     = 4 // Mhz instead of  14.318 / 14
	cpuClockEuroMhz = 14.238 / 14
)

const (
	maxWaitDuration = 100 * time.Millisecond
	cpuSpinLoops    = 100
)

// Run starts the Apple2 emulation
func (a *Apple2) Run() {
	a.Start(false)
}

// Start the Apple2 emulation, can start paused
func (a *Apple2) Start(paused bool) {
	// Start the processor
	a.cpu.Reset()
	referenceTime := time.Now()
	speedReferenceTime := referenceTime
	speedReferenceCycles := uint64(0)

	a.paused = paused

	for {
		// Run 6502 steps
		if !a.paused || (a.stepping && (a.nextStep || a.stopAtPC)) {
			for i := 0; i < cpuSpinLoops; i++ {
				// Conditional tracing
				//a.cpu.SetTrace((pc >= 0xc300 && pc <= 0xc400) || (pc >= 0xc800 && pc <= 0xce00))

				// Execution
				a.cpu.ExecuteInstruction()

				// Special tracing
				a.executionTrace()

				// Waiting on PC or stepping instructions
				if a.stopAtPC {
					pc := a.cpu.GetPC()
					if (pc == a.stopAtAddr) {
						a.stopAtPC = false
						a.cpu.SetTrace(true)
						break
					}
				} else if (a.stepping) {
					a.nextStep = false
					break
				}
			}

			if a.cycleBreakpoint != 0 && a.cpu.GetCycles() >= a.cycleBreakpoint {
				a.breakPoint = true
				a.cycleBreakpoint = 0
				a.paused = true
			}
		} else {
			time.Sleep(200 * time.Millisecond)
		}

		// Execute meta commands
		commandsPending := true
		for commandsPending {
			select {
			case command := <-a.commandChannel:
				switch command {
				case CommandKill:
					return
				case CommandPause:
					if !a.paused {
						a.paused = true
						a.stopAtPC = false
					}
				case CommandStep:
					if !a.stepping {
						a.paused = true
						a.stepping = true
						a.nextStep = true
						a.stopAtPC = false
					}
				case CommandNextStep:
					if a.stepping {
						a.nextStep = true
						a.stopAtPC = false
					}
				case CommandStart:
					if a.paused {
						a.paused = false
						a.stepping = false
						a.nextStep = false
						a.stopAtPC = false
						referenceTime = time.Now()
						speedReferenceTime = referenceTime
					}
				case CommandPauseUnpause:
					a.paused = !a.paused
					referenceTime = time.Now()
					speedReferenceTime = referenceTime
				case CommandStepNostep:
					a.stepping = !a.stepping
				default:
					// Execute the other commands
					a.executeCommand(command)
				}
			default:
				commandsPending = false
			}
		}

		if a.cycleDurationNs != 0 && a.fastRequestsCounter <= 0 {
			// Wait until next 6502 step has to run
			clockDuration := time.Since(referenceTime)
			simulatedDuration := time.Duration(float64(a.cpu.GetCycles()) * a.cycleDurationNs)
			waitDuration := simulatedDuration - clockDuration
			if waitDuration > maxWaitDuration || -waitDuration > maxWaitDuration {
				// We have to wait too long or are too much behind. Let's fast forward
				referenceTime = referenceTime.Add(-waitDuration)
				waitDuration = 0
			}
			if waitDuration > 0 {
				time.Sleep(waitDuration)
			}
		}

		if a.showSpeed && a.cpu.GetCycles()-speedReferenceCycles > 1000000 {
			// Calculate speed in MHz every million cycles
			newTime := time.Now()
			newCycles := a.cpu.GetCycles()
			elapsedCycles := float64(newCycles - speedReferenceCycles)
			freq := 1000.0 * elapsedCycles / float64(newTime.Sub(speedReferenceTime).Nanoseconds())
			fmt.Printf("Freq: %f Mhz\n", freq)
			speedReferenceTime = newTime
			speedReferenceCycles = newCycles
		}
	}
}

func (a *Apple2) reset() {
	a.cpu.Reset()
	a.mmu.reset()
	for _, c := range a.cards {
		if c != nil {
			c.reset()
		}
	}
}

// IsPaused returns true when emulator is paused
func (a *Apple2) IsPaused() bool {
	return a.paused
}

// IsStepping returns true when emulator is stepping cycle by cycle
func (a *Apple2) IsStepping() bool {
	return a.stepping
}

// SetUntilPC runs until the emulator reaches the specific PC
func (a *Apple2) IsRunningUntilPC() bool {
	return a.stopAtPC
}

// SetUntilPC runs until the emulator reaches the specific PC
func (a *Apple2) SetUntilPC(pc uint32) {
	a.cpu.SetTrace(false)
	a.stepping = true
	a.stopAtPC = true
	a.stopAtAddr = pc
}

func (a *Apple2) GetCycles() uint64 {
	return a.cpu.GetCycles()
}

// SetCycleBreakpoint sets a cycle number to pause the emulator. 0 to disable
func (a *Apple2) SetCycleBreakpoint(cycle uint64) {
	a.cycleBreakpoint = cycle
	a.breakPoint = false
}

func (a *Apple2) BreakPoint() bool {
	return a.breakPoint
}

func (a *Apple2) setProfiling(value bool) {
	a.profile = value
}

// IsProfiling returns true when profiling
func (a *Apple2) IsProfiling() bool {
	return a.profile
}

// SetForceCaps allows the caps state to be toggled at runtime
func (a *Apple2) SetForceCaps(value bool) {
	a.forceCaps = value
}

// IsForceCaps returns true when all letters are forced to upper case
func (a *Apple2) IsForceCaps() bool {
	return a.forceCaps
}

const (
	// CommandToggleSpeed toggles cpu speed between full speed and actual Apple II speed
	CommandToggleSpeed = iota + 1
	// CommandShowSpeed toggles printinf the current freq in Mhz
	CommandShowSpeed
	// CommandDumpDebugInfo dumps useful info
	CommandDumpDebugInfo
	// CommandNextCharGenPage cycles the CharGen page if several
	CommandNextCharGenPage
	// CommandToggleCPUTrace toggle tracing of CPU execution
	CommandToggleCPUTrace
	// CommandCPUTraceOn enables tracing of CPU execution
	CommandCPUTraceOn
	// CommandCPUTraceOn disable tracing of CPU execution
	CommandCPUTraceOff
	// CommandKill stops the cpu execution loop
	CommandKill
	// CommandReset executes a 6502 reset
	CommandReset
	// CommandPauseUnpause allows the Pause button to freeze the emulator for a coffee break
	CommandPauseUnpause
	// CommandStepNostep puts the emulator into insruction by instruction stepping
	CommandStepNostep
	// CommandPause pauses the emulator
	CommandPause
	// CommandStart restarts the emulator
	CommandStart
	// CommandStep enters step by step execution mode
	CommandStep
	// CommandNextStep executes the next instruction (in step by step mode)
	CommandNextStep
)

// SendCommand enqueues a command to the emulator thread
func (a *Apple2) SendCommand(command int) {
	a.commandChannel <- command
}

func (a *Apple2) executeCommand(command int) {
	switch command {
	case CommandToggleSpeed:
		if a.cycleDurationNs == 0 {
			//fmt.Println("Slow")
			a.cycleDurationNs = 1000.0 / CPUClockMhz
		} else {
			//fmt.Println("Fast")
			a.cycleDurationNs = 0
		}
	case CommandShowSpeed:
		a.showSpeed = !a.showSpeed
	case CommandDumpDebugInfo:
		a.dumpDebugInfo()
	case CommandNextCharGenPage:
		a.cg.nextPage()
		fmt.Printf("Chargen page %v\n", a.cg.page)
	case CommandToggleCPUTrace:
		a.cpu.SetTrace(!a.cpu.GetTrace())
	case CommandCPUTraceOn:
		a.cpu.SetTrace(true)
	case CommandCPUTraceOff:
		a.cpu.SetTrace(false)
	case CommandReset:
		a.reset()
	}
}

func (a *Apple2) RequestFastMode() {
	// Note: if the fastMode is shorter than maxWaitDuration, there won't be any gain.
	if a.fastMode {
		atomic.AddInt32(&a.fastRequestsCounter, 1)
	}
}

func (a *Apple2) ReleaseFastMode() {
	if a.fastMode {
		atomic.AddInt32(&a.fastRequestsCounter, -1)
	}
}

func (a *Apple2) executionTrace() {
	if a.tracers != nil {
		for _, v := range a.tracers {
			v.inspect()
		}
	}
}

func (a *Apple2) dumpDebugInfo() {
	// See "Apple II Monitors Peeled"
	pageZeroSymbols := map[int]string{
		0x36: "CSWL",
		0x37: "CSWH",
		0x38: "KSWL",
		0x39: "KSWH",
		0xe2: "ACJVAFLDL", // Apple Pascal
		0xe3: "ACJVAFLDH", // Apple Pascal
		0xec: "JVBFOLDL",  // Apple Pascal
		0xed: "JVBFOLDH",  // Apple Pascal
		0xee: "JVAFOLDL",  // Apple Pascal
		0xef: "JVAFOLDH",  // Apple Pascal
	}

	fmt.Printf("Page zero values:\n")
	for _, k := range []int{0x36, 0x37, 0x38, 0x39, 0xe2, 0xe3, 0xec, 0xed, 0xee, 0xef} {
		d := a.mmu.physicalMainRAM.data[k]
		fmt.Printf("  %v(0x%x): 0x%02x\n", pageZeroSymbols[k], k, d)
	}
}
