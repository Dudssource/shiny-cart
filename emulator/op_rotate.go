package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#RLCA
func op_rotate_rlca(c *Cpu, opcode uint8) {

	// m-cycles = 1
	c.requiredCycles = 1

	// flags
	flags := c.reg.r_flags()

	// carry=off
	// zero=off
	// half-carry=off
	// subtraction=off
	flags &= ^c_flag & ^z_flag & ^n_flag & ^h_flag

	// accumulator
	a := c.reg.r8(reg_a)

	// 0b10000000
	// set carry=on
	if (a & 0x80) > 0 {
		flags |= c_flag
	}

	// RL
	a <<= 1

	// A = A << 1 (RL)
	c.reg.w8(reg_a, a)

	// save flags
	c.reg.w_flag(flags)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#RRCA
func op_rotate_rrca(c *Cpu, opcode uint8) {

	// m-cycles = 1
	c.requiredCycles = 1

	// flags
	flags := c.reg.r_flags()

	// carry=off
	// zero=off
	// half-carry=off
	// subtraction=off
	flags &= ^c_flag & ^z_flag & ^n_flag & ^h_flag

	// accumulator
	a := c.reg.r8(reg_a)

	// 0b00000001
	// set carry=on
	if (a & 0x1) == 1 {
		flags |= c_flag
	}

	// RR
	a >>= 1

	// A = A >> 1 (SR)
	c.reg.w8(reg_a, a)

	// save flags
	c.reg.w_flag(flags)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#RLA
func op_rotate_rla(c *Cpu, opcode uint8) {

	// m-cycles = 1
	c.requiredCycles = 1

	// flags
	flags := c.reg.r_flags()

	// get carry-flag
	carry := (flags & c_flag) >> 4

	// carry=off
	// zero=off
	// half-carry=off
	// subtraction=off
	flags &= ^c_flag & ^z_flag & ^n_flag & ^h_flag

	// accumulator
	a := c.reg.r8(reg_a)

	// 0b10000000
	// set carry=on
	if (a & 0x80) > 0 {
		flags |= c_flag
	}

	// RL A
	a <<= 1

	// turn the rightmost bit on
	if carry == 1 {
		a |= 1
	}

	// A = A << 1 (RL)
	c.reg.w8(reg_a, a)

	// save flags
	c.reg.w_flag(flags)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#RRA
func op_rotate_rra(c *Cpu, opcode uint8) {

	// m-cycles = 1
	c.requiredCycles = 1

	// flags
	flags := c.reg.r_flags()

	// get carry-flag
	carry := (flags & c_flag) >> 4

	// carry=off
	// zero=off
	// half-carry=off
	// subtraction=off
	flags &= ^c_flag & ^z_flag & ^n_flag & ^h_flag

	// accumulator
	a := c.reg.r8(reg_a)

	// 0b10000000
	// set carry=on
	if (a & 0x1) == 1 {
		flags |= c_flag
	}

	// RR A
	a >>= 1

	// turn the leftmost bit on
	if carry == 1 {
		a |= 0x80
	}

	// A = A >> 1 (RR)
	c.reg.w8(reg_a, a)

	// save flags
	c.reg.w_flag(flags)
}

// https://rgbds.gbdev.io/docs/v0.8.0/gbz80.7#RR_r8
// https://rgbds.gbdev.io/docs/v0.8.0/gbz80.7#RR__HL_
func op_rr_r8(c *Cpu, opcode uint8) {

	c.requiredCycles = 2

	flags := c.reg.r_flags()

	// reset z, n, h
	flags &= ^(h_flag | n_flag | z_flag)

	// save current carry flag
	old_carry := uint8((flags & c_flag) >> 4)

	// get operand
	var nn uint8
	r8 := (opcode & 0x7)

	// check indirect hl or not
	if r8 == reg_indirect_hl {

		c.requiredCycles = 4

		hl := c.reg.r16(reg_hl)
		nn = c.memory.Read(hl)
	} else {
		nn = c.reg.r8(r8)
	}

	// get the rightmost bit as the new value for carry
	new_carry := nn & 0x1

	// shift right by 1
	nn = (nn & 0xFE) >> 1

	// set the leftmost bit as the old carry
	nn |= (old_carry << 7)

	if nn == 0 {
		flags |= z_flag
	}

	// write
	if r8 == reg_indirect_hl {
		hl := c.reg.r16(reg_hl)
		c.memory.Write(hl, nn)
	} else {
		c.reg.w8(r8, nn)
	}

	// set the new carry flag
	flags |= flag(new_carry << 4)

	// save the flags
	c.reg.w_flag(flags)
}

// https://rgbds.gbdev.io/docs/v0.9.0/gbz80.7#RL_r8
// https://rgbds.gbdev.io/docs/v0.9.0/gbz80.7#RL__HL_
func op_rl_r8(c *Cpu, opcode uint8) {

	c.requiredCycles = 2

	flags := c.reg.r_flags()

	// reset z, n, h
	flags &= ^(h_flag | n_flag | z_flag)

	// save current carry flag
	old_carry := uint8((flags & c_flag) >> 4)

	// get operand
	var nn uint8
	r8 := (opcode & 0x7)

	// check indirect hl or not
	if r8 == reg_indirect_hl {

		c.requiredCycles = 4

		hl := c.reg.r16(reg_hl)
		nn = c.memory.Read(hl)
	} else {
		nn = c.reg.r8(r8)
	}

	// get the leftmost bit as the new value for carry
	new_carry := nn & 0x80 >> 7

	// shift left by 1
	nn = nn & 0x7F << 1

	// set the rightmost bit as the old carry
	nn |= old_carry & 0x1

	if nn == 0 {
		flags |= z_flag
	}

	// write
	if r8 == reg_indirect_hl {
		hl := c.reg.r16(reg_hl)
		c.memory.Write(hl, nn)
	} else {
		c.reg.w8(r8, nn)
	}

	// set the new carry flag
	flags |= flag(new_carry << 4)

	// save the flags
	c.reg.w_flag(flags)
}
