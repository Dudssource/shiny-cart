package emulator

// https://rgbds.gbdev.io/docs/v0.8.0/gbz80.7#RST_vec
func op_rst(c *Cpu, opcode uint8) {
	c.requiredCycles = 4

	vec := map[uint8]uint8{
		0b000: 0x00,
		0b001: 0x08,
		0b010: 0x10,
		0b011: 0x18,
		0b100: 0x20,
		0b101: 0x28,
		0b110: 0x30,
		0b111: 0x38,
	}

	tgt := (opcode & 0x38) >> 3
	c.sp--
	pc := c.pc
	c.memory.Write(c.sp, pc.High())
	c.sp--
	c.memory.Write(c.sp, pc.Low())

	c.previousPC = c.pc
	c.pc = NewWord(0x00, vec[tgt])
}
