package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#AND_A,r8
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#AND_A,_HL_
func op_and_a_r8(c *Cpu, opcode uint8) {

	// dst
	dst := opcode & 0x7

	var (
		nn uint8
	)

	if dst == reg_indirect_hl {

		// m-cycles = 2
		c.requiredCycles = 2

		// HL
		hl := c.reg.r16(reg_hl)

		// mem[HL]
		nn = c.memory.Read(hl)

	} else {

		// m-cycles = 1
		c.requiredCycles = 1

		// nn = R8
		nn = c.reg.r8(dst)
	}

	// AND A
	and_a(c, nn)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#AND_A,n8
func op_and_a_imm8(c *Cpu, _ uint8) {

	// m-cycles = 2
	c.requiredCycles = 2

	// mem[PC]
	nn := c.fetch()

	// AND A
	and_a(c, nn)
}

// and_a AND A, R8|IMM8|[HL]
func and_a(c *Cpu, nn uint8) {

	// read flags
	flags := c.reg.r_flags()

	// z_flag=off
	// n_flag=off
	// c_flag=off
	flags &= ^z_flag & ^n_flag & ^c_flag

	// h_flag=on
	flags |= h_flag

	// reg[a]
	a := c.reg.r8(reg_a)

	// a = A AND NN
	rr := a & nn

	// zero-flag=on
	if rr == 0 {
		flags |= z_flag
	}

	// a = A AND NN
	c.reg.w8(reg_a, rr)

	// save flags
	c.reg.w_flag(flags)
}
