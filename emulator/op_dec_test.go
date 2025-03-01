package emulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDec(t *testing.T) {

	t.Run("test 1", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{},
			},
			reg: Registers{},
			sp:  0x0,
		}

		c.reg.w8(reg_l, 0x01)

		// DEC L
		op_dec_r8(c, 0b00101101)
		assert.Equal(t, uint8(0x0), c.reg.r8(reg_l))
		assert.True(t, c.reg.r_flags()&z_flag > 0)
		assert.True(t, c.reg.r_flags()&h_flag == 0)
		assert.True(t, c.reg.r_flags()&n_flag > 0)
	})

	t.Run("test 2", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{
					0x0,
					0x0,
				},
			},
			reg: Registers{},
			sp:  0x0,
		}

		c.reg.w16(reg_hl, 0x1)

		// DEC L
		op_dec_r8(c, 0b00110101)
		assert.Equal(t, uint8(0xFF), c.memory.Read(c.reg.r16(reg_hl)))
		assert.True(t, c.reg.r_flags()&z_flag == 0)
		assert.True(t, c.reg.r_flags()&h_flag > 0)
		assert.True(t, c.reg.r_flags()&n_flag > 0)
	})
}
