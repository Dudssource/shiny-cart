package emulator

// https://rgbds.gbdev.io/docs/v0.8.0/gbz80.7#PUSH_r16
// https://rgbds.gbdev.io/docs/v0.8.0/gbz80.7#PUSH_AF
func op_push_r16stk(c *Cpu, opcode uint8) {

	c.requiredCycles = 4

	dst := (opcode & 0x30) >> 4
	var reg Word

	switch dst {
	case 0x0:
		reg = c.reg.r16(reg_bc)
	case 0x1:
		reg = c.reg.r16(reg_de)
	case 0x2:
		reg = c.reg.r16(reg_hl)
	case 0x3:
		flags := c.reg.r_flags()
		f := uint8(flags) & 0xF0
		reg = NewWord(c.reg.r8(reg_a), f)
	}

	c.sp--
	c.memory.Write(c.sp, reg.High())
	c.sp--
	c.memory.Write(c.sp, reg.Low())
}

// https://rgbds.gbdev.io/docs/v0.8.0/gbz80.7#POP_AF
// https://rgbds.gbdev.io/docs/v0.8.0/gbz80.7#POP_r16
func op_pop_r16stk(c *Cpu, opcode uint8) {

	c.requiredCycles = 3

	lsb := c.memory.Read(c.sp)
	c.sp++
	msb := c.memory.Read(c.sp)
	c.sp++

	dst := (opcode & 0x30) >> 4

	switch dst {
	case 0x0:
		c.reg.w16(reg_bc, NewWord(msb, lsb))
	case 0x1:
		c.reg.w16(reg_de, NewWord(msb, lsb))
	case 0x2:
		c.reg.w16(reg_hl, NewWord(msb, lsb))
	case 0x3:
		flags := c.reg.r_flags()
		c.reg.w8(reg_f, uint8(flags)|(lsb&0xF0))
		c.reg.w8(reg_a, msb)
	}
}
