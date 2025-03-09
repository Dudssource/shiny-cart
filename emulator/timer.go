package emulator

import (
	"log"
	"time"
)

const (
	PORT_DIV  = Word(0xFF04)
	PORT_TIMA = Word(0xFF05)
	PORT_TMA  = Word(0xFF06)
	PORT_TAC  = Word(0xFF07)
)

var (
	// https://gbdev.io/pandocs/Timer_and_Divider_Registers.html#ff07--tac-timer-control
	timaClock = map[uint8]int{
		0x0: 256,
		0x1: 4,
		0x2: 16,
		0x3: 64,
	}
)

type Timer struct {
	c        *Cpu
	overflow bool

	// timer v2
	counter       uint16
	lastCycle     uint8
	overflowDelay int

	time      time.Time
	profiling int
}

func NewTimer(c *Cpu) *Timer {
	return &Timer{
		c: c,
	}
}

func (t *Timer) sync2(_ int) {

	if t.c.memory.resetTimer {
		log.Printf("DIV RESET")
		// reset counter
		t.counter = 0
		// acknowledge
		t.c.memory.resetTimer = false
	}

	// overflow (div is incremented every t-cycle)
	if t.counter == 0xFFFF {
		//log.Printf("DIV OVERFLOW")
		t.counter = 0x0
	} else {
		t.counter++
	}

	// only the upper 8 bit are mapped to memory
	// (must use memory.mem directly to bypass write, because writing to 0xFF04 would reset div)
	t.c.memory.mem[PORT_DIV] = uint8((t.counter & 0xFF00) >> 8)

	/* TIMA REGISTER */
	tac := t.c.memory.Read(PORT_TAC)
	timaFrequency := tac & 0x3

	incr := uint16(0)

	switch timaFrequency {
	case 0x00:
		incr = (t.counter & 0x200) >> 9
	case 0x01:
		incr = (t.counter & 0x8) >> 3
	case 0x02:
		incr = (t.counter & 0x20) >> 5
	case 0x03:
		incr = (t.counter & 0x80) >> 7
	}

	currentTimaValue := int(t.c.memory.Read(PORT_TIMA))
	andResult := ((tac & 0x4) >> 2) & uint8(incr)

	if !t.overflow && t.lastCycle == 1 && andResult == 0 {

		if t.time.IsZero() || time.Since(t.time) >= time.Second {
			//log.Printf("TIMA %d (%d) Hz", t.profiling, timaFrequency)
			t.profiling = 0
			t.time = time.Now()
		} else {
			t.profiling++
		}

		// overflow
		if currentTimaValue == 0xFF {
			//log.Printf("TIMA OVERFLOW")
			// reset TIMA
			t.c.memory.Write(PORT_TIMA, 0)
			currentTimaValue = 0
			t.overflow = true
			t.overflowDelay = 4
		} else {
			// increment TIMA
			t.c.memory.Write(PORT_TIMA, uint8(currentTimaValue+1))
		}
	}

	// store result
	t.lastCycle = andResult

	// TIMA enabled
	if (tac & 0x4) > 0 {

		if t.overflow && t.overflowDelay == 0 {
			tma := t.c.memory.Read(PORT_TMA)
			iflag := t.c.memory.Read(INTERRUPT_FLAG)

			// if t.c.debug {
			//log.Printf("TIMA=%d, CYCLE %d, FREQ=%d, TMA=%d, IFLAG=%.8b\n", currentTimaValue, cycle, timaClock[timaFrequency], tma, iflag)
			// }

			// set TIMA as TMA (Modulo)
			t.c.memory.Write(PORT_TIMA, tma)
			// request TIMER interruption
			t.c.memory.Write(INTERRUPT_FLAG, iflag|0x4)
			// overflow processed
			t.overflow = false
		} else if t.overflow && t.overflowDelay > 0 {
			// TIMA was written during delay, abort reload and interruption
			if currentTimaValue > 0 {
				log.Printf("TIMA ABORTED")
				t.overflow = false
				t.overflowDelay = 0
			} else {
				t.overflowDelay--
			}
		}
	}
}

func (t *Timer) sync(cycle int) {

	/* DIV REGISTER */
	currentDivValue := int(t.c.memory.Read(PORT_DIV))

	if cycle%64 == 0 { // 16384 Hz
		// overflow
		if currentDivValue+1 > 0xFF {
			// Reset DIV
			t.c.memory.mem[PORT_DIV] = 0

		} else {
			currentDivValue++
			// increment DIV
			t.c.memory.mem[PORT_DIV] = uint8(currentDivValue)
		}
	}

	/* TIMA REGISTER */
	tac := t.c.memory.Read(PORT_TAC)
	timaEnabled := (tac & 0x4) > 0

	if timaEnabled {

		timaFrequency := timaClock[tac&0x3]
		currentTimaValue := int(t.c.memory.Read(PORT_TIMA))

		if t.overflow {
			tma := t.c.memory.Read(PORT_TMA)
			iflag := t.c.memory.Read(INTERRUPT_FLAG)

			if t.c.debug {
				log.Printf("TIMA=%d, CYCLE %d, FREQ=%d, TMA=%d, IFLAG=%.8b\n", currentTimaValue, cycle, timaFrequency, tma, iflag)
			}

			// set TIMA as TMA (Modulo)
			t.c.memory.Write(PORT_TIMA, tma)
			// request TIMER interruption
			t.c.memory.Write(INTERRUPT_FLAG, iflag|0x4)
			// overflow processed
			t.overflow = false
		}

		if cycle%timaFrequency == 0 {
			//log.Printf("CYCLE %d %d\n", cycle, timaFrequency)

			// overflow
			if currentTimaValue == 0xFF {
				// reset TIMA
				t.c.memory.Write(PORT_TIMA, 0)
				t.overflow = true
			} else {
				// increment TIMA
				t.c.memory.Write(PORT_TIMA, uint8(currentTimaValue+1))
			}
		}
	}
}

func (t *Timer) init() {

}
