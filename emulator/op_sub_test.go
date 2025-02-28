package emulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSub(t *testing.T) {

	t.Run("test 1", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{},
			},
			reg: Registers{},
			sp:  0x0,
		}

		c.reg.w8(reg_a, 0x3E)
		c.reg.w8(reg_e, 0x3E)

		// SUB A, E
		op_sub_a_r8(c, 0b10010011)
		assert.Equal(t, uint8(0x0), c.reg.r8(reg_a))
		assert.True(t, c.reg.r_flags()&z_flag > 0)
		assert.True(t, c.reg.r_flags()&h_flag == 0)
		assert.True(t, c.reg.r_flags()&n_flag > 0)
		assert.True(t, c.reg.r_flags()&c_flag == 0)
	})

	t.Run("test 2", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x0F,
				},
			},
			reg: Registers{},
			sp:  0x0,
		}

		c.reg.w8(reg_a, 0x3E)

		// SUB A, 0x0F
		op_sub_a_imm8(c, 0b11010110)
		assert.Equal(t, uint8(0x2F), c.reg.r8(reg_a))
		assert.True(t, c.reg.r_flags()&z_flag == 0)
		assert.True(t, c.reg.r_flags()&h_flag > 0)
		assert.True(t, c.reg.r_flags()&n_flag > 0)
		assert.True(t, c.reg.r_flags()&c_flag == 0)
	})

	t.Run("test 3", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x0F,
					0x40,
				},
			},
			reg: Registers{},
			sp:  0x0,
		}

		c.reg.w8(reg_a, 0x3E)
		c.reg.w16(reg_hl, 0x1)

		// SUB A, (HL)
		op_sub_a_r8(c, 0b10010110)
		assert.Equal(t, uint8(0xFE), c.reg.r8(reg_a))
		assert.True(t, c.reg.r_flags()&z_flag == 0)
		assert.True(t, c.reg.r_flags()&h_flag == 0)
		assert.True(t, c.reg.r_flags()&n_flag > 0)
		assert.True(t, c.reg.r_flags()&c_flag > 0)
	})
}
