package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#CP_A,r8
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#CP_A,_HL_
func op_cp_a_r8(c *Cpu, opcode uint8) {

	// 0b00000111
	dst := opcode & 0x7

	var (
		nn int
	)

	// SUB a, [HL]
	if dst == reg_indirect_hl {

		// m-cycles = 2
		c.requiredCycles = 2

		// HL
		hl := c.reg.r16(reg_hl)

		// mem[HL]
		nn = int(c.memory.Read(hl))

	} else {

		// m-cycles = 1
		c.requiredCycles = 1

		// nn
		nn = int(c.reg.r8(dst))
	}

	// CP A
	cp_a(c, nn)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#CP_A,n8
func op_cp_a_imm8(c *Cpu, _ uint8) {

	// m-cycles = 2
	c.requiredCycles = 2

	// mem[PC]
	nn := int(c.fetch())

	// CP A
	cp_a(c, nn)
}

// cp_a CP A, R8|IMM8,[HL]
func cp_a(c *Cpu, nn int) {

	// read flags
	flags := c.reg.r_flags()

	// set n_flag=on
	flags |= n_flag

	// set z_flag=off
	// set c_flag=off
	// set h_flag=off
	flags &= ^(z_flag & c_flag & h_flag)

	// A
	a := int(c.reg.r8(reg_a))

	// borrow from bit 3
	if (a&0xF)-(nn&0xF) < 0 {
		// set half-carry=on
		flags |= h_flag
	}

	result := a - int(nn)

	// r8 > a
	if result < 0 {
		flags |= c_flag
		result = 256 + result
	}

	// zero
	if result == 0 {
		flags |= z_flag
	}

	// save flags
	c.reg.w_flag(flags)
}
