package emulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpAdd(t *testing.T) {

	t.Run("test 1", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x2,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		op_add_sp_imm8(c, 0)
		assert.Equal(t, Word(0xFFFA), c.sp)
	})

	t.Run("test 2", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0xFE,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		op_add_sp_imm8(c, 0)
		assert.Equal(t, Word(0xFFF6), c.sp)
	})

	t.Run("test 3", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0xFE,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w16(reg_hl, 0x8A23)
		c.reg.w16(reg_bc, 0x0605)

		// ADD HL, BC
		op_add_hl_r16(c, 0b00001001)
		assert.Equal(t, Word(0x9028), c.reg.r16(reg_hl))
		assert.True(t, c.reg.r_flags()&h_flag > 0)
		assert.True(t, c.reg.r_flags()&n_flag == 0)
		assert.True(t, c.reg.r_flags()&c_flag == 0)
	})

	t.Run("test 4", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0xFE,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w16(reg_hl, 0x8A23)
		c.reg.w16(reg_bc, 0x0605)

		// ADD HL, BC
		op_add_hl_r16(c, 0b00101001)
		assert.Equal(t, Word(0x1446), c.reg.r16(reg_hl))
		assert.True(t, c.reg.r_flags()&h_flag > 0)
		assert.True(t, c.reg.r_flags()&n_flag == 0)
		assert.True(t, c.reg.r_flags()&c_flag > 0)
	})

	t.Run("test 5", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0xFE,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w8(reg_a, 0x3A)
		c.reg.w8(reg_b, 0xC6)

		// ADD A, B
		op_add_a_r8(c, 0b10000000)
		assert.Equal(t, uint8(0x0), c.reg.r8(reg_a))
		assert.True(t, c.reg.r_flags()&h_flag > 0)
		assert.True(t, c.reg.r_flags()&n_flag == 0)
		assert.True(t, c.reg.r_flags()&c_flag > 0)
		assert.True(t, c.reg.r_flags()&z_flag > 0)
	})
}
