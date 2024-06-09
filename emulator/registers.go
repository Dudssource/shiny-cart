package emulator

const (
	reg_b  = 0x0
	reg_c  = 0x1
	reg_d  = 0x2
	reg_e  = 0x3
	reg_h  = 0x4
	reg_l  = 0x5
	reg_a  = 0x7
	reg_bc = 0x0
	reg_de = 0x1
	reg_hl = 0x2
	reg_sp = 0x3
)

type Registers [64]uint8

func (c *Registers) w8(r uint8, v uint8) {
	c[r] = v
}

func (c *Registers) w16(r uint8, v uint16) {
	c[r] = uint8((v & 0xFF00) >> 8)
	c[r+1] = uint8((v & 0xFF))
}

func (c *Registers) r8(r uint8) uint8 {
	return c[r]
}

func (c *Registers) r16(r uint8) uint16 {
	return uint16(c[r])<<8 | uint16(c[r+1])
}
