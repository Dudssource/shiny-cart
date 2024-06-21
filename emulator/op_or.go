package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#OR_A,r8
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#OR_A,_HL_
func op_or_a_r8(c *Cpu, opcode uint8) {

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

	// OR A
	or_a(c, nn)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#OR_A,n8
func op_or_a_imm8(c *Cpu, _ uint8) {

	// m-cycles = 2
	c.requiredCycles = 2

	// mem[PC]
	nn := c.fetch()

	// OR A
	or_a(c, nn)
}

// or_a OR A, R8|IMM8|[HL]
func or_a(c *Cpu, nn uint8) {

	// read flags
	flags := c.reg.r_flags()

	// z_flag=off
	// n_flag=off
	// c_flag=off
	// h_flag=off
	flags &= ^z_flag & ^n_flag & ^c_flag &^ h_flag

	// reg[a]
	a := c.reg.r8(reg_a)

	// a = A OR NN
	rr := a | nn

	// zero-flag=on
	if rr == 0 {
		flags |= z_flag
	}

	// a = A OR NN
	c.reg.w8(reg_a, rr)

	// save flags
	c.reg.w_flag(flags)
}
