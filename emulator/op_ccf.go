package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#CCF
func op_ccf(c *Cpu, _ uint8) {

	// m-cycles = 1
	c.requiredCycles = 1

	// read flags
	flags := c.reg.r_flags()

	// read c-flag flipping with XOR
	flags = (flags ^ c_flag)

	// save flags
	c.reg.w_flag(flags)
}
