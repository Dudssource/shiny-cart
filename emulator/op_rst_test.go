package emulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpRst(t *testing.T) {

	t.Run("test 1", func(t *testing.T) {

		c := &Cpu{
			pc: 0x8001,
			memory: &Memory{
				mem: []uint8{0x0, 0x0},
			},
			reg: Registers{},
			sp:  0x2,
		}

		op_rst(c, 0b11001111)
		assert.Equal(t, Word(0x0008), c.pc)
	})
}
