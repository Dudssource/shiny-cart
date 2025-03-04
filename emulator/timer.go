package emulator

import "log"

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
	tmaValue uint8
}

func NewTimer(c *Cpu) *Timer {
	return &Timer{
		c: c,
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
				t.tmaValue = t.c.memory.Read(PORT_TMA)
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
