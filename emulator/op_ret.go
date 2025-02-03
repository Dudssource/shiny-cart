package emulator

import "log"

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#RET
func op_ret(c *Cpu, _ uint8) {
	// m-cycles = 4
	c.requiredCycles = 4

	lsb := c.memory.Read(c.sp)
	c.sp++
	msb := c.memory.Read(c.sp)
	c.sp++
	c.pc = NewWord(msb, lsb)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#RETI
func op_reti(c *Cpu, opcode uint8) {
	// ret
	op_ret(c, opcode)

	// enable interrupts
	c.ime = 1

	log.Println("DISABLE RETI")
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#RET_cc
func op_ret_cond(c *Cpu, opcode uint8) {

	// my-cycles = 2
	c.requiredCycles = 2

	match := eval(c.reg.r_flags(), opcode)

	if match {

		c.requiredCycles = 5

		lsb := c.memory.Read(c.sp)
		c.sp++
		msb := c.memory.Read(c.sp)
		c.sp++

		c.pc = NewWord(msb, lsb)
	}
}
