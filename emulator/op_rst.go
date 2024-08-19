package emulator

// https://rgbds.gbdev.io/docs/v0.8.0/gbz80.7#RST_vec
func op_rst(c *Cpu, opcode uint8) {
	c.requiredCycles = 4

	tgt := (opcode & 0x38) >> 3
	c.sp--
	pc := c.pc
	c.memory.Write(c.sp, pc.High())
	c.sp--
	c.memory.Write(c.sp, pc.Low())
	c.pc = Word(tgt)
}
