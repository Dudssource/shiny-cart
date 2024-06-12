package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#ADD_HL,r16
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#ADD_HL,SP
func op_add_hl_r16(c *Cpu, opcode uint8) {

	// m-cycles = 2
	c.requiredCycles = 2

	// read flags
	flags := c.reg.r_flags()

	// no subtraction
	// disable carry flag as default
	// disable half-carry flag as default
	flags &= ^n_flag & ^c_flag & ^h_flag

	// 0b00110000
	dst := (opcode & 0x30) >> 4

	// nn = r16
	nn := c.reg.r16(dst)

	// HL
	hl := c.reg.r16(reg_hl)

	// HL + NN
	result := uint32(hl + nn)

	// if result is greater than 0xFFF, set half-carry=on (bit 11)
	if (hl&0xFFF + nn&0xFFF) > 0xFFF {
		flags |= h_flag
	}

	// if result is greater than 0xFFFF, set carry=on (bit 15)
	if result > 0xFFFF {
		flags |= c_flag
		result = result - 0xFFFF
	}

	// set HL
	c.reg.w16(reg_h, Word(result))

	// save flags
	c.reg.w_flag(flags)
}
