package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#ADD_HL,r16
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#ADD_HL,SP
func op_add_hl_r16(c *Cpu, opcode uint8) {

	// m-cycles = 2
	c.requiredCycles = 2

	// read flags
	flags := c.reg.r_flags()

	// no subtraction
	// disable carry flag as default
	// disable half-carry flag as default
	flags &= ^n_flag & ^c_flag & ^h_flag

	// 0b00110000
	dst := (opcode & 0x30) >> 4

	// nn = r16
	nn := c.reg.r16(dst)

	// HL
	hl := c.reg.r16(reg_hl)

	// HL + NN
	result := uint32(hl + nn)

	// if result is greater than 0xFFF, set half-carry=on (bit 11)
	if (hl&0xFFF + nn&0xFFF) > 0xFFF {
		flags |= h_flag
	}

	// if result is greater than 0xFFFF, set carry=on (bit 15)
	if result > 0xFFFF {
		flags |= c_flag
		result = result - 0xFFFF
	}

	// set HL
	c.reg.w16(reg_h, Word(result))

	// save flags
	c.reg.w_flag(flags)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#ADD_A,r8
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#ADD_A,_HL_
func op_add_a_r8(c *Cpu, opcode uint8) {

	// 0b00000111
	dst := (opcode & 0x7)

	var (
		nn uint8
	)

	// ADD A, [HL]
	if dst == reg_indirect_hl {

		// m-cycles = 2
		c.requiredCycles = 2

		// HL
		hl := c.reg.r16(reg_hl)

		// mem[HL]
		nn = c.memory.Read(hl)

	} else {

		// nn = r8
		nn = c.reg.r8(dst)
	}

	// A
	add_a(c, nn)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#ADC_A,r8
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#ADC_A,_HL_
func op_adc_a_r8(c *Cpu, opcode uint8) {

	// 0b00000111
	dst := (opcode & 0x7)

	var (
		nn uint8 = uint8((c.reg.r_flags() & c_flag) >> 4)
	)

	// ADC A, [HL]
	if dst == reg_indirect_hl {

		// m-cycles = 2
		c.requiredCycles = 2

		// HL
		hl := c.reg.r16(reg_hl)

		// mem[HL]
		nn += c.memory.Read(hl)

	} else {

		// m-cycles = 1
		c.requiredCycles = 1

		// nn = r8
		nn += c.reg.r8(dst)
	}

	// A
	add_a(c, nn)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#ADD_A,n8
func op_add_a_imm8(c *Cpu, _ uint8) {

	// m-cycles = 2
	c.requiredCycles = 2

	// mem[PC]
	nn := c.fetch()

	// A
	add_a(c, nn)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#ADC_A,n8
func op_adc_a_imm8(c *Cpu, _ uint8) {

	// m-cycles = 2
	c.requiredCycles = 2

	// carry
	carry := uint8((c.reg.r_flags() & c_flag) >> 4)

	// mem[PC]
	nn := c.fetch()

	// A
	add_a(c, nn+carry)
}

// add_a ADD A, R8|IMM8|[HL]
func add_a(c *Cpu, nn uint8) {

	// read flags
	flags := c.reg.r_flags()

	// no subtraction
	// disable carry flag as default
	// disable half-carry flag as default
	// disable zero flag as default
	flags &= ^n_flag & ^c_flag & ^h_flag & ^z_flag

	// A
	a := c.reg.r8(reg_a)

	// if result is greater than 0xF, set half-carry=on (bit 3)
	if (a&0xF + nn&0xF) > 0xF {
		flags |= h_flag
	}

	// A + NN
	result := uint16(a + nn)

	// if result is greater than 0xFF, set carry=on (bit 7)
	if result > 0xFF {
		flags |= c_flag
		result = result - 0xFF
	}

	// z-flag = on
	if result == 0 {
		flags |= z_flag
	}

	// set A
	c.reg.w8(reg_a, uint8(result))

	// save flags
	c.reg.w_flag(flags)
}
