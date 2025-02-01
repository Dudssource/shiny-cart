package emulator

import "log"

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD_r8,r8
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD_r8,_HL_
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD__HL_,r8
func op_ld_r8_r8(c *Cpu, opcode uint8) {

	// 0b00111000
	dst := (opcode & 0x38) >> 3

	// 0b00000111
	src := opcode & 0x7

	// dst == [hl]
	if dst == reg_indirect_hl {

		// m-cycles=2
		c.requiredCycles = 2

		// LD [hl], r8
		c.memory.Write(c.reg.r16(reg_hl), c.reg.r8(src))

		if c.debug {
			log.Printf("LOAD [%.8X], %d\n", c.reg.r16(reg_hl), c.reg.r8(src))
		}

		return
	}

	// src == [hl]
	if src == reg_indirect_hl {

		// m-cycles=2
		c.requiredCycles = 2

		// LD dst, [hl]
		c.reg.w8(dst, c.memory.Read(c.reg.r16(reg_hl)))

		if c.debug {
			log.Printf("LOAD %d, [%.8X]\n", dst, c.reg.r16(reg_hl))
		}

	} else {

		// m-cycles=1
		c.requiredCycles = 1

		// LD dst, src
		c.reg.w8(dst, c.reg.r8(src))

		if c.debug {
			log.Printf("LOAD %d, %d\n", dst, src)
		}
	}
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD_r8,n8
func op_ld_r8_imm8(c *Cpu, opcode uint8) {

	// m-cycles=2
	c.requiredCycles = 2

	// 0b00111000
	dst := (opcode & 0x38) >> 3

	val := c.fetch()

	// dst == [hl]
	if dst == reg_indirect_hl {

		// m-cycles=3
		c.requiredCycles = 3

		// LD [hl], imm8
		c.memory.Write(c.reg.r16(reg_hl), val)

		if c.debug {
			log.Printf("LD [%.8X], %d\n", c.reg.r16(reg_hl), val)
		}

		return
	}

	if c.debug {
		log.Printf("LD %d, [%d]\n", dst, val)
	}

	// LD r8, imm8
	c.reg.w8(dst, val)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD_r16,n16
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD_SP,n16
func op_ld_r16_imm16(c *Cpu, opcode uint8) {

	// m-cycles = 3
	c.requiredCycles = 3

	// fetch little-endian
	lo := c.fetch()
	hi := c.fetch()

	// 0b00110000
	dst := (opcode & 0x30) >> 4

	if c.debug {
		log.Printf("LD %d, %.8X\n", dst, NewWord(hi, lo))
	}

	// LD r16, imm16
	c.reg.w16(dst, NewWord(hi, lo))
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD__r16_,A
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD__HLI_,A
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD__HLD_,A
func op_ld_r16mem_a(c *Cpu, opcode uint8) {

	// m-cycles = 2
	c.requiredCycles = 2

	// 0b00110000
	dst := (opcode & 0x30) >> 4

	// dst == [hl+]
	if dst == 0x2 {
		// hl = reg[hl]
		hl := c.reg.r16(reg_hl)

		// LD [hli], A
		c.memory.Write(hl, c.reg.r8(reg_a))

		// INC hl
		hl++

		// SAVE hl
		c.reg.w16(reg_hl, hl)

		return
	}

	// dst == [hl-]
	if dst == 0x3 {

		// hl = reg[hl]
		hl := c.reg.r16(reg_hl)

		// LD [hld], A
		c.memory.Write(hl, c.reg.r8(reg_a))

		// DEC hl
		hl--

		// SAVE hl
		c.reg.w16(reg_hl, hl)

		return
	}

	// LD [r16mem], A
	c.memory.Write(c.reg.r16(dst), c.reg.r8(reg_a))
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD_A,_r16_
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD_A,_HLD_
// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD_A,_HLI_
func op_ld_a_r16mem(c *Cpu, opcode uint8) {

	// m-cycles = 2
	c.requiredCycles = 2

	// 0b00110000
	src := (opcode & 0x30) >> 4

	// src == [hl+]
	if src == 0x2 {
		// hl = reg[hl]
		hl := c.reg.r16(reg_hl)

		// LD A, [hli]
		c.reg.w8(reg_a, c.memory.Read(hl))

		// INC hl
		hl++

		// SAVE hl
		c.reg.w16(reg_hl, hl)

		return
	}

	// src == [hl-]
	if src == 0x3 {

		// hl = reg[hl]
		hl := c.reg.r16(reg_hl)

		// LD A, [hld]
		c.reg.w8(reg_a, c.memory.Read(hl))

		// DEC hl
		hl--

		// SAVE hl
		c.reg.w16(reg_hl, hl)

		return
	}

	// LD A, [r16mem]
	c.reg.w8(reg_a, c.memory.Read(c.reg.r16(src)))
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD__n16_,SP
func op_ld_imm16_mem_sp(c *Cpu, opcode uint8) {

	// m-cycles = 5
	c.requiredCycles = 5

	// address little-endian
	lo := c.fetch()
	hi := c.fetch()

	// uint16
	nn := NewWord(hi, lo)

	// get SP
	sp := c.reg.r16(reg_sp)

	// write to memory the lsb of sp, little-endian
	c.memory.Write(nn, sp.Low())

	// inc address
	nn++

	// write to memory the msb of sp, little-endian
	c.memory.Write(nn, sp.High())
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD_A,_n16_
func op_ld_a_imm16(c *Cpu, _ uint8) {
	c.requiredCycles = 4
	z := c.fetch()
	w := c.fetch()
	wz := c.memory.Read(NewWord(w, z))
	c.reg.w8(reg_a, wz)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD__n16_,A
func op_ld_imm16_a(c *Cpu, _ uint8) {
	c.requiredCycles = 4
	nn_lsb := c.fetch()
	nn_msb := c.fetch()
	c.memory.Write(NewWord(nn_msb, nn_lsb), c.reg.r8(reg_a))
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD_HL,SP+e8
func op_ld_sp_e(c *Cpu, _ uint8) {
	c.requiredCycles = 3
	z := c.fetch()
	flags := c.reg.r_flags()
	flags &= ^z_flag & ^n_flag
	sp := c.sp.Low()
	result := uint16(sp + z)

	if (sp&0xF)+(z&0xF) > 0xF {
		flags |= h_flag
	}

	if result > 0xFF {
		flags |= c_flag
	}

	cf := uint8(flags >> 4)

	var adj = uint8(0)

	if z&uint8(z_flag) > 0 {
		adj = 0xFF
	}

	c.reg.w8(reg_h, c.sp.High()+adj+cf)
}

// https://rgbds.gbdev.io/docs/v0.7.0/gbz80.7#LD_SP,HL
func op_ld_sp_hl(c *Cpu, _ uint8) {
	c.requiredCycles = 2
	c.sp = c.reg.r16(reg_hl)
}
