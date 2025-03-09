package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#JP_n16
func op_jp_imm16(c *Cpu, opcode uint8) {

	// m-cycles = 4
	c.requiredCycles = 4

	lsb := c.fetch()
	msb := c.fetch()
	c.previousPC = c.pc
	c.pc = NewWord(msb, lsb)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#JP_HL
func op_jp_hl(c *Cpu, opcode uint8) {

	// m-cycles = 1
	c.requiredCycles = 1

	// read hl into pc
	c.previousPC = c.pc
	c.pc = c.reg.r16(reg_hl)
}

// https://rgbds.gbdev.io/docs/v0.8.0/gbz80.7#JP_cc,n16
func op_jp_cond(c *Cpu, opcode uint8) {
	c.requiredCycles = 3
	nn_lsb := c.fetch()
	nn_msb := c.fetch()
	nn := NewWord(nn_msb, nn_lsb)
	match := eval(c.reg.r_flags(), opcode)
	if match {
		c.requiredCycles = 4
		c.previousPC = c.pc
		c.pc = nn
	}
}
