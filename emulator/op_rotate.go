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
