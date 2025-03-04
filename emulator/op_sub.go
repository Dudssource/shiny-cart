package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#SUB_A,r8
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#SUB_A,_HL_
func op_sub_a_r8(c *Cpu, opcode uint8) {

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

	// SUB A
	sub_a(c, nn, 0)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#SUB_A,n8
func op_sub_a_imm8(c *Cpu, opcode uint8) {

	// m-cycles = 2
	c.requiredCycles = 2

	// mem[PC]
	nn := int(c.fetch())

	// SUB A
	sub_a(c, nn, 0)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#SBC_A,n8
func op_sbc_a_imm8(c *Cpu, _ uint8) {

	// m-cycles = 2
	c.requiredCycles = 2

	// carry
	carry := (c.reg.r_flags() & c_flag) >> 4

	// mem[PC]
	nn := int(c.fetch())

	// SBC A
	sub_a(c, nn, carry)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#SBC_A,r8
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#SBC_A,_HL_
func op_sbc_a_r8(c *Cpu, opcode uint8) {
	// 0b00000111
	dst := opcode & 0x7

	var (
		// nn = carry
		nn    int
		carry = (c.reg.r_flags() & c_flag) >> 4
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

	// SUB A
	sub_a(c, nn, carry)
}

// sub_a SUB A, R8|IMM8|[HL]
func sub_a(c *Cpu, nn int, carry flag) {

	// read flags
	flags := c.reg.r_flags()

	// set z_flag=off
	// set c_flag=off
	// set h_flag=off
	flags &= ^(z_flag | c_flag | h_flag)

	// set n_flag=on
	flags |= n_flag

	// A
	a := int(c.reg.r8(reg_a))

	// borrow from bit 3
	if (a&0xF - nn&0xF - int(carry)) < 0 {
		// set half-carry=on
		flags |= h_flag
	}

	result := a - int(nn) - int(carry)

	// r8 > a
	if result < 0 {
		flags |= c_flag
		result = 256 + result
	}

	// zero
	if result == 0 {
		flags |= z_flag
	}

	// reg[a] = r
	c.reg.w8(reg_a, uint8(result))

	// save flags
	c.reg.w_flag(flags)
}
