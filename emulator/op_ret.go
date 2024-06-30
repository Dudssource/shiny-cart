package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#RET
func op_ret(c *Cpu, _ uint8) {
	// m-cycles = 4
	c.requiredCycles = 4

	c.sp--
	lsb := c.stack[c.sp]
	c.sp--
	msb := c.stack[c.sp]
	c.pc = NewWord(msb, lsb)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#RETI
func op_reti(c *Cpu, opcode uint8) {
	// ret
	op_ret(c, opcode)

	// enable interrupts
	c.ime = 1
}
