package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#DEC_r16
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#DEC_SP
func op_dec_r16(c *Cpu, opcode uint8) {

	// m-cycles = 2
	c.requiredCycles = 2

	// 0b00110000
	dst := (opcode & 0x30) >> 4

	// DEC r16
	c.reg.w16(dst, c.reg.r16(dst)-1)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#DEC_r8
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#DEC__HL_
func op_dec_r8(c *Cpu, opcode uint8) {

	// 0b00111000
	dst := (opcode & 0x38) >> 3

	// flags
	flags := c.reg.r_flags()

	// subtraction
	flags |= n_flag

	// disable z_flag by default
	// disable h_flag by default
	flags &= (^z_flag & ^h_flag)

	// DEC [HL]
	if dst == reg_indirect_hl {

		// m-cycles = 3
		c.requiredCycles = 3

		// HL
		hl := c.reg.r16(reg_hl)

		// mem[HL]
		ihl := c.memory.Read(hl)

		// result
		result := int8(ihl - 1)

		// set zero flag
		if result == 0 {
			flags |= z_flag
		}

		// set half-carry flag
		if (ihl&0xF)-1 < 0 {
			flags |= h_flag
		}

		// mem[HL]--
		c.memory.Write(hl, uint8(result))

	} else {

		// m-cycles = 1
		c.requiredCycles = 1

		// R8
		r8 := c.reg.r8(dst)

		// DEC r8
		result := r8 - 1

		// zero-flag=on
		if result == 0 {
			flags |= z_flag
		}

		// half-carry=on
		if (r8&0xF)-1 < 0 {
			flags |= h_flag
		}

		// R8--
		c.reg.w8(dst, result)
	}

	// save flags
	c.reg.w_flag(flags)
}
