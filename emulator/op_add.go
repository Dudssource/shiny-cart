package emulator

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#ADD_HL,r16
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#ADD_HL,SP
func op_add_hl_r16(c *Cpu, opcode uint8) {

	// m-cycles = 2
	c.requiredCycles = 2

	// read flags
	flags := c.reg.r_flag()

	// set n_flag=off
	flags &= ^n_flag

	// 0b00110000
	dst := (opcode & 0x30) >> 4

	// nn = r16
	nn := c.reg.r16(dst)

	// HL
	hl := c.reg.r16(reg_h)

	// L + lsb(NN)
	result_l := hl.Low() + nn.Low()

	// if result is greater than 0xFF, set carry=on
	if result_l > 0xFF {
		flags &= c_flag
	}

	// if result is greater than 0xF, set carry n=on
	// TODO: Review
	if (result_l&0xF)>>3 == 1 {
		flags &= n_flag
	}

	// set L
	c.reg.w8(reg_l, result_l)

	// H + msb(NN) + flags.C
	result_h := hl.High() + nn.High() + uint8((flags&c_flag)>>4)

	// if result is greater than 0xFF, set carry=on
	if (result_h & 0xFF) > 0xF {
		flags &= c_flag
	}

	// if result is greater than 0xF, set carry n=on
	// TODO: Review
	if (result_h&0xF)>>3 == 1 {
		flags &= n_flag
	}

	// set H
	c.reg.w8(reg_h, result_h)

	// save flags
	c.reg.w_flag(flags)
}
