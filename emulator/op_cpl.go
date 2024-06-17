package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#CPL
func op_cpl(c *Cpu, _ uint8) {

	// m-cycles = 1
	c.requiredCycles = 1

	// A = ~A
	c.reg.w8(reg_a, ^c.reg.r8(reg_a))

	// read flags
	flags := c.reg.r_flags()

	// set n and h to 1
	flags |= n_flag | h_flag

	// save flags
	c.reg.w_flag(flags)
}
