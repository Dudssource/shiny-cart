package emulator

import "log"

// https://gbdev.io/pandocs/halt.html#halt
// https://rgbds.gbdev.io/docs/v0.9.0/gbz80.7#HALT
func op_halt(c *Cpu, _ uint8) {
	c.remainingCycles = 1
	interruptPending := (c.memory.Read(INTERRUPT_ENABLE) & c.memory.Read(INTERRUPT_FLAG)) > 0
	if c.debug {
		log.Printf("HALT requested : ime=%d, pending=%t\n", c.ime, interruptPending)
	}

	if c.ime == 0 && interruptPending { // ime not set and interrupt pending
		c.halted = false
		c.haltBug = true
		log.Printf("HALT BUG TIGGRERED\n")
	} else {
		c.halted = true
	}
}
