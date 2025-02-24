package emulator

// https://gbdev.io/pandocs/halt.html#halt
// https://rgbds.gbdev.io/docs/v0.9.0/gbz80.7#HALT
func op_halt(c *Cpu, _ uint8) {
	c.step = true
}
