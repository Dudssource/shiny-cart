package emulator

import "log"

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#JR_n16
func op_jr_imm8(c *Cpu, _ uint8) {

	// m-cycles = 3
	c.requiredCycles = 3

	// retrieve N from memory and increase PC
	offset := c.fetch()

	// check sign bit
	signed := ((offset & 0x80) >> 7) > 0

	if signed {
		offset = ^offset + 1
		c.pc -= Word(offset)
	} else {
		c.pc += Word(offset)
	}

	if c.debug {
		log.Printf("JR %.8X\n", c.pc)
	}
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#JR_cc,n16
func op_jr_cond_imm8(c *Cpu, opcode uint8) {

	// my-cycles = 2
	c.requiredCycles = 2

	match := eval(c.reg.r_flags(), opcode)

	// retrieve N from memory and increase PC
	offset := c.fetch()

	if match {

		// my-cycles = 3
		c.requiredCycles = 3

		// check sign bit
		signed := ((offset & 0x80) >> 7) > 0

		if signed {
			offset = ^offset + 1
			c.pc -= Word(offset)
		} else {
			c.pc += Word(offset)
		}
	}
}
