package emulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpRlc(t *testing.T) {

	t.Run("test 1", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x0,
					0x0,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w8(reg_b, 0x85)

		_, n := c.decode(0xCB)
		assert.Equal(t, "CB rlc r8", n)

		// RLC B
		op_rlc_r8(c, 0b00000000)

		assert.Equal(t, uint8(0x0B), c.reg.r8(reg_b))
		assert.True(t, c.reg.r_flags()&c_flag > 0)
		assert.True(t, c.reg.r_flags()&z_flag == 0)
		assert.True(t, c.reg.r_flags()&h_flag == 0)
	})

	t.Run("test 2", func(t *testing.T) {

		c := &Cpu{
			pc: 0x2,
			memory: &Memory{
				mem: []uint8{
					0x2,
					0x0,
					0b00000110,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w8(reg_b, 0x85)
		c.reg.w16(reg_hl, 0x1)

		_, n := c.decode(0xCB)
		assert.Equal(t, "CB rlc r8", n)

		// RLC (HL)
		op_rlc_r8(c, 0b00000110)

		assert.Equal(t, uint8(0x0), c.memory.Read(c.reg.r16(reg_hl)))
		assert.True(t, c.reg.r_flags()&c_flag == 0)
		assert.True(t, c.reg.r_flags()&z_flag > 0)
		assert.True(t, c.reg.r_flags()&h_flag == 0)
	})
}
