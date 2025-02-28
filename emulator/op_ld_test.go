package emulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLd(t *testing.T) {

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

		op_ld_sp_e(c, 0)
		assert.Equal(t, Word(0xFFFA), c.reg.r16(reg_hl))
		assert.True(t, c.reg.r_flags()&c_flag == 0)
		assert.True(t, c.reg.r_flags()&h_flag == 0)
		assert.True(t, c.reg.r_flags()&n_flag == 0)
		assert.True(t, c.reg.r_flags()&z_flag == 0)
	})
}
