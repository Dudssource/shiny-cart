package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#DAA
func op_daa(c *Cpu, _ uint8) {

	// m-cycles = 1
	c.requiredCycles = 1

	// A
	a := c.reg.r8(reg_a)

	// high nibble from a
	hi_a := a & 0xF0 >> 4

	// low nibble from a
	lo_a := a & 0xF

	// flags
	flags := c.reg.r_flags()

	// carry flag
	carry := (flags & c_flag) > 0

	// n flag
	nf := (flags & n_flag) > 0

	// half-carry flag
	halfcarry := (flags & h_flag) > 0

	// http://www.z80.info/zip/z80-documented.pdf
	// http://z80-heaven.wikidot.com/instructions-set:daa
	// https://www.smspower.org/Development/BinaryCodedDecimal
	diff := uint8(0x00)

	if !carry {

		// CF=0 AND hi_a 0-9 AND HF=1 AND lo_i 0-9
		if hi_a <= 9 && lo_a <= 9 {

			// turn c_flag off
			flags &= ^c_flag

			if halfcarry {
				diff = 0x06
			}
		}

		// CF=0 AND hi_a 0-8 AND lo_i a-f
		if hi_a <= 8 && (lo_a >= 0xA && lo_a <= 0xF) {
			// turn c_flag off
			flags &= ^c_flag
			diff = 0x06
		}

		// CF=0 AND hi_a a-f AND H=0 AND lo_i 0-9
		if (hi_a >= 0xA && hi_a <= 0xF) && lo_a <= 9 {

			// turn c_flag on
			flags |= c_flag

			if !halfcarry {
				diff = 0x60
			} else {
				diff = 0x66
			}
		}

		// CF=0 AND hi_a 9-f AND lo_i a-f
		if (hi_a >= 0x9 && hi_a <= 0xF) && (lo_a >= 0xA && lo_a <= 0xF) {
			// turn c_flag on
			flags |= c_flag
			diff = 0x66
		}
	}

	if carry {

		// turn c_flag on
		flags |= c_flag

		// CF=1 AND H=0 AND lo_i 0-9
		if lo_a <= 9 {
			if !halfcarry {
				diff = 0x60
			} else {
				diff = 0x66
			}
		}

		// CF=1 AND lo_i a-f
		if lo_a >= 0xA && lo_a <= 0xF {
			diff = 0x66
		}
	}

	if nf {
		a -= diff
	} else {
		a += diff
	}

	// turn z-flag on
	if a == 0 {
		flags |= z_flag
	}

	// turn half-carry flag off
	flags &= ^h_flag

	// SAVE A
	c.reg.w8(reg_a, a)

	// SAVE flags
	c.reg.w_flag(flags)
}
