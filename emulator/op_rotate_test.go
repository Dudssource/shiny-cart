package emulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_op_rr_r8(t *testing.T) {

	t.Run("r8 carry 1", func(t *testing.T) {
		c := &Cpu{
			reg: Registers{
				reg_f: uint8(c_flag | h_flag),
				reg_a: uint8(0x1),
			},
		}
		// rr A
		op := uint8(0b00011111)
		op_rr_r8(c, op)

		assert.Equal(t, uint8(0b10000000), c.reg.r8(reg_a))
		assert.Equal(t, c_flag, c.reg.r_flags())
	})

	t.Run("rl carry wk", func(t *testing.T) {
		c := &Cpu{
			reg: Registers{
				reg_f: uint8(c_flag | h_flag),
				reg_a: uint8(0b00010111),
			},
		}
		// rl A
		op := uint8(0b00011111)
		op_rl_r8(c, op)

		assert.Equal(t, uint8(0b00101111), c.reg.r8(reg_a))
		assert.Equal(t, flag(0), c.reg.r_flags()&c_flag)
	})

	t.Run("r8 carry 0", func(t *testing.T) {
		c := &Cpu{
			reg: Registers{
				reg_f: uint8(0),
				reg_e: uint8(0xff),
			},
		}
		// rr E
		op := uint8(0b00011011)
		op_rr_r8(c, op)

		assert.Equal(t, uint8(0x7f), c.reg.r8(reg_e))
		assert.Equal(t, c_flag, c.reg.r_flags())
	})

	t.Run("z flag set", func(t *testing.T) {
		c := &Cpu{
			reg: Registers{
				reg_f: uint8(0),
				reg_e: uint8(0x1),
			},
		}
		// rr E
		op := uint8(0b00011011)
		op_rr_r8(c, op)

		assert.Equal(t, uint8(0x0), c.reg.r8(reg_e))
		assert.Equal(t, c_flag|z_flag, c.reg.r_flags())
	})

	t.Run("hl carry 1", func(t *testing.T) {
		c := &Cpu{
			reg: Registers{
				reg_f: uint8(c_flag),
			},
			memory: &Memory{
				mem: make(memoryArea, 1024),
			},
		}
		m := Word(0x40)
		c.reg.w16(reg_hl, m)
		c.memory.Write(m, 0x1)

		// rr [HL]
		op := uint8(0b00011110)
		op_rr_r8(c, op)

		assert.Equal(t, uint8(0b10000000), c.memory.Read(m))
		assert.Equal(t, c_flag, c.reg.r_flags())
	})

	t.Run("test 1", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x2,
					0xFF,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w8(reg_a, 0x1)

		// RR A
		op_rr_r8(c, 0b00011111)

		assert.Equal(t, uint8(0x0), c.reg.r8(reg_a))
		assert.True(t, c.reg.r_flags()&c_flag > 0)
		assert.True(t, c.reg.r_flags()&z_flag > 0)
		assert.True(t, c.reg.r_flags()&h_flag == 0)
		assert.True(t, c.reg.r_flags()&n_flag == 0)
	})

	t.Run("test 2", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x2,
					0x8A,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w8(reg_a, 0x1)
		c.reg.w16(reg_hl, 0x1)

		// RR (HL)
		op_rr_r8(c, 0b00011110)

		assert.Equal(t, uint8(0x45), c.memory.Read(c.reg.r16(reg_hl)))
		assert.True(t, c.reg.r_flags()&c_flag == 0)
		assert.True(t, c.reg.r_flags()&z_flag == 0)
		assert.True(t, c.reg.r_flags()&h_flag == 0)
		assert.True(t, c.reg.r_flags()&n_flag == 0)
	})

	t.Run("test 3", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x2,
					0x8A,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w8(reg_c, 0x1)

		// RRC C
		op_rrc_r8(c, 0b00001001)

		assert.Equal(t, uint8(0x80), c.reg.r8(reg_c))
		assert.True(t, c.reg.r_flags()&c_flag > 0)
		assert.True(t, c.reg.r_flags()&z_flag == 0)
		assert.True(t, c.reg.r_flags()&h_flag == 0)
		assert.True(t, c.reg.r_flags()&n_flag == 0)
	})

	t.Run("test 4", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x2,
					0x0,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w8(reg_c, 0x1)
		c.reg.w16(reg_hl, 0x1)

		// RRC (HL)
		op_rrc_r8(c, 0b00001110)

		assert.Equal(t, uint8(0x0), c.memory.Read(c.reg.r16(reg_hl)))
		assert.True(t, c.reg.r_flags()&c_flag == 0)
		assert.True(t, c.reg.r_flags()&z_flag > 0)
		assert.True(t, c.reg.r_flags()&h_flag == 0)
		assert.True(t, c.reg.r_flags()&n_flag == 0)
	})

	t.Run("test 5", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x2,
					0x0,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w8(reg_l, 0x80)

		// RL L
		op_rl_r8(c, 0b00010101)

		assert.Equal(t, uint8(0x0), c.reg.r8(reg_l))
		assert.True(t, c.reg.r_flags()&c_flag > 0)
		assert.True(t, c.reg.r_flags()&z_flag > 0)
		assert.True(t, c.reg.r_flags()&n_flag == 0)
	})

	t.Run("test 6", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x2,
					0x11,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w8(reg_l, 0x80)
		c.reg.w16(reg_hl, 0x1)

		// RL (HL)
		op_rl_r8(c, 0b00010110)

		assert.Equal(t, uint8(0x22), c.memory.Read(c.reg.r16(reg_hl)))
		assert.True(t, c.reg.r_flags()&c_flag == 0)
		assert.True(t, c.reg.r_flags()&z_flag == 0)
		assert.True(t, c.reg.r_flags()&n_flag == 0)
	})

	t.Run("test 7", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x2,
					0x11,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w8(reg_a, 0x81)

		// RRA
		op_rotate_rra(c, 0b00011111)

		assert.Equal(t, uint8(0x40), c.reg.r8(reg_a))
		assert.True(t, c.reg.r_flags()&c_flag > 0)
		assert.True(t, c.reg.r_flags()&z_flag == 0)
		assert.True(t, c.reg.r_flags()&n_flag == 0)
		assert.True(t, c.reg.r_flags()&h_flag == 0)
	})

	t.Run("test 8", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x2,
					0x11,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w8(reg_a, 0x3B)

		// RRCA
		op_rotate_rrca(c, 0b00001111)

		assert.Equal(t, uint8(0x9D), c.reg.r8(reg_a))
		assert.True(t, c.reg.r_flags()&c_flag > 0)
		assert.True(t, c.reg.r_flags()&z_flag == 0)
		assert.True(t, c.reg.r_flags()&n_flag == 0)
		assert.True(t, c.reg.r_flags()&h_flag == 0)
	})

	t.Run("test 9", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x2,
					0x11,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w8(reg_a, 0x85)

		// RLCA
		op_rotate_rlca(c, 0b00000111)

		assert.Equal(t, uint8(0x0B), c.reg.r8(reg_a))
		assert.True(t, c.reg.r_flags()&c_flag > 0)
		assert.True(t, c.reg.r_flags()&z_flag == 0)
		assert.True(t, c.reg.r_flags()&n_flag == 0)
		assert.True(t, c.reg.r_flags()&h_flag == 0)
	})

	t.Run("test 10", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x2,
					0x11,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w8(reg_a, 0x95)
		c.reg.w_flag(c_flag)

		_, n := c.decode(0b00010111)
		assert.Equal(t, "rla", n)

		// RLA
		op_rotate_rla(c, 0b00010111)

		assert.Equal(t, uint8(0x2B), c.reg.r8(reg_a))
		assert.True(t, c.reg.r_flags()&c_flag > 0)
		assert.True(t, c.reg.r_flags()&z_flag == 0)
		assert.True(t, c.reg.r_flags()&n_flag == 0)
		assert.True(t, c.reg.r_flags()&h_flag == 0)
	})
}
