package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD_r8,r8
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD_r8,_HL_
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD__HL_,r8
func ld_r8_r8(c *Cpu, opcode uint8) {

	// 0b00111000
	dst := (opcode & 0x38) >> 3
	// 0b00000111
	src := opcode & 0x7

	// dst == [hl]
	if dst == reg_indirect_hl {

		// LD [hl], r8
		c.memory.Write(c.reg.r16(reg_hl), c.reg.r8(src))
	}

	// src == [hl]
	if src == reg_indirect_hl {

		// LD dst, [hl]
		c.reg.w8(dst, c.memory.Read(c.reg.r16(reg_hl)))

	} else {

		// LD dst, src
		c.reg.w8(dst, c.reg.r8(src))
	}
}
