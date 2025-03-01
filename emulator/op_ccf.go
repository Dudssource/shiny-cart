package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#CCF
func op_ccf(c *Cpu, _ uint8) {

	// m-cycles = 1
	c.requiredCycles = 1

	// read flags
	flags := c.reg.r_flags()

	// h-flag=off
	// n-flag=off
	flags &= ^(h_flag | n_flag)

	carry := (flags & c_flag) >> 4
	if carry == 1 {
		flags &= ^c_flag
	} else {
		flags |= c_flag
	}

	// save flags
	c.reg.w_flag(flags)
}
