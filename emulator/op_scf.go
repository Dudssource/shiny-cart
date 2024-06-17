package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#SCF
func op_scf(c *Cpu, _ uint8) {

	// m-cycles = 1
	c.requiredCycles = 1

	// flags
	flags := c.reg.r_flags()

	// set carry=on
	flags |= c_flag

	// set half-carry=off
	// set subtraction=off
	flags &= ^h_flag & ^n_flag

	// save flags
	c.reg.w_flag(flags)
}
