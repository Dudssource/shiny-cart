package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#NOP
func op_nop(c *Cpu, opcode uint8) {
	c.requiredCycles = 1
}
