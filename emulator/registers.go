package emulator

// https://gbdev.io/pandocs/CPU_Instruction_Set.html#cpu-instruction-set
const (
	reg_b           = 0x0
	reg_c           = 0x1
	reg_d           = 0x2
	reg_e           = 0x3
	reg_h           = 0x4
	reg_l           = 0x5
	reg_indirect_hl = 0x6
	reg_a           = 0x7
	reg_f           = 0x8
	reg_bc          = 0x0
	reg_de          = 0x1
	reg_hl          = 0x2
	reg_sp          = 0x3
)

type flag uint8

// https://gbdev.io/pandocs/CPU_Registers_and_Flags.html#the-flags-register-lower-8-bits-of-af-register
const (
	z_flag flag = 0b10000000 // Zero Flag
	n_flag flag = 0b01000000 // Subtraction flag (BCD)
	h_flag flag = 0b00100000 // Half Carry flag (BCD)
	c_flag flag = 0b00010000 // Carry flag
)

type Registers [64]uint8

func (c *Registers) w8(r uint8, v uint8) {
	c[r] = v
}

func (c *Registers) w16(r uint8, v Word) {
	// hi
	c[r*2] = uint8(v.High())

	// lo
	c[(r*2)+1] = uint8(v.Low())
}

func (c *Registers) r8(r uint8) uint8 {
	if r == reg_f {
		return c[r] & 0xF0
	}
	return c[r]
}

func (c *Registers) r16(r uint8) Word {
	return Word(uint16(c[r*2])<<8 | uint16(c[(r*2)+1]))
}

func (c *Registers) w_flag(v flag) {
	c.w8(reg_f, uint8(v))
}

func (c *Registers) r_flags() flag {
	return flag(c.r8(reg_f))
}
