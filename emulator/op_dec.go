package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#DEC_r16
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#DEC_SP
func op_dec_r16(c *Cpu, opcode uint8) {

	// m-cycles = 2
	c.requiredCycles = 2

	// 0b00110000
	dst := (opcode & 0x30) >> 4

	// DEC r16
	c.reg.w16(dst, c.reg.r16(dst)-1)
}
