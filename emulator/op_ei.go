package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#EI
func op_ei(c *Cpu, _ uint8) {
	// m-cycles = 1
	c.requiredCycles = 1
	c.ime = 1
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#DI
func op_di(c *Cpu, _ uint8) {
	// m-cycles = 1
	c.requiredCycles = 1
	c.ime = 0
}
