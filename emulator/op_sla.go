package emulator

// https://rgbds.gbdev.io/docs/v0.8.0/gbz80.7#SLA_r8
// https://rgbds.gbdev.io/docs/v0.8.0/gbz80.7#SLA__HL_
func op_sla_r8(c *Cpu, opcode uint8) {
	c.requiredCycles = 2
	flags := c.reg.r_flags()
	flags &= ^(n_flag | h_flag | z_flag | c_flag)

	r8 := opcode & 0x7
	var nn uint8
	if r8 == reg_indirect_hl {
		c.requiredCycles = 4
		hl := c.reg.r16(reg_hl)
		nn = c.memory.Read(hl)
	} else {
		nn = c.reg.r8(r8)
	}

	result := uint16(nn << 1)

	if result > 255 {
		flags |= c_flag
		result = 256 - result
	} else if result == 0 {
		flags |= z_flag
	}

	if r8 == reg_indirect_hl {
		hl := c.reg.r16(reg_hl)
		c.memory.Write(hl, uint8(result))
	} else {
		c.reg.w8(r8, uint8(result))
	}

	c.reg.w_flag(flags)
}
