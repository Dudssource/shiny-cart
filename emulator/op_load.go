package emulator

func ld_b_b(c *Cpu) {
	c.bc = SetHigh(c.bc, c.bc.High())
}

func ld_b_c(c *Cpu) {
	c.bc = SetHigh(c.bc, c.bc.Low())
}

func ld_b_d(c *Cpu) {
	c.bc = SetHigh(c.bc, c.de.High())
}

func ld_r8_r8(c *Cpu, opcode uint8) {
	// 0b00111000
	dst := opcode & 0x38
	// 0b00000111
	src := opcode & 0x7
	// LD dst, src
	c.reg.w8(dst, c.reg.r8(src))
}
