package emulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_op_rr_r8(t *testing.T) {

	t.Run("r8 carry 1", func(t *testing.T) {
		c := &Cpu{
			reg: Registers{
				reg_f: uint8(c_flag | h_flag),
				reg_a: uint8(0x1),
			},
		}
		// rr A
		op := uint8(0b00011111)
		op_rr_r8(c, op)

		assert.Equal(t, uint8(0b10000000), c.reg.r8(reg_a))
		assert.Equal(t, c_flag, c.reg.r_flags())
	})

	t.Run("r8 carry 0", func(t *testing.T) {
		c := &Cpu{
			reg: Registers{
				reg_f: uint8(0),
				reg_e: uint8(0xff),
			},
		}
		// rr E
		op := uint8(0b00011011)
		op_rr_r8(c, op)

		assert.Equal(t, uint8(0x7f), c.reg.r8(reg_e))
		assert.Equal(t, c_flag, c.reg.r_flags())
	})

	t.Run("z flag set", func(t *testing.T) {
		c := &Cpu{
			reg: Registers{
				reg_f: uint8(0),
				reg_e: uint8(0x1),
			},
		}
		// rr E
		op := uint8(0b00011011)
		op_rr_r8(c, op)

		assert.Equal(t, uint8(0x0), c.reg.r8(reg_e))
		assert.Equal(t, c_flag|z_flag, c.reg.r_flags())
	})

	t.Run("hl carry 1", func(t *testing.T) {
		c := &Cpu{
			reg: Registers{
				reg_f: uint8(c_flag),
			},
			memory: &Memory{},
		}
		m := Word(0x40)
		c.reg.w16(reg_hl, m)
		c.memory.Write(m, 0x1)

		// rr [HL]
		op := uint8(0b00011110)
		op_rr_r8(c, op)

		assert.Equal(t, uint8(0b10000000), c.memory.Read(m))
		assert.Equal(t, c_flag, c.reg.r_flags())
	})
}
