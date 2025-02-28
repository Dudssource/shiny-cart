package emulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInc(t *testing.T) {

	t.Run("test 1", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: []uint8{},
			},
			reg: Registers{},
			sp:  0x0,
		}

		c.reg.w8(reg_a, 0xFF)

		// INC A
		op_inc_r8(c, 0b00111100)
		assert.Equal(t, uint8(0x0), c.reg.r8(reg_a))
		assert.True(t, c.reg.r_flags()&z_flag > 0)
		assert.True(t, c.reg.r_flags()&h_flag > 0)
		assert.True(t, c.reg.r_flags()&n_flag == 0)
	})
}
