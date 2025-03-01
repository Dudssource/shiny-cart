package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#DAA
func op_daa(c *Cpu, _ uint8) {

	// m-cycles = 1
	c.requiredCycles = 1

	// A
	a := c.reg.r8(reg_a)

	// flags
	flags := c.reg.r_flags()

	// carry flag
	cf := (flags & c_flag) > 0

	// n flag
	nf := (flags & n_flag) > 0

	// half-carry flag
	hf := (flags & h_flag) > 0

	// reset c-flag
	flags &= ^(c_flag | z_flag | h_flag)

	if !nf {
		if cf || a > 0x99 {
			a += 0x60
			flags |= c_flag
		}
		if hf || (a&0x0f) > 0x09 {
			a += 0x6
		}
	} else {
		if cf {
			a -= 0x60
		}
		if hf {
			a -= 0x6
		}
	}

	if a == 0 {
		flags |= z_flag
	}

	// SAVE A
	c.reg.w8(reg_a, a)

	// SAVE flags
	c.reg.w_flag(flags)
}
