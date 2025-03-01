package emulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStk(t *testing.T) {

	t.Run("test 1", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: make(memoryArea, 0xFFFF),
			},
			reg: Registers{},
			sp:  0xFFFE,
		}

		c.reg.w16(reg_bc, 0xFFFC)

		// PUSH BC
		op_push_r16stk(c, 0b11000101)
		assert.Equal(t, Word(0xFFFC), c.sp)
		assert.Equal(t, uint8(0xFF), c.memory.Read(0xFFFD))
		assert.Equal(t, uint8(0xFC), c.memory.Read(0xFFFC))
	})

	t.Run("test 2", func(t *testing.T) {

		c := &Cpu{
			pc: 0x0,
			memory: &Memory{
				mem: make(memoryArea, 0xFFFF),
			},
			reg: Registers{},
			sp:  0xFFFE,
		}

		c.reg.w16(reg_bc, 0xFFFC)

		// PUSH BC
		op_push_r16stk(c, 0b11000101)
		// POP DE
		op_pop_r16stk(c, 0b11010001)
		assert.Equal(t, Word(0xFFFE), c.sp)
		assert.Equal(t, Word(0xFFFC), c.reg.r16(reg_de))
	})
}
