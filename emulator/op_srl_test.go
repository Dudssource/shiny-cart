package emulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpSrl(t *testing.T) {

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
		c.reg.w8(reg_a, 0x1)

		// SRL A
		op_srl_r8(c, 0b00111111)

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
					0xFF,
				},
			},
			reg: Registers{},
			sp:  0xFFF8,
		}

		c.reg.w8(reg_a, 0x1)
		c.reg.w16(reg_hl, 0x1)

		// SRL (HL)
		op_srl_r8(c, 0b00111110)

		assert.Equal(t, uint8(0x7F), c.memory.Read(c.reg.r16(reg_hl)))
		assert.True(t, c.reg.r_flags()&c_flag > 0)
		assert.True(t, c.reg.r_flags()&z_flag == 0)
		assert.True(t, c.reg.r_flags()&h_flag == 0)
		assert.True(t, c.reg.r_flags()&n_flag == 0)
	})
}
