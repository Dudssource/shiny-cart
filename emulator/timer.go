package emulator

import (
	"log"
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
	c            *Cpu
	lastDivValue int
}

func NewTimer(c *Cpu) *Timer {
	return &Timer{
		c: c,
	}
}

func (t *Timer) sync(cycle int) {

	/* DIV REGISTER */
	currentDivValue := int(t.c.memory.Read(PORT_DIV))
	// some value was written to DIV or we had an overflow
	if t.lastDivValue > 0 && t.lastDivValue != currentDivValue {
		t.lastDivValue = 0
		// Reset DIV
		t.c.memory.Write(PORT_DIV, 0)
	} else if cycle%64 == 0 { // 16384 Hz
		// overflow
		if currentDivValue+1 > 0xFF {
			// Reset DIV
			t.c.memory.Write(PORT_DIV, 0)
		} else {
			currentDivValue++
			// increment DIV
			t.c.memory.Write(PORT_DIV, uint8(currentDivValue))
			t.lastDivValue = currentDivValue
		}
	}

	/* TIMA REGISTER */
	tac := t.c.memory.Read(PORT_TAC)
	timaEnabled := (tac & 0x4) > 0

	if timaEnabled {
		timaFrequency := tac & 0x3
		if cycle%timaClock[timaFrequency] == 0 {
			currentTimaValue := int(t.c.memory.Read(PORT_TIMA))
			// overflow
			if currentTimaValue+1 > 0xFF {
				tma := t.c.memory.Read(PORT_TMA)
				iflag := t.c.memory.Read(INTERRUPT_FLAG) | 0x4
				log.Printf("TIMA=%d, CYCLE %d, FREQ=%d, TMA=%d, IFLAG=%.8b\n", currentTimaValue, cycle, timaClock[timaFrequency], tma, iflag)
				// set TIMA as TMA (Modulo)
				t.c.memory.Write(PORT_TIMA, tma)
				// request TIMER interruption
				t.c.memory.Write(INTERRUPT_FLAG, iflag)
			} else {
				// increment TIMA
				t.c.memory.Write(PORT_TIMA, uint8(currentTimaValue+1))
			}
		}
	}
}

func (t *Timer) init() {

}
