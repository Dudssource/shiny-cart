package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#CALL_n16
func op_call_imm16(c *Cpu, _ uint8) {

	// m-cycles = 6
	c.requiredCycles = 6

	// set pc to n16
	lsb := c.fetch()
	msb := c.fetch()
	pc := c.pc
	c.pc = NewWord(msb, lsb)

	// store pc into stack
	c.sp--
	c.memory.Write(c.sp, pc.High())
	c.sp--
	c.memory.Write(c.sp, pc.Low())
}

// https://rgbds.gbdev.io/docs/v0.8.0/gbz80.7#CALL_cc,n16
func op_call_cond(c *Cpu, opcode uint8) {
	c.requiredCycles = 3
	z := c.fetch()
	w := c.fetch()

	match := eval(c.reg.r_flags(), opcode)

	if match {
		c.requiredCycles = 6
		c.sp--
		c.memory.Write(c.sp, c.pc.High())
		c.sp--
		c.memory.Write(c.sp, c.pc.Low())
		c.pc = NewWord(w, z)
	}

}
