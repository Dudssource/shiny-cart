package emulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpDaa(t *testing.T) {

	t.Run("half carry on", func(t *testing.T) {
		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: make(memoryArea, 0xFFFF),
			},
			reg: Registers{},
			sp:  0xFFFE,
		}
		c.reg.w8(reg_a, 0x3C)
		c.reg.w_flag(h_flag)
		op_daa(c, 0x0)
		assert.Equal(t, uint8(0x42), c.reg.r8(reg_a))
	})

	t.Run("op ld inc", func(t *testing.T) {
		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: make(memoryArea, 0xFFFF),
			},
			reg: Registers{},
			sp:  0xFFFE,
		}
		c.pc = 0x1
		c.memory.Write(0x1, 0x09)

		// LD a, $09
		op_ld_r8_imm8(c, 0b00111100)

		// INC r8
		op_inc_r8(c, 0b00111100)

		// DAA
		op_daa(c, 0x0)

		assert.Equal(t, uint8(0x10), c.reg.r8(reg_a))
	})
}
