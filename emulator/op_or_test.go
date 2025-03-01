package emulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpOr(t *testing.T) {

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
		c.reg.w8(reg_a, 0x5A)

		op_or_a_r8(c, 0b10110111)

		assert.Equal(t, uint8(0x5A), c.reg.r8(reg_a))
		assert.True(t, c.reg.r_flags()&z_flag == 0)
	})

	t.Run("test 2", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x3,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}
		c.reg.w8(reg_a, 0x5A)

		// OR 3
		op_or_a_r8(c, 0b11110110)

		assert.Equal(t, uint8(0x5B), c.reg.r8(reg_a))
		assert.True(t, c.reg.r_flags()&z_flag == 0)
	})

	t.Run("test 3", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x2,
					0x0F,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}
		c.reg.w8(reg_a, 0x5A)
		c.reg.w16(reg_hl, 0x1)

		op_or_a_r8(c, 0b10110110)

		assert.Equal(t, uint8(0x5F), c.reg.r8(reg_a))
		assert.True(t, c.reg.r_flags()&z_flag == 0)
	})
}
