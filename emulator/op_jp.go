package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#JP_n16
func op_jp_imm16(c *Cpu, opcode uint8) {

	// m-cycles = 4
	c.requiredCycles = 4

	lsb := c.fetch()
	msb := c.fetch()
	c.pc = NewWord(msb, lsb)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#JP_HL
func op_jp_hl(c *Cpu, opcode uint8) {

	// m-cycles = 1
	c.requiredCycles = 1

	// read hl into pc
	c.pc = c.reg.r16(reg_hl)
}
