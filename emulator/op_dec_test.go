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
}
