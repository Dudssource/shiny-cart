package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#INC_r16
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#INC_SP
func op_inc_r16(c *Cpu, opcode uint8) {

	// m-cycles = 2
	c.requiredCycles = 2

	// 0b00110000
	dst := (opcode & 0x30) >> 4

	if dst == reg_sp {
		// INC SP
		c.sp++
	} else {
		// INC r16
		c.reg.w16(dst, c.reg.r16(dst)+1)
	}
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#INC_r8
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#INC__HL_
func op_inc_r8(c *Cpu, opcode uint8) {

	// flags
	flags := c.reg.r_flags()

	// no subtraction
	// disable zero flag as default
	// disable half-carry flag as default
	flags &= ^n_flag & ^z_flag & ^h_flag

	// 0b00111000
	dst := (opcode & 0x38) >> 3

	// INC [HL]
	if dst == reg_indirect_hl {

		// m-cycles = 3
		c.requiredCycles = 3

		// HL
		hl := c.reg.r16(reg_hl)

		// mem[HL]
		ihl := c.memory.Read(hl)

		// INC [HL]
		result := ihl + 1

		// set z-flag=on
		if result == 0 {
			flags |= z_flag
		}

		// set half-carry=on
		if (ihl&0xF + 1) > 0xF {
			flags |= h_flag
		}

		// mem[HL]+1
		c.memory.Write(hl, result)

	} else {

		// m-cycles = 1
		c.requiredCycles = 1

		// r8
		r8 := c.reg.r8(dst)

		// INC r8
		result := r8 + 1

		// set z-flag=on
		if result == 0 {
			flags |= z_flag
		}

		// set half-carry=on
		if (r8&0xF + 1) > 0xF {
			flags |= h_flag
		}

		// R8++
		c.reg.w8(dst, result)
	}

	// save flags
	c.reg.w_flag(flags)
}
