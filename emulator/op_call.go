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
	nn_msb := pc.High()
	nn_lsb := pc.Low()
	c.stack[c.sp] = nn_msb
	c.sp++
	c.stack[c.sp] = nn_lsb
	c.sp++
}
