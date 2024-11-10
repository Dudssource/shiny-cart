package emulator

// https://rgbds.gbdev.io/docs/v0.8.0/gbz80.7#RRC_r8
// https://rgbds.gbdev.io/docs/v0.8.0/gbz80.7#RRC__HL_
func op_rrc_r8(c *Cpu, opcode uint8) {
	c.requiredCycles = 2
	flags := c.reg.r_flags()
	flags &= ^(z_flag | h_flag | n_flag | c_flag)
	dst := opcode & 0x3

	var nn uint8

	if dst == reg_indirect_hl {

		c.requiredCycles = 4
		hl := c.reg.r16(reg_hl)
		nn = c.memory.Read(hl)

	} else {
		nn = c.reg.r8(dst)
	}

	rm := (nn & 0x1) << 7

	result := (nn >> 1) | rm
	if rm > 0 {
		flags |= c_flag
	}

	if result == 0 {
		flags |= z_flag
	}

	if dst == reg_indirect_hl {
		hl := c.reg.r16(reg_hl)
		c.memory.Write(hl, uint8(result))
	} else {
		c.reg.w8(dst, uint8(result))
	}

	c.reg.w_flag(flags)
}
