package emulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCp(t *testing.T) {

	t.Run("test 1", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{},
			},
			reg: Registers{},
			sp:  0x0,
		}

		c.reg.w8(reg_a, 0x3C)
		c.reg.w8(reg_b, 0x2F)

		// CP A, B
		op_cp_a_r8(c, 0b10111000)
		assert.True(t, c.reg.r_flags()&z_flag == 0)
		assert.True(t, c.reg.r_flags()&h_flag > 0)
		assert.True(t, c.reg.r_flags()&n_flag > 0)
		assert.True(t, c.reg.r_flags()&c_flag == 0)
	})

	t.Run("test 2", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x3C,
				},
			},
			reg: Registers{},
			sp:  0x0,
		}

		c.reg.w8(reg_a, 0x3C)

		// CP A, 0x3C
		op_cp_a_r8(c, 0b11111110)
		assert.True(t, c.reg.r_flags()&z_flag > 0)
		assert.True(t, c.reg.r_flags()&h_flag == 0)
		assert.True(t, c.reg.r_flags()&n_flag > 0)
		assert.True(t, c.reg.r_flags()&c_flag == 0)
	})

	t.Run("test 3", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x3C,
					0x40,
				},
			},
			reg: Registers{},
			sp:  0x0,
		}

		c.reg.w8(reg_a, 0x3C)
		c.reg.w16(reg_hl, 0x1)

		// CP A, (HL)
		op_cp_a_r8(c, 0b10111110)
		assert.True(t, c.reg.r_flags()&z_flag == 0)
		assert.True(t, c.reg.r_flags()&h_flag == 0)
		assert.True(t, c.reg.r_flags()&n_flag > 0)
		assert.True(t, c.reg.r_flags()&c_flag > 0)
	})
}
