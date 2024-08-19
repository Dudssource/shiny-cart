package emulator

// https://rgbds.gbdev.io/docs/v0.8.0/gbz80.7#RRC_r8
// https://rgbds.gbdev.io/docs/v0.8.0/gbz80.7#RRC__HL_
func op_rlc_r8(c *Cpu, opcode uint8) {
	c.requiredCycles = 2
	flags := c.reg.r_flags()
	flags |= ^z_flag
	dst := opcode & 0x3

	var nn uint8

	if dst == reg_indirect_hl {

		c.requiredCycles = 4
		hl := c.reg.r16(reg_hl)
		nn = c.memory.Read(hl)

	} else {
		nn = c.reg.r8(dst)
	}

	result := int(nn >> 1)
	if result < 0x0 {
		flags |= c_flag
		result = 256 + result
	}

	if dst == reg_indirect_hl {
		hl := c.reg.r16(reg_hl)
		c.memory.Write(hl, uint8(result))
	} else {
		c.reg.w8(dst, uint8(result))
	}

	c.reg.w_flag(flags)
}
