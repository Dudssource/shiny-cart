package emulator

// https://rgbds.gbdev.io/docs/v0.8.0/gbz80.7#RES_u3,r8
// https://rgbds.gbdev.io/docs/v0.8.0/gbz80.7#RES_u3,_HL_
func op_res_r8(c *Cpu, opcode uint8) {
	c.requiredCycles = 2

	r8 := opcode & 0x7
	var nn uint8
	if r8 == reg_indirect_hl {
		c.requiredCycles = 4
		hl := c.reg.r16(r8)
		nn = c.memory.Read(hl)
	} else {
		nn = c.reg.r8(r8)
	}

	b3 := (nn & 0x38) >> 3
	result := nn & ^(0x1 << b3)

	if r8 == reg_indirect_hl {
		hl := c.reg.r16(reg_hl)
		c.memory.Write(hl, uint8(result))
	} else {
		c.reg.w8(r8, uint8(result))
	}
}
