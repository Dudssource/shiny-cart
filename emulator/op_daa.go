package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#DAA
// https://forums.nesdev.org/viewtopic.php?t=9088
func op_daa(c *Cpu, _ uint8) {

	// m-cycles = 1
	c.requiredCycles = 1

	// A
	a := int(c.reg.r8(reg_a))

	// flags
	flags := c.reg.r_flags()

	// carry flag
	cf := (flags & c_flag) > 0

	// n flag
	nf := (flags & n_flag) > 0

	// half-carry flag
	hf := (flags & h_flag) > 0

	if !nf {
		if hf || (a&0xf) > 0x9 {
			a += 0x06
		}
		if cf || a > 0x9F {
			a += 0x60
		}
	} else {
		if hf {
			a = (a - 0x6) & 0xFF
		}
		if cf {
			a -= 0x60
		}
	}

	flags &= ^(h_flag | z_flag)

	if (a & 0x100) == 0x100 {
		flags |= c_flag
	}

	a &= 0xFF

	if a == 0 {
		flags |= z_flag
	}

	// SAVE A
	c.reg.w8(reg_a, uint8(a))

	// SAVE flags
	c.reg.w_flag(flags)
}
