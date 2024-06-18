package emulator

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
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#JR_cc,n16
func op_jr_cond_imm8(c *Cpu, opcode uint8) {

	// my-cycles = 2
	c.requiredCycles = 2

	const (
		cond_nz = 0x0
		cond_z  = 0x1
		cond_nc = 0x2
		cond_c  = 0x3
	)

	// read flags
	flags := c.reg.r_flags()

	// 0b00011000
	cond := (opcode & 0x18) >> 3

	var match bool

	switch cond {
	case cond_nz:
		match = (flags & z_flag) == 0
	case cond_z:
		match = (flags & z_flag) != 0
	case cond_nc:
		match = (flags & c_flag) == 0
	case cond_c:
		match = (flags & c_flag) != 0
	}

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
